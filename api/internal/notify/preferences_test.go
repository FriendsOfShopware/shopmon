package notify

import (
	"context"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type fakePrefStore struct {
	prefs []queries.NotificationPreference
}

func (f fakePrefStore) ListNotificationPreferences(_ context.Context, _ string) ([]queries.NotificationPreference, error) {
	return f.prefs, nil
}

func pref(scopeType, scopeID, eventType, channel string, enabled bool) queries.NotificationPreference {
	return queries.NotificationPreference{
		ScopeType: scopeType,
		ScopeID:   scopeID,
		EventType: eventType,
		Channel:   channel,
		Enabled:   enabled,
	}
}

func statusEvent() Event {
	return Event{
		Type:      EventStatusDegraded,
		ScopeType: ScopeEnvironment,
		ScopeID:   "7",
		OrgID:     "org1",
	}
}

func channelsOf(r PreferenceResolver, prefs []queries.NotificationPreference, defaults []ChannelName) []ChannelName {
	res := queryResolver{q: fakePrefStore{prefs: prefs}}
	return res.Channels(context.Background(), Recipient{ID: "u1"}, statusEvent(), defaults)
}

func contains(chs []ChannelName, want ChannelName) bool {
	for _, c := range chs {
		if c == want {
			return true
		}
	}
	return false
}

func TestListEventTypes(t *testing.T) {
	types := ListEventTypes()
	if len(types) != len(orderedEventTypes) {
		t.Fatalf("expected %d event types, got %d", len(orderedEventTypes), len(types))
	}
	if types[0].Type != string(EventStatusDegraded) {
		t.Fatalf("expected first event type %q, got %q", EventStatusDegraded, types[0].Type)
	}
	// data_fetch_error is in-app only; the metadata must reflect that.
	for _, et := range types {
		if et.Type == string(EventDataFetchError) {
			if len(et.DefaultChannels) != 1 || et.DefaultChannels[0] != string(ChannelInApp) {
				t.Fatalf("data_fetch_error default channels = %v, want [in_app]", et.DefaultChannels)
			}
		}
	}
}

func TestResolverDefaultsWhenNoPreferences(t *testing.T) {
	defaults := []ChannelName{ChannelInApp, ChannelEmail}
	got := channelsOf(nil, nil, defaults)
	if len(got) != 2 {
		t.Fatalf("expected both default channels, got %v", got)
	}
}

func TestResolverGlobalChannelDisable(t *testing.T) {
	defaults := []ChannelName{ChannelInApp, ChannelEmail}
	// Email disabled globally for all events.
	prefs := []queries.NotificationPreference{
		pref("global", "", "", "email", false),
	}
	got := channelsOf(nil, prefs, defaults)
	if contains(got, ChannelEmail) {
		t.Fatalf("email should be disabled globally, got %v", got)
	}
	if !contains(got, ChannelInApp) {
		t.Fatalf("in_app should remain enabled, got %v", got)
	}
}

func TestResolverEnvironmentOverridesGlobal(t *testing.T) {
	defaults := []ChannelName{ChannelInApp, ChannelEmail}
	// Email off globally, but explicitly on for this environment (more specific).
	prefs := []queries.NotificationPreference{
		pref("global", "", "", "email", false),
		pref("environment", "7", "", "email", true),
	}
	got := channelsOf(nil, prefs, defaults)
	if !contains(got, ChannelEmail) {
		t.Fatalf("environment-scoped enable should win over global disable, got %v", got)
	}
}

func TestResolverEventSpecificOverridesWildcard(t *testing.T) {
	defaults := []ChannelName{ChannelInApp, ChannelEmail}
	// Email on for all events at env scope, but off specifically for degraded.
	prefs := []queries.NotificationPreference{
		pref("environment", "7", "", "email", true),
		pref("environment", "7", string(EventStatusDegraded), "email", false),
	}
	got := channelsOf(nil, prefs, defaults)
	if contains(got, ChannelEmail) {
		t.Fatalf("event-specific disable should win over wildcard enable, got %v", got)
	}
}

func TestResolverNeverAddsChannelsOutsideDefaults(t *testing.T) {
	// data_fetch defaults to in_app only; an enabled email pref must not add email.
	res := queryResolver{q: fakePrefStore{prefs: []queries.NotificationPreference{
		pref("environment", "7", "", "email", true),
	}}}
	ev := statusEvent()
	ev.Type = EventDataFetchError
	got := res.Channels(context.Background(), Recipient{ID: "u1"}, ev, []ChannelName{ChannelInApp})
	if contains(got, ChannelEmail) {
		t.Fatalf("email must not be added for an in_app-only event, got %v", got)
	}
	if !contains(got, ChannelInApp) {
		t.Fatalf("in_app should be present, got %v", got)
	}
}
