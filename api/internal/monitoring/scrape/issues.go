package scrape

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"sort"
	"strconv"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
)

// Scrape dataset vocabulary: one entry per independently fetched piece of
// shop data. An issue belongs to exactly one dataset, and a dataset being
// "evaluated" in a scrape is what allows its previous issues to resolve.
const (
	datasetAuth           = "auth"
	datasetConfig         = "config"
	datasetPlugins        = "plugins"
	datasetApps           = "apps"
	datasetScheduledTasks = "scheduled_tasks"
	datasetQueue          = "queue"
	datasetCacheInfo      = "cache_info"
	datasetFroshTools     = "frosh_tools"
	datasetSBOM           = "sbom"
	datasetAccounts       = "accounts"
)

// transientNotifyThreshold is how many consecutive scrapes a transient
// failure (timeout, shop 5xx, …) must survive before owners are told. At the
// hourly cadence that is roughly three hours of continuous failure.
const transientNotifyThreshold = 3

// backoffThreshold mirrors the scheduler's connection-issue threshold: when
// the consecutive total-failure count reaches it, scheduled scrapes demote
// from hourly to daily and owners get the "paused" notice. Keep in sync with
// internal/jobs (scheduler.go).
const backoffThreshold = 3

// scrapeIssueInput is one classified refresh problem observed in a scrape.
type scrapeIssueInput struct {
	key           string
	kind          string
	dataset       string
	messageKey    string
	params        map[string]any
	deterministic bool
}

// issueFromError classifies an upstream failure into an owner-facing issue.
func issueFromError(dataset string, err error) scrapeIssueInput {
	return issueFromFailure(dataset, shopware.ClassifyFailure(err))
}

func issueFromFailure(dataset string, f shopware.Failure) scrapeIssueInput {
	kind := string(f.Class)
	params := map[string]any{"dataset": dataset}
	if len(f.MissingPrivileges) > 0 {
		params["privileges"] = strings.Join(f.MissingPrivileges, ", ")
	}
	return scrapeIssueInput{
		key:           dataset + "." + kind,
		kind:          kind,
		dataset:       dataset,
		params:        params,
		deterministic: f.Deterministic(),
	}
}

// resolveMessageKey picks the most specific catalog key that exists:
// "scrapeIssue.<dataset>.<kind>" (e.g. per-dataset missing-privilege texts),
// falling back to the generic "scrapeIssue.<kind>".
func (h *Service) resolveMessageKey(in scrapeIssueInput) string {
	specific := "scrapeIssue." + in.dataset + "." + in.kind
	if h.notifier.Translator().Has(specific) {
		return specific
	}
	return "scrapeIssue." + in.kind
}

// shouldNotifyIssue is the notification policy: deterministic failures tell
// owners on first sight (they never heal on their own), transient ones only
// after transientNotifyThreshold consecutive scrapes, and an occurrence never
// notifies twice.
func shouldNotifyIssue(deterministic bool, consecutiveCount int32, alreadyNotified bool) bool {
	if alreadyNotified {
		return false
	}
	if deterministic {
		return true
	}
	return consecutiveCount >= transientNotifyThreshold
}

type issueRecordResult struct {
	// toNotify are the issues crossing the notification policy in this scrape.
	toNotify []scrapeIssueInput
	// recoveredNotified reports that at least one previously *notified* issue
	// resolved in this scrape.
	recoveredNotified bool
}

