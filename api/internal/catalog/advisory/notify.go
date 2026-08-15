package advisory

import (
	"context"
	"encoding/json"
	"log/slog"
	"strconv"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
)

// maxNotifiedReasons caps how many advisory rows are named in one alert. The
// rest are covered by the count, so a large burst stays readable.
const maxNotifiedReasons = 8

// notifyNewAdvisories alerts an environment's subscribers about advisories that
// have just started matching it.
//
// One event per environment per sync, never one per advisory: matches arrive in
// bursts (Shopware publishes several advisories together), so per-advisory
// alerts would deliver a handful of near-identical emails in the same minute.
//
// Suppressed advisories are filtered out before dispatch. An acknowledged
// advisory that still emails would defeat the point of acknowledging it.
//
// Every early return here leaves the advisory unmarked, so the next pass
// retries it: the caller has already committed the match rows, and a transient
// failure below must not cost the alert permanently.
func (s *Service) notifyNewAdvisories(ctx context.Context, environmentID int32, matches []sbom.Match, known map[string]bool) {
	if s.notifier == nil || len(matches) == 0 {
		return
	}

	suppressed, err := s.suppressedAdvisoryIDs(ctx, environmentID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load suppressions before advisory alert",
			"environmentId", environmentID, "error", err)
		// Err toward silence: alerting on an advisory the owner may already have
		// acknowledged is worse than a delayed alert, which the next sync fixes.
		return
	}

	seen := make(map[string]bool, len(matches))
	fresh := make([]string, 0, len(matches))
	for _, match := range matches {
		if known[match.AdvisoryID] || suppressed[match.AdvisoryID] || seen[match.AdvisoryID] {
			continue
		}
		seen[match.AdvisoryID] = true
		fresh = append(fresh, match.AdvisoryID)
	}
	if len(fresh) == 0 {
		return
	}

	env, err := s.queries.GetEnvironmentForScrape(ctx, environmentID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load environment for advisory alert",
			"environmentId", environmentID, "error", err)
		return
	}

	subscribers, err := s.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: env.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(environmentID)),
	})
	if err != nil {
		slog.WarnContext(ctx, "failed to load advisory alert subscribers",
			"environmentId", environmentID, "error", err)
		return
	}
	if len(subscribers) == 0 {
		// Nobody to tell. Mark them anyway: there is nothing to retry, and
		// leaving them unmarked would re-evaluate the same advisories on every
		// pass forever.
		s.markNotified(ctx, environmentID, fresh)
		return
	}

	recipients := make([]notify.Recipient, 0, len(subscribers))
	for _, u := range subscribers {
		recipients = append(recipients, notify.Recipient{
			ID:     u.ID,
			Name:   u.Name,
			Email:  u.Email,
			Locale: u.Locale,
		})
	}

	reasons := s.advisoryReasons(ctx, matches, fresh)

	name := env.Name
	if env.ShopName != nil && *env.ShopName != "" {
		name = *env.ShopName + " · " + env.Name
	}

	titleKey := "notification.advisoryDetected.title"
	messageKey := "notification.advisoryDetected.message"
	subjectKey := "email.advisoryDetected.subject"
	if len(fresh) == 1 {
		titleKey = "notification.advisoryDetected.titleOne"
		messageKey = "notification.advisoryDetected.messageOne"
		subjectKey = "email.advisoryDetected.subjectOne"
	}

	res := s.notifier.Dispatch(ctx, notify.Event{
		Type:      notify.EventAdvisoryDetected,
		Level:     notify.LevelWarning,
		ScopeType: notify.ScopeEnvironment,
		ScopeID:   strconv.Itoa(int(environmentID)),
		OrgID:     env.OrganizationID,
		// Keyed per environment rather than per advisory, so the in-app upsert
		// collapses a burst of new advisories into one entry.
		DedupKey:        "advisory_detected:" + strconv.Itoa(int(environmentID)),
		TitleKey:        titleKey,
		MessageKey:      messageKey,
		EmailSubjectKey: subjectKey,
		Params: map[string]any{
			"name":  name,
			"count": len(fresh),
		},
		Reasons: reasons,
		Link:    advisoryAlertLink(environmentID, fresh),
	}, recipients)

	// Only record the advisories as notified when nothing failed. A channel
	// error here means these subscribers were not told, and marking anyway
	// would drop them from every future rematch — the alert would never be
	// retried. A dispatch that delivered nothing but failed nothing (every
	// channel skipped by the email dedup lock or by recipient preference) has
	// nothing to retry, so it counts as handled.
	if res.Failed > 0 {
		slog.WarnContext(ctx, "advisory alert partially failed; leaving advisories unnotified for retry",
			"environmentId", environmentID, "delivered", res.Delivered,
			"failed", res.Failed, "advisories", len(fresh))
		return
	}

	s.markNotified(ctx, environmentID, fresh)
}

// advisoryAlertLink sends a single-advisory alert to that advisory's page, and
// a burst to the environment so the owner can see every hit in one place.
func advisoryAlertLink(environmentID int32, fresh []string) notify.Link {
	if len(fresh) == 1 {
		return notify.Link{
			Name:   "account.advisories.detail",
			Params: map[string]string{"id": fresh[0]},
		}
	}
	return notify.Link{
		Name:   "account.environments.detail",
		Params: map[string]string{"environmentId": strconv.Itoa(int(environmentID))},
	}
}

