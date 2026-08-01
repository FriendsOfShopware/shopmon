package advisory

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
)

// maxNotifiedTitles caps how many advisory titles are named in one alert. The
// rest are covered by the count, so a large burst stays readable.
const maxNotifiedTitles = 3

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

	titles, err := s.advisoryTitles(ctx, fresh)
	if err != nil {
		slog.WarnContext(ctx, "failed to load advisory titles for alert", "error", err)
	}

	name := env.Name
	if env.ShopName != nil && *env.ShopName != "" {
		name = *env.ShopName + " · " + env.Name
	}

	s.notifier.Dispatch(ctx, notify.Event{
		Type:      notify.EventAdvisoryDetected,
		Level:     notify.LevelWarning,
		ScopeType: notify.ScopeEnvironment,
		ScopeID:   strconv.Itoa(int(environmentID)),
		OrgID:     env.OrganizationID,
		// Keyed per environment rather than per advisory, so the in-app upsert
		// collapses a burst of new advisories into one entry.
		DedupKey:   "advisory_detected:" + strconv.Itoa(int(environmentID)),
		TitleKey:   "notification.advisoryDetected.title",
		MessageKey: "notification.advisoryDetected.message",
		Params: map[string]any{
			"name":  name,
			"count": len(fresh),
			"first": titles,
		},
		Link: notify.Link{
			Name:   "account.environments.detail",
			Params: map[string]string{"environmentId": strconv.Itoa(int(environmentID))},
		},
	}, recipients)

	s.markNotified(ctx, environmentID, fresh)
}

// markNotified records that subscribers have been told about these advisories.
// Written only after dispatch, so an alert lost to a transient failure is
// retried on the next pass rather than being silently swallowed.
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

// advisoryTitles returns a short, human-readable summary of the first few new
// advisories so the alert says what was found rather than only how many.
func (s *Service) advisoryTitles(ctx context.Context, advisoryIDs []string) (string, error) {
	limit := min(len(advisoryIDs), maxNotifiedTitles)

	titles := make([]string, 0, limit)
	for _, id := range advisoryIDs[:limit] {
		row, err := s.queries.GetComposerAdvisory(ctx, id)
		if err != nil {
			return "", err
		}
		titles = append(titles, row.Title)
	}

	out := ""
	for i, title := range titles {
		if i > 0 {
			out += ", "
		}
		out += title
	}
	if len(advisoryIDs) > limit {
		out += ", …"
	}
	return out, nil
}