// recordScrapeIssues persists the issues observed in one scrape and resolves
// stored issues whose dataset was evaluated without reproducing them.
// Datasets a scrape never reached (e.g. everything after an auth failure)
// must stay out of evaluated, or their unresolved problems would silently
// disappear.
func (h *Service) recordScrapeIssues(ctx context.Context, env queries.GetAllEnvironmentsRow, current []scrapeIssueInput, evaluated map[string]bool) issueRecordResult {
	var res issueRecordResult

	old, err := h.queries.GetEnvironmentScrapeIssues(ctx, env.ID)
	if err != nil {
		// Without the previous state we cannot tell a new problem from a
		// recovery; skip this scrape's issue bookkeeping rather than guessing.
		slog.WarnContext(ctx, "failed to load scrape issues", "environmentId", env.ID, "error", err)
		return res
	}

	currentKeys := make(map[string]bool, len(current))
	for i := range current {
		in := &current[i]
		in.messageKey = h.resolveMessageKey(*in)
		currentKeys[in.key] = true

		params, err := json.Marshal(in.params)
		if err != nil {
			params = []byte("{}")
		}
		row, err := h.queries.UpsertEnvironmentScrapeIssue(ctx, queries.UpsertEnvironmentScrapeIssueParams{
			EnvironmentID: env.ID,
			IssueKey:      in.key,
			Kind:          in.kind,
			Dataset:       in.dataset,
			MessageKey:    in.messageKey,
			Params:        params,
		})
		if err != nil {
			slog.WarnContext(ctx, "failed to record scrape issue", "environmentId", env.ID, "issue", in.key, "error", err)
			continue
		}
		if shouldNotifyIssue(in.deterministic, row.ConsecutiveCount, row.NotifiedAt.Valid) {
			res.toNotify = append(res.toNotify, *in)
		}
	}

	var resolvedKeys []string
	for _, o := range old {
		if evaluated[o.Dataset] && !currentKeys[o.IssueKey] {
			resolvedKeys = append(resolvedKeys, o.IssueKey)
			if o.NotifiedAt.Valid {
				res.recoveredNotified = true
			}
		}
	}
	if len(resolvedKeys) > 0 {
		if err := h.queries.DeleteEnvironmentScrapeIssuesByKeys(ctx, queries.DeleteEnvironmentScrapeIssuesByKeysParams{
			EnvironmentID: env.ID,
			Column2:       resolvedKeys,
		}); err != nil {
			slog.WarnContext(ctx, "failed to resolve scrape issues", "environmentId", env.ID, "error", err)
			// The rows are still there, so claiming a recovery would be wrong.
			res.recoveredNotified = false
		}
	}

	return res
}

// handleTotalScrapeFailure records a scrape that produced no data at all
// (auth or config stage) and applies the notification policy, including the
// one-time "refresh slowed down" notice when the consecutive-failure count
// crosses the scheduler backoff threshold.
func (h *Service) handleTotalScrapeFailure(ctx context.Context, env queries.GetAllEnvironmentsRow, dataset string, err error, evaluated map[string]bool) {
	current := []scrapeIssueInput{issueFromError(dataset, err)}
	res := h.recordScrapeIssues(ctx, env, current, evaluated)

	// env.ConnectionIssueCount predates the increment
	// UpdateEnvironmentScrapeError just applied for this failure.
	if env.ConnectionIssueCount+1 == backoffThreshold {
		// The paused notice replaces the plain issue notice for this scrape:
		// it carries the same reason plus the backoff explanation.
		h.dispatchScrapePaused(ctx, env, current)
		return
	}
	if len(res.toNotify) > 0 {
		h.dispatchScrapeIssues(ctx, env, res.toNotify)
	}
}

func issueReasons(issues []scrapeIssueInput) []notify.StatusReason {
	reasons := make([]notify.StatusReason, 0, len(issues))
	for _, in := range issues {
		source := "Shopware"
		if in.dataset == datasetFroshTools || in.dataset == datasetSBOM {
			source = "FroshTools"
		}
		reasons = append(reasons, notify.StatusReason{
			Level:  "red",
			Key:    in.messageKey,
			Params: in.params,
			Source: source,
		})
	}
	return reasons
}

func (h *Service) markIssuesNotified(ctx context.Context, envID int32, issues []scrapeIssueInput) {
	for _, in := range issues {
		if err := h.queries.MarkEnvironmentScrapeIssueNotified(ctx, queries.MarkEnvironmentScrapeIssueNotifiedParams{
			EnvironmentID: envID,
			IssueKey:      in.key,
		}); err != nil {
			slog.WarnContext(ctx, "failed to mark scrape issue notified", "environmentId", envID, "issue", in.key, "error", err)
		}
	}
}

