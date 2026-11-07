package handler_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testDSN = "slack://default?webhook_url=https%3A%2F%2Fhooks.slack.com%2Fservices%2FT000%2FB000%2Fsecret"
const destinationBase = "/api/organizations/org-1/notification-destinations"

func destinationRequest(t *testing.T, env *testutil.TestEnv, method, url, token string, body any) *http.Response {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		b, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(b)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := testutil.NewRequest(t, method, env.Server.URL+url, reader)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	t.Cleanup(func() { _ = resp.Body.Close() })
	return resp
}

func destinationStatus(t *testing.T, env *testutil.TestEnv, path, token string) api.NotificationDestinationStatus {
	t.Helper()
	resp := destinationRequest(t, env, http.MethodGet, path, token, nil)
	require.Equal(t, http.StatusOK, resp.StatusCode)
	var result api.NotificationDestinationStatus
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	return result
}
func TestNotificationDestinationLifecycle(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Shop")
	for _, base := range []string{destinationBase, fmt.Sprintf("/api/organizations/org-1/shops/%d/notification-destinations", shopID)} {
		t.Run(base, func(t *testing.T) {
			status := destinationStatus(t, env, base, token)
			require.Empty(t, status.Destinations)
			assert.True(t, status.CanManage)
			for _, dsn := range []string{testDSN, "telegram://123:secret@default?channel=12345"} {
				resp := destinationRequest(t, env, http.MethodPost, base, token, api.SaveNotificationDestinationRequest{Name: "Alerts", DSN: dsn})
				require.Equal(t, http.StatusNoContent, resp.StatusCode)
			}
			status = destinationStatus(t, env, base, token)
			require.Len(t, status.Destinations, 2)
			assert.Equal(t, "slack", status.Destinations[0].Provider)
			assert.Equal(t, "telegram", status.Destinations[1].Provider)
			raw := destinationRequest(t, env, http.MethodGet, base, token, nil)
			body, err := io.ReadAll(raw.Body)
			require.NoError(t, err)
			assert.NotContains(t, string(body), "secret")
			assert.NotContains(t, string(body), "webhook_url")
			id := status.Destinations[0].ID
			path := base + "/" + id
			var stored string
			require.NoError(t, env.Pool.QueryRow(t.Context(), "SELECT dsn FROM notification_destination WHERE id=$1", id).Scan(&stored))
			assert.NotContains(t, stored, "secret")
			assert.NotEqual(t, testDSN, stored)
			resp := destinationRequest(t, env, http.MethodPut, path, token, api.SaveNotificationDestinationRequest{Name: "Renamed"})
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "Renamed", destinationStatus(t, env, base, token).Destinations[0].Name)
			resp = destinationRequest(t, env, http.MethodPost, path+"/test", token, nil)
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, testDSN, env.Destinations.Posts[len(env.Destinations.Posts)-1].DSN)
			env.Destinations.Err = errors.New("secret in upstream error")
			resp = destinationRequest(t, env, http.MethodPost, path+"/test", token, nil)
			require.Equal(t, http.StatusBadGateway, resp.StatusCode)
			body, err = io.ReadAll(resp.Body)
			require.NoError(t, err)
			assert.NotContains(t, string(body), "secret")
			env.Destinations.Err = nil
			resp = destinationRequest(t, env, http.MethodPut, path, token, api.SaveNotificationDestinationRequest{Name: "Changed", DSN: "gotify://secret@notify.example.com"})
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			assert.Equal(t, "gotify", destinationStatus(t, env, base, token).Destinations[0].Provider)
			resp = destinationRequest(t, env, http.MethodDelete, path, token, nil)
			require.Equal(t, http.StatusNoContent, resp.StatusCode)
			require.Len(t, destinationStatus(t, env, base, token).Destinations, 1)
			resp = destinationRequest(t, env, http.MethodPost, path+"/test", token, nil)
			assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		})
	}
}
func TestNotificationDestinationPermissionsAndScope(t *testing.T) {
	env := testutil.Setup(t)
	owner := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	member := env.SeedUser(t, "user-2", "Member", "member@example.com", "user")
	env.SeedMember(t, "org-1", "user-2", "member")
	outsider := env.SeedUser(t, "user-3", "Outsider", "outsider@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other", "other", "user-3")
	foreignShop := env.SeedShop(t, "org-2", "Foreign")
	ownShop := env.SeedShop(t, "org-1", "Own")
	body := api.SaveNotificationDestinationRequest{Name: "Alerts", DSN: testDSN}
	require.Equal(t, http.StatusNoContent, destinationRequest(t, env, http.MethodPost, destinationBase, owner, body).StatusCode)
	id := destinationStatus(t, env, destinationBase, owner).Destinations[0].ID
	assert.False(t, destinationStatus(t, env, destinationBase, member).CanManage)
	for _, method := range []string{http.MethodPost, http.MethodPut, http.MethodDelete} {
		path := destinationBase + "/" + id
		if method == http.MethodPost {
			path += "/test"
		}
		assert.Equal(t, http.StatusForbidden, destinationRequest(t, env, method, path, member, body).StatusCode)
	}
	assert.Equal(t, http.StatusForbidden, destinationRequest(t, env, http.MethodPost, destinationBase, member, body).StatusCode)
	assert.Equal(t, http.StatusForbidden, destinationRequest(t, env, http.MethodGet, destinationBase, outsider, nil).StatusCode)
	assert.Equal(t, http.StatusUnauthorized, destinationRequest(t, env, http.MethodGet, destinationBase, "", nil).StatusCode)
	for _, shop := range []string{"0", "4294967296", fmt.Sprint(foreignShop)} {
		base := "/api/organizations/org-1/shops/" + shop + "/notification-destinations"
		assert.Equal(t, http.StatusNotFound, destinationRequest(t, env, http.MethodGet, base, owner, nil).StatusCode)
		assert.Equal(t, http.StatusNotFound, destinationRequest(t, env, http.MethodPost, base, owner, body).StatusCode)
	}
	// A valid shop path cannot address an organization destination or move it.
	shopBase := fmt.Sprintf("/api/organizations/org-1/shops/%d/notification-destinations/", ownShop)
	for _, base := range []string{shopBase, "/api/organizations/org-2/notification-destinations/"} {
		token := owner
		if base != shopBase {
			token = outsider
		}
		assert.Equal(t, http.StatusNotFound, destinationRequest(t, env, http.MethodPut, base+id, token, body).StatusCode)
		assert.Equal(t, http.StatusNotFound, destinationRequest(t, env, http.MethodDelete, base+id, token, nil).StatusCode)
		assert.Equal(t, http.StatusNotFound, destinationRequest(t, env, http.MethodPost, base+id+"/test", token, nil).StatusCode)
	}
	require.Len(t, destinationStatus(t, env, destinationBase, owner).Destinations, 1)
}
func TestNotificationDestinationRejectsInvalidSecrets(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	for _, dsn := range []string{"", "https://example.com/secret", "slack://secret@default?channel=C123", "slack://default?webhook_url=http://localhost/secret", "telegram://secret@default"} {
		resp := destinationRequest(t, env, http.MethodPost, destinationBase, token, api.SaveNotificationDestinationRequest{Name: "Alerts", DSN: dsn})
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		raw, err := io.ReadAll(resp.Body)
		require.NoError(t, err)
		assert.NotContains(t, string(raw), "secret")
	}
	assert.Empty(t, destinationStatus(t, env, destinationBase, token).Destinations)
}
