package uptime

import (
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
)

// Notifier delivers uptime incident alerts to environment subscribers through
// the shared notify dispatcher, honouring per-user channel preferences.
type Notifier struct {
	queries    *queries.Queries
	dispatcher *notify.Dispatcher
}

var _ Alerter = (*Notifier)(nil)

func NewNotifier(q *queries.Queries, dispatcher *notify.Dispatcher) *Notifier {
	return &Notifier{queries: q, dispatcher: dispatcher}
}

func (n *Notifier) AlertDown(ctx context.Context, m Monitor, eventID int32, res ProbeResult) {
	reason := res.Err
	if reason == "" && res.StatusCode > 0 {
		reason = fmt.Sprintf("status %d", res.StatusCode)
	}
	if reason == "" {
		reason = "unknown"
	}

	ev := notify.Event{
		Type:       notify.EventUptimeDown,
		Level:      notify.LevelError,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(m.EnvironmentID)),
		OrgID:      m.OrganizationID,
		DedupKey:   fmt.Sprintf("uptime-down-%d-%d", m.EnvironmentID, eventID),
		TitleKey:   "notification.uptimeDown.title",
		MessageKey: "notification.uptimeDown.message",
		Params: map[string]any{
			"name":   m.EnvironmentName,
			"url":    m.URL,
			"reason": reason,
		},
		Link: environmentLink(m.EnvironmentID),
	}
	dr := n.dispatcher.Dispatch(ctx, ev, n.recipients(ctx, m))
	slog.Info("uptime: down alert dispatched",
		"environmentId", m.EnvironmentID, "delivered", dr.Delivered,
		"failed", dr.Failed, "skipped", dr.Skipped)
}

func (n *Notifier) AlertRecovered(ctx context.Context, m Monitor, eventID int32, downFor time.Duration) {
	ev := notify.Event{
		Type:       notify.EventUptimeRecovered,
		Level:      notify.LevelInfo,
		ScopeType:  notify.ScopeEnvironment,
		ScopeID:    strconv.Itoa(int(m.EnvironmentID)),
		OrgID:      m.OrganizationID,
		DedupKey:   fmt.Sprintf("uptime-recovered-%d-%d", m.EnvironmentID, eventID),
		TitleKey:   "notification.uptimeRecovered.title",
		MessageKey: "notification.uptimeRecovered.message",
		Params: map[string]any{
			"name":     m.EnvironmentName,
			"url":      m.URL,
			"duration": humanDuration(downFor),
		},
		Link: environmentLink(m.EnvironmentID),
	}
	dr := n.dispatcher.Dispatch(ctx, ev, n.recipients(ctx, m))
	slog.Info("uptime: recovery alert dispatched",
		"environmentId", m.EnvironmentID, "delivered", dr.Delivered,
		"failed", dr.Failed, "skipped", dr.Skipped)
}

// recipients resolves the environment's subscribers into notify recipients.
// It returns nil (and logs) on error so alerting degrades gracefully.
func (n *Notifier) recipients(ctx context.Context, m Monitor) []notify.Recipient {
	subscribers, err := n.queries.GetEnvironmentNotificationSubscribers(ctx, queries.GetEnvironmentNotificationSubscribersParams{
		OrganizationID: m.OrganizationID,
		EnvironmentID:  strconv.Itoa(int(m.EnvironmentID)),
	})
	if err != nil {
		slog.Warn("uptime: failed to get notification subscribers", "environmentId", m.EnvironmentID, "error", err)
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

// environmentLink builds the frontend route reference for the environment
// detail page, matching the scrape notifications.
func environmentLink(envID int32) notify.Link {
	return notify.Link{
		Name: "account.environments.detail",
		Params: map[string]string{
			"environmentId": strconv.Itoa(int(envID)),
		},
	}
}

// humanDuration renders a duration compactly for notification text.
func humanDuration(d time.Duration) string {
	if d < time.Minute {
		return fmt.Sprintf("%ds", int(d.Seconds()))
	}
	d = d.Round(time.Second)
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	switch {
	case h > 0:
		return fmt.Sprintf("%dh %dm", h, m)
	case m > 0:
		return fmt.Sprintf("%dm %ds", m, s)
	default:
		return fmt.Sprintf("%ds", s)
	}
}