// dispatchScrapeIssues sends one batched notification for all issues that
// crossed the policy in this scrape and marks them notified only when a
// channel actually delivered, so a failed send is retried next scrape.
func (h *Service) dispatchScrapeIssues(ctx context.Context, env queries.GetAllEnvironmentsRow, issues []scrapeIssueInput) {
	keys := make([]string, 0, len(issues))
	for _, in := range issues {
		keys = append(keys, in.key)
	}
	sort.Strings(keys)
	digest := sha256.Sum256([]byte(strings.Join(keys, ",")))

	result := h.notifier.Dispatch(ctx, notify.Event{
		Type:       notify.EventScrapeIssue,
		Level:      notify.LevelError,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(env.ID)),
		OrgID:      env.OrganizationID,
		DedupKey:   fmt.Sprintf("environment.scrape-issue.%d.%s", env.ID, hex.EncodeToString(digest[:6])),
		TitleKey:   "notification.scrapeIssue.title",
		MessageKey: "notification.scrapeIssue.message",
		Params: map[string]any{
			"name": environmentDisplayName(env.Name, env.ShopName),
		},
		Reasons: issueReasons(issues),
		Link:    environmentLink(env.ID),
	}, recipientsOrNone(h.environmentRecipients(ctx, env)))

	if result.Delivery() {
		h.markIssuesNotified(ctx, env.ID, issues)
	}
}

// dispatchScrapePaused tells owners that repeated total failures demoted the
// environment to daily retries. The underlying issues count as notified: the
// paused message carries their reasons.
func (h *Service) dispatchScrapePaused(ctx context.Context, env queries.GetAllEnvironmentsRow, issues []scrapeIssueInput) {
	result := h.notifier.Dispatch(ctx, notify.Event{
		Type:       notify.EventScrapePaused,
		Level:      notify.LevelError,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(env.ID)),
		OrgID:      env.OrganizationID,
		DedupKey:   fmt.Sprintf("environment.scrape-paused.%d", env.ID),
		TitleKey:   "notification.scrapePaused.title",
		MessageKey: "notification.scrapePaused.message",
		Params: map[string]any{
			"name": environmentDisplayName(env.Name, env.ShopName),
		},
		Reasons: issueReasons(issues),
		Link:    environmentLink(env.ID),
	}, recipientsOrNone(h.environmentRecipients(ctx, env)))

	if result.Delivery() {
		h.markIssuesNotified(ctx, env.ID, issues)
	}
}

// dispatchScrapeRecovered tells owners a previously reported refresh problem
// is gone. Callers only invoke it on a scrape with no remaining issues, so
// "resolved" is never claimed while another problem persists.
func (h *Service) dispatchScrapeRecovered(ctx context.Context, env queries.GetAllEnvironmentsRow) {
	h.notifier.Dispatch(ctx, notify.Event{
		Type:       notify.EventScrapeRecovered,
		Level:      notify.LevelInfo,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(env.ID)),
		OrgID:      env.OrganizationID,
		DedupKey:   fmt.Sprintf("environment.scrape-recovered.%d", env.ID),
		TitleKey:   "notification.scrapeRecovered.title",
		MessageKey: "notification.scrapeRecovered.message",
		Params: map[string]any{
			"name": environmentDisplayName(env.Name, env.ShopName),
		},
		Link: environmentLink(env.ID),
	}, recipientsOrNone(h.environmentRecipients(ctx, env)))
}

// recipientsOrNone drops a subscriber lookup error for the dispatch-only paths:
// they persist no "was notified" marker, so a failed lookup costs one round of
// notifications and the next scrape tries again.
func recipientsOrNone(recipients []notify.Recipient, _ error) []notify.Recipient {
	return recipients
}