// markNotified records that subscribers have been told about these advisories.
// Written only after a dispatch that reported no channel failures, so an alert
// lost to a transient failure is retried on the next pass rather than being
// silently swallowed.
func (s *Service) markNotified(ctx context.Context, environmentID int32, advisoryIDs []string) {
	if len(advisoryIDs) == 0 {
		return
	}
	if err := s.queries.MarkEnvironmentAdvisoriesNotified(ctx, queries.MarkEnvironmentAdvisoriesNotifiedParams{
		EnvironmentID: environmentID,
		Column2:       advisoryIDs,
	}); err != nil {
		// The alert went out; failing to record it means the next pass repeats
		// it. Noisy, but strictly better than losing it.
		slog.WarnContext(ctx, "failed to record advisory notification",
			"environmentId", environmentID, "error", err)
	}
}

func (s *Service) suppressedAdvisoryIDs(ctx context.Context, environmentID int32) (map[string]bool, error) {
	rows, err := s.queries.ListSuppressedAdvisoriesForEnvironment(ctx, environmentID)
	if err != nil {
		return nil, err
	}
	out := make(map[string]bool, len(rows))
	for _, row := range rows {
		out[row.AdvisoryID] = true
	}
	return out, nil
}

// advisoryReasons builds the per-advisory lines that the email and in-app
// notification list under the summary. Each line names the identifier,
// severity, affected package, installed version, and a fix when we know one.
func (s *Service) advisoryReasons(ctx context.Context, matches []sbom.Match, freshIDs []string) []notify.StatusReason {
	byAdvisory := make(map[string][]sbom.Match, len(freshIDs))
	for _, match := range matches {
		byAdvisory[match.AdvisoryID] = append(byAdvisory[match.AdvisoryID], match)
	}

	reasons := make([]notify.StatusReason, 0, min(len(freshIDs), maxNotifiedReasons))
	for _, id := range freshIDs {
		if len(reasons) >= maxNotifiedReasons {
			break
		}
		row, err := s.queries.GetComposerAdvisory(ctx, id)
		if err != nil {
			slog.WarnContext(ctx, "failed to load advisory for alert", "advisoryId", id, "error", err)
			continue
		}
		hits := byAdvisory[id]
		if len(hits) == 0 {
			hits = []sbom.Match{{AdvisoryID: id}}
		}
		for _, hit := range hits {
			if len(reasons) >= maxNotifiedReasons {
				break
			}
			reasons = append(reasons, advisoryReason(row, hit))
		}
	}
	return reasons
}

func advisoryReason(row queries.ComposerAdvisory, hit sbom.Match) notify.StatusReason {
	severity := displaySeverity(row)
	params := map[string]any{
		"title":    row.Title,
		"id":       advisoryPublicID(row),
		"severity": severity,
	}
	if hit.PackageName != "" {
		params["package"] = hit.PackageName
		params["installedVersion"] = hit.InstalledVersion
		params["current"] = hit.InstalledVersion
	}
	if fix := recommendedFix(row, hit.InstalledVersion); fix != "" {
		params["recommended"] = fix
	}

	return notify.StatusReason{
		Level:  severityLevel(severity),
		Key:    "check.security.advisoryAlert",
		Params: params,
		Source: "security",
	}
}

// advisoryPublicID prefers the identifier an operator can look up: CVE, then
// GHSA, then the catalog's own id.
func advisoryPublicID(row queries.ComposerAdvisory) string {
	if row.Cve != nil && strings.TrimSpace(*row.Cve) != "" {
		return strings.TrimSpace(*row.Cve)
	}
	if row.GhsaID != nil && strings.TrimSpace(*row.GhsaID) != "" {
		return strings.TrimSpace(*row.GhsaID)
	}
	return row.AdvisoryID
}

func displaySeverity(row queries.ComposerAdvisory) string {
	raw := ""
	if row.SeverityOverride != nil && strings.TrimSpace(*row.SeverityOverride) != "" {
		raw = *row.SeverityOverride
	} else if row.Severity != nil {
		raw = *row.Severity
	}
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "Unknown"
	}
	return titleCaseWord(raw)
}

func titleCaseWord(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func severityLevel(severity string) string {
	switch strings.ToLower(severity) {
	case "critical", "high":
		return "red"
	default:
		return "yellow"
	}
}

// recommendedFix is the least-disruptive upgrade we can name: an admin-set
// recommendation first, then the GitHub first-patched version for the
// installed Shopware line.
func recommendedFix(row queries.ComposerAdvisory, installedVersion string) string {
	if row.RecommendedUpgrade != nil && strings.TrimSpace(*row.RecommendedUpgrade) != "" {
		return strings.TrimSpace(*row.RecommendedUpgrade)
	}
	if len(row.FirstPatchedVersions) == 0 {
		return ""
	}
	var byLine map[string]string
	if err := json.Unmarshal(row.FirstPatchedVersions, &byLine); err != nil || len(byLine) == 0 {
		return ""
	}
	if line := shopwareLine(installedVersion); line != "" {
		if v := strings.TrimSpace(byLine[line]); v != "" {
			return v
		}
	}
	best := ""
	for _, v := range byLine {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if best == "" || v < best {
			best = v
		}
	}
	return best
}
