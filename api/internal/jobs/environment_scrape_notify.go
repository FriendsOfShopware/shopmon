package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strconv"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
)

// environmentLink builds the frontend route reference for an environment detail
// page, shared by all environment notifications.
func environmentLink(envID int32) notify.Link {
	return notify.Link{
		Name: "account.environments.detail",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(envID)),
		},
	}
}

// environmentRecipients resolves the subscribers of an environment into notify
// recipients. It returns nil (and logs) on error so callers can simply skip.
func (h *EnvironmentScrapeHandler) environmentRecipients(ctx context.Context, env queries.GetAllEnvironmentsRow) []notify.Recipient {
	subscribers, err := h.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: env.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(env.ID)),
	})
	if err != nil {
		slog.Warn("failed to get notification subscribers", "environmentId", env.ID, "error", err)
		return nil
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
	return recipients
}

// statusTransition captures a recorded status change so the caller can persist
// it to the timeline within the scrape transaction.
type statusTransition struct {
	oldStatus string
	newStatus string
	reasons   []notify.StatusReason
}

// handleStatusTransition detects a status change (in either direction),
// dispatches the appropriate notification with the checks that caused it, and
// returns the transition for timeline persistence. It returns nil when the
// status is unchanged.
func (h *EnvironmentScrapeHandler) handleStatusTransition(ctx context.Context, env queries.GetAllEnvironmentsRow, oldChecks []queries.EnvironmentCheck, newChecks []checker.Check, newStatus string) *statusTransition {
	envDetail, err := h.queries.GetEnvironmentByID(ctx, env.ID)
	if err != nil {
		slog.Warn("failed to get environment details for status comparison", "environmentId", env.ID, "error", err)
		return nil
	}

	oldStatus := envDetail.Status
	oldWeight := statusWeight[oldStatus]
	newWeight := statusWeight[newStatus]
	if newWeight == oldWeight {
		return nil // unchanged
	}

	degraded := newWeight > oldWeight
	reasons := computeStatusReasons(oldChecks, newChecks, degraded)

	ev := notify.Event{
		Level:     notify.LevelInfo,
		ScopeType: notify.ScopeEnvironment,
		ScopeID:   strconv.Itoa(int(env.ID)),
		OrgID:     env.OrganizationID,
		Params: map[string]any{
			"name": env.Name,
			"from": oldStatus,
			"to":   newStatus,
		},
		Reasons: reasons,
		Link:    environmentLink(env.ID),
	}

	if degraded {
		ev.Type = notify.EventStatusDegraded
		ev.Level = notify.LevelWarning
		ev.DedupKey = fmt.Sprintf("environment.change-status.%d", env.ID)
		ev.TitleKey = "notification.statusDegraded.title"
		ev.MessageKey = "notification.statusDegraded.message"
	} else {
		ev.Type = notify.EventStatusRecovered
		ev.DedupKey = fmt.Sprintf("environment.recover-status.%d", env.ID)
		ev.TitleKey = "notification.statusRecovered.title"
		ev.MessageKey = "notification.statusRecovered.message"
	}

	h.notifier.Dispatch(ctx, ev, h.environmentRecipients(ctx, env))

	return &statusTransition{oldStatus: oldStatus, newStatus: newStatus, reasons: reasons}
}

// checkWeight orders check levels for comparison (higher is worse).
func checkWeight(level string) int {
	switch level {
	case "red":
		return 3
	case "yellow":
		return 2
	default:
		return 1
	}
}

// computeStatusReasons returns the checks responsible for a status transition:
// the checks that worsened (degradation) or improved (recovery).
func computeStatusReasons(oldChecks []queries.EnvironmentCheck, newChecks []checker.Check, degraded bool) []notify.StatusReason {
	reasons := []notify.StatusReason{}

	if degraded {
		oldLevel := make(map[string]string, len(oldChecks))
		for _, c := range oldChecks {
			oldLevel[c.CheckID] = c.Level
		}
		for _, c := range newChecks {
			prev := oldLevel[c.ID]
			if prev == "" {
				prev = "green"
			}
			if string(c.Level) != "green" && checkWeight(string(c.Level)) > checkWeight(prev) {
				reasons = append(reasons, notify.StatusReason{
					Level:  string(c.Level),
					Key:    c.MessageKey,
					Params: c.MessageParams,
					Source: c.Source,
				})
			}
		}
		return reasons
	}

	newLevel := make(map[string]string, len(newChecks))
	for _, c := range newChecks {
		newLevel[c.ID] = string(c.Level)
	}
	for _, c := range oldChecks {
		next := newLevel[c.CheckID]
		if next == "" {
			next = "green"
		}
		if c.Level != "green" && checkWeight(c.Level) > checkWeight(next) {
			reasons = append(reasons, notify.StatusReason{
				Level:  c.Level,
				Key:    derefString(c.MessageKey),
				Params: decodeParams(c.Params),
				Source: c.Source,
			})
		}
	}
	return reasons
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func decodeParams(raw []byte) map[string]any {
	if len(raw) == 0 {
		return nil
	}
	var params map[string]any
	if err := json.Unmarshal(raw, &params); err != nil {
		return nil
	}
	return params
}

// notifyAuthError dispatches notifications when authentication fails.
func (h *EnvironmentScrapeHandler) notifyAuthError(ctx context.Context, env queries.GetAllEnvironmentsRow, authErr error) {
	errMsg := fmt.Sprintf("%v", authErr)
	if len(errMsg) > 50 {
		errMsg = errMsg[:50] + "..."
	}

	h.notifier.Dispatch(ctx, notify.Event{
		Type:       notify.EventAuthError,
		Level:      notify.LevelError,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(env.ID)),
		OrgID:      env.OrganizationID,
		DedupKey:   fmt.Sprintf("environment.update-auth-error.%d", env.ID),
		TitleKey:   "notification.authError.title",
		MessageKey: "notification.authError.message",
		Params: map[string]any{
			"name":  env.Name,
			"error": errMsg,
		},
		Link: environmentLink(env.ID),
	}, h.environmentRecipients(ctx, env))
}

// notifyDataFetchError dispatches notifications when data fetching fails.
func (h *EnvironmentScrapeHandler) notifyDataFetchError(ctx context.Context, env queries.GetAllEnvironmentsRow, _ string) {
	h.notifier.Dispatch(ctx, notify.Event{
		Type:       notify.EventDataFetchError,
		Level:      notify.LevelError,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(env.ID)),
		OrgID:      env.OrganizationID,
		DedupKey:   fmt.Sprintf("environment.not.updated_%d", env.ID),
		TitleKey:   "notification.dataFetchError.title",
		MessageKey: "notification.dataFetchError.message",
		Params: map[string]any{
			"name": env.Name,
		},
		Link: environmentLink(env.ID),
	}, h.environmentRecipients(ctx, env))
}
