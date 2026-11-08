package handler_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSharedDestinationPreferencePermissionsAndInheritance(t *testing.T) {
	env := testutil.Setup(t)
	owner := env.SeedUser(t, "owner", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Team", "team", "owner")
	member := env.SeedUser(t, "member", "Member", "member@example.com", "user")
	env.SeedMember(t, "org-1", "member", "member")
	other := env.SeedUser(t, "other", "Other", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other", "other", "other")
	shop := env.SeedShop(t, "org-1", "Shop")
	otherShop := env.SeedShop(t, "org-1", "Other shop")
	envID := env.SeedEnvironment(t, "org-1", shop, "Production", "https://shop.example")
	otherEnv := env.SeedEnvironment(t, "org-1", otherShop, "Other", "https://other.example")
	base := fmt.Sprintf("/api/organizations/org-1/shops/%d/notification-destinations", shop)
	require.Equal(t, 204, destinationRequest(t, env, http.MethodPost, base, owner, api.SaveNotificationDestinationRequest{Name: "Team", DSN: testDSN}).StatusCode)
	id := destinationStatus(t, env, base, owner).Destinations[0].ID
	prefs := destinationBase + "/" + id + "/preferences"
	global := map[string]any{"scopeType": "global", "scopeId": "", "eventType": "status_degraded", "channel": "destination", "enabled": false}
	require.Equal(t, 204, destinationRequest(t, env, http.MethodPut, prefs, owner, global).StatusCode)
	resp := destinationRequest(t, env, http.MethodGet, prefs, member, nil)
	require.Equal(t, 200, resp.StatusCode)
	var rows []api.NotificationPreference
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rows))
	require.Len(t, rows, 1)
	assert.False(t, rows[0].Enabled)
	assert.Equal(t, 403, destinationRequest(t, env, http.MethodPut, prefs, member, global).StatusCode)
	assert.Equal(t, 403, destinationRequest(t, env, http.MethodGet, prefs, other, nil).StatusCode)
	assert.Equal(t, 404, destinationRequest(t, env, http.MethodGet, "/api/organizations/org-2/notification-destinations/"+id+"/preferences", other, nil).StatusCode)
	override := map[string]any{"scopeType": "environment", "scopeId": fmt.Sprint(envID), "eventType": "status_degraded", "channel": "destination", "enabled": true}
	require.Equal(t, 204, destinationRequest(t, env, http.MethodPut, prefs, owner, override).StatusCode)
	override["scopeId"] = fmt.Sprint(otherEnv)
	assert.Equal(t, 404, destinationRequest(t, env, http.MethodPut, prefs, owner, override).StatusCode)
	query := fmt.Sprintf("?scopeType=environment&scopeId=%d&eventType=status_degraded&channel=destination", envID)
	assert.Equal(t, 403, destinationRequest(t, env, http.MethodDelete, prefs+query, member, nil).StatusCode)
	require.Equal(t, 204, destinationRequest(t, env, http.MethodDelete, prefs+query, owner, nil).StatusCode)
	resp = destinationRequest(t, env, http.MethodGet, prefs, owner, nil)
	require.Equal(t, 200, resp.StatusCode)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rows))
	require.Len(t, rows, 1)
	assert.Equal(t, "global", rows[0].ScopeType)
	for key, value := range map[string]any{"eventType": "unknown.event", "channel": "email", "scopeType": "organization", "scopeId": "unexpected"} {
		invalid := map[string]any{}
		for k, v := range global {
			invalid[k] = v
		}
		invalid[key] = value
		assert.Equal(t, 400, destinationRequest(t, env, http.MethodPut, prefs, owner, invalid).StatusCode, key)
	}
	// A user's personal preference endpoint must never mutate a shared channel.
	assert.Equal(t, 400, destinationRequest(t, env, http.MethodPut, "/api/account/notification-preferences", owner, global).StatusCode)
	require.Equal(t, 204, destinationRequest(t, env, http.MethodDelete, base+"/"+id, owner, nil).StatusCode)
	var count int
	require.NoError(t, env.Pool.QueryRow(t.Context(), "SELECT COUNT(*) FROM destination_notification_preference WHERE destination_id=$1", id).Scan(&count))
	assert.Zero(t, count)
}
