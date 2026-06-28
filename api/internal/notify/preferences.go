package notify

import (
	"context"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

// PreferenceResolver decides which channels a recipient receives a given event
// on. It is the seam for user-configurable notification settings: the default
// implementation honours the registry defaults, and the database-backed
// resolver applies the per-user, per-scope, per-event override chain without
// touching dispatch or channel code.
type PreferenceResolver interface {
	Channels(ctx context.Context, r Recipient, ev Event, defaults []ChannelName) []ChannelName
}

// defaultResolver delivers an event on its registry default channels (used in
// tests and as the fail-open fallback).
type defaultResolver struct{}

func (defaultResolver) Channels(_ context.Context, _ Recipient, _ Event, defaults []ChannelName) []ChannelName {
	return defaults
}

// prefLister is the narrow query surface the resolver needs (satisfied by
// *queries.Queries), kept small for testability.
type prefLister interface {
	ListNotificationPreferences(ctx context.Context, userID string) ([]queries.NotificationPreference, error)
}

// queryResolver applies stored preferences. Channel selection is the
// intersection of the event's registry default channels with the recipient's
// effective per-channel preference (most-specific scope wins, default enabled).
// Channels outside the registry defaults are never used for an event type, so
// e.g. data-fetch errors stay in-app even if the user enabled email elsewhere.
type queryResolver struct {
	q prefLister
}

func (r queryResolver) Channels(ctx context.Context, rec Recipient, ev Event, defaults []ChannelName) []ChannelName {
	prefs, err := r.q.ListNotificationPreferences(ctx, rec.ID)
	if err != nil {
		// Fail open to the registry defaults so a preferences read error does
		// not silently drop notifications.
		slog.Warn("notify: failed to load preferences, using defaults", "userId", rec.ID, "error", err)
		return defaults
	}

	out := make([]ChannelName, 0, len(defaults))
	for _, ch := range defaults {
		if channelEnabled(prefs, ch, ev) {
			out = append(out, ch)
		}
	}
	return out
}

// channelEnabled resolves the effective enabled state for a channel by walking
// the scope chain from most to least specific; the first matching row wins.
// When no row matches, the channel defaults to enabled.
func channelEnabled(prefs []queries.NotificationPreference, ch ChannelName, ev Event) bool {
	type scope struct {
		scopeType string
		scopeID   string
		eventType string
	}
	chain := []scope{
		{string(ScopeEnvironment), ev.ScopeID, string(ev.Type)},
		{string(ScopeEnvironment), ev.ScopeID, ""},
		{string(ScopeOrganization), ev.OrgID, string(ev.Type)},
		{string(ScopeOrganization), ev.OrgID, ""},
		{string(ScopeGlobal), "", string(ev.Type)},
		{string(ScopeGlobal), "", ""},
	}

	for _, s := range chain {
		for i := range prefs {
			p := prefs[i]
			if p.Channel == string(ch) &&
				p.ScopeType == s.scopeType &&
				p.ScopeID == s.scopeID &&
				p.EventType == s.eventType {
				return p.Enabled
			}
		}
	}
	return true
}
