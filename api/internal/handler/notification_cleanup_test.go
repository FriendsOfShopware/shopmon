package handler_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Deleting a user must cascade to their notifications and preferences. Before
// the user_notification ON DELETE cascade, a user with any notification could
// not be deleted (foreign-key violation).
func TestDeleteUserCascadesNotificationData(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "user-del", "Del", "del@example.com", "user")
	env.SeedNotification(t, "user-del", "k1", "warning", "T", "M")
	require.NoError(t, env.Queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID:    "user-del",
		ScopeType: "global",
		ScopeID:   "",
		EventType: "",
		Channel:   "email",
		Enabled:   false,
		Config:    json.RawMessage("{}"),
	}))

	require.NoError(t, env.Queries.DeleteUser(ctx, "user-del"))

	notifs, err := env.Queries.ListNotifications(ctx, "user-del")
	require.NoError(t, err)
	assert.Empty(t, notifs)

	prefs, err := env.Queries.ListNotificationPreferences(ctx, "user-del")
	require.NoError(t, err)
	assert.Empty(t, prefs)
}

// Removing a user from an organization must purge their environment-scoped
// notification data for that org's environments, while leaving global
// preferences and unrelated notifications intact.
func TestCleanupUserOrgNotifications(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "u1", "U1", "u1@example.com", "user")
	env.SeedOrganization(t, "org1", "Org", "org", "u1")
	shopID := env.SeedShop(t, "org1", "Shop")
	envID := env.SeedEnvironment(t, "org1", shopID, "Prod", "https://example.com")
	envIDStr := strconv.Itoa(envID)

	// Environment-scoped subscription marker (should be removed).
	require.NoError(t, env.Queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID: "u1", ScopeType: "environment", ScopeID: envIDStr, EventType: "", Channel: "", Enabled: true, Config: json.RawMessage("{}"),
	}))
	// Global preference (should survive).
	require.NoError(t, env.Queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID: "u1", ScopeType: "global", ScopeID: "", EventType: "", Channel: "email", Enabled: false, Config: json.RawMessage("{}"),
	}))

	// Notification linked to the environment (should be removed).
	link := json.RawMessage(fmt.Sprintf(`{"name":"account.environments.detail","params":{"environmentId":"%s"}}`, envIDStr))
	require.NoError(t, env.Queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
		UserID: "u1", Key: "environment.change-status." + envIDStr, Level: "warning",
		Title: "T", Message: "M", Params: json.RawMessage("{}"), Link: link,
	}))
	// Unrelated notification (no environment link, should survive).
	env.SeedNotification(t, "u1", "other", "warning", "T", "M")

	require.NoError(t, env.Queries.DeleteUserOrgNotificationPreferences(ctx, queries.DeleteUserOrgNotificationPreferencesParams{
		UserID: "u1", OrganizationID: "org1",
	}))
	require.NoError(t, env.Queries.DeleteUserOrgNotifications(ctx, queries.DeleteUserOrgNotificationsParams{
		UserID: "u1", OrganizationID: "org1",
	}))

	prefs, err := env.Queries.ListNotificationPreferences(ctx, "u1")
	require.NoError(t, err)
	require.Len(t, prefs, 1)
	assert.Equal(t, "global", prefs[0].ScopeType)

	notifs, err := env.Queries.ListNotifications(ctx, "u1")
	require.NoError(t, err)
	require.Len(t, notifs, 1)
	assert.Equal(t, "other", notifs[0].Key)
}

// Deleting an environment must remove the environment-scoped preferences and
// linked notifications it leaves behind (they reference it by text, not a FK).
func TestDeleteEnvironmentCleansNotifications(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	env.SeedUser(t, "u1", "U1", "u1@example.com", "user")
	env.SeedOrganization(t, "org1", "Org", "org", "u1")
	shopID := env.SeedShop(t, "org1", "Shop")
	envID := env.SeedEnvironment(t, "org1", shopID, "Prod", "https://example.com")
	envIDStr := strconv.Itoa(envID)

	require.NoError(t, env.Queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID: "u1", ScopeType: "environment", ScopeID: envIDStr, EventType: "", Channel: "", Enabled: true, Config: json.RawMessage("{}"),
	}))
	link := json.RawMessage(fmt.Sprintf(`{"name":"account.environments.detail","params":{"environmentId":"%s"}}`, envIDStr))
	require.NoError(t, env.Queries.UpsertNotification(ctx, queries.UpsertNotificationParams{
		UserID: "u1", Key: "environment.change-status." + envIDStr, Level: "warning",
		Title: "T", Message: "M", Params: json.RawMessage("{}"), Link: link,
	}))

	require.NoError(t, env.Queries.DeleteEnvironmentNotificationPreferences(ctx, envIDStr))
	require.NoError(t, env.Queries.DeleteEnvironmentNotifications(ctx, envIDStr))

	prefs, err := env.Queries.ListNotificationPreferences(ctx, "u1")
	require.NoError(t, err)
	assert.Empty(t, prefs)

	notifs, err := env.Queries.ListNotifications(ctx, "u1")
	require.NoError(t, err)
	assert.Empty(t, notifs)
}
