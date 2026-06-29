package notify

import (
	"context"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func channelsOf(prefs []queries.NotificationPreference, defaults []ChannelName) []ChannelName {
	res := queryResolver{q: fakePrefStore{prefs: prefs}}
	return res.Channels(context.Background(), Recipient{ID: "u1"}, statusEvent(), defaults)
}

func TestListEventTypes(t *testing.T) {
	types := ListEventTypes()
	require.Len(t, types, len(orderedEventTypes))
	assert.Equal(t, string(EventStatusDegraded), types[0].Type)

	// Every event type exposes at least one default channel.
	for _, et := range types {
		assert.NotEmpty(t, et.DefaultChannels, et.Type)
	}
}

func TestResolverDefaultsWhenNoPreferences(t *testing.T) {
	got := channelsOf(nil, []ChannelName{ChannelInApp, ChannelEmail})
	assert.ElementsMatch(t, []ChannelName{ChannelInApp, ChannelEmail}, got)
}

func TestResolverGlobalChannelDisable(t *testing.T) {
	// Email disabled globally for all events.
	prefs := []queries.NotificationPreference{
		pref("global", "", "", "email", false),
	}
	got := channelsOf(prefs, []ChannelName{ChannelInApp, ChannelEmail})
	assert.NotContains(t, got, ChannelEmail, "email disabled globally")
	assert.Contains(t, got, ChannelInApp, "in_app remains enabled")
}

func TestResolverEnvironmentOverridesGlobal(t *testing.T) {
	// Email off globally, but explicitly on for this environment (more specific).
	prefs := []queries.NotificationPreference{
		pref("global", "", "", "email", false),
		pref("environment", "7", "", "email", true),
	}
	got := channelsOf(prefs, []ChannelName{ChannelInApp, ChannelEmail})
	assert.Contains(t, got, ChannelEmail, "environment-scoped enable wins over global disable")
}

func TestResolverEventSpecificOverridesWildcard(t *testing.T) {
	// Email on for all events at env scope, but off specifically for degraded.
	prefs := []queries.NotificationPreference{
		pref("environment", "7", "", "email", true),
		pref("environment", "7", string(EventStatusDegraded), "email", false),
	}
	got := channelsOf(prefs, []ChannelName{ChannelInApp, ChannelEmail})
	assert.NotContains(t, got, ChannelEmail, "event-specific disable wins over wildcard enable")
}

func TestResolverNeverAddsChannelsOutsideDefaults(t *testing.T) {
	// The resolver only delivers on the channels it is given; an enabled pref for
	// a channel outside the defaults must not add it.
	prefs := []queries.NotificationPreference{
		pref("environment", "7", "", "email", true),
	}
	got := channelsOf(prefs, []ChannelName{ChannelInApp})
	assert.NotContains(t, got, ChannelEmail, "channel outside defaults is never added")
	assert.Contains(t, got, ChannelInApp)
}
