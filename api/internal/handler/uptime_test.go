package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateUptimeSettings(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	url := "https://status.example.com/health"
	body, _ := json.Marshal(api.UptimeSettingsRequest{
		Enabled:           true,
		Url:               &url,
		IntervalSeconds:   60,
		ExpectedStatus:    200,
		FailureThreshold:  3,
		RecoveryThreshold: 2,
	})

	req := testutil.NewRequest(t, "PUT", fmt.Sprintf("%s/api/environments/%d/uptime-settings", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// The monitor should now be reported by the uptime read endpoint.
	view := getUptime(t, env.Server.URL, token, environmentID, "24h")
	assert.True(t, view.Settings.Enabled)
	assert.Equal(t, 60, view.Settings.IntervalSeconds)
	assert.Equal(t, 200, view.Settings.ExpectedStatus)
	require.NotNil(t, view.Settings.Url)
	assert.Equal(t, url, *view.Settings.Url)
}

func TestUpdateUptimeSettingsValidation(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Interval below the allowed minimum.
	body, _ := json.Marshal(api.UptimeSettingsRequest{
		Enabled:           true,
		IntervalSeconds:   5,
		ExpectedStatus:    0,
		FailureThreshold:  3,
		RecoveryThreshold: 2,
	})

	req := testutil.NewRequest(t, "PUT", fmt.Sprintf("%s/api/environments/%d/uptime-settings", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnprocessableEntity, resp.StatusCode)
}

func TestGetUptimeSettingsAuthorization(t *testing.T) {
	env := testutil.Setup(t)
	owner := env.SeedUser(t, "owner-1", "Owner", "owner@example.com", "user")
	stranger := env.SeedUser(t, "stranger-1", "Stranger", "stranger@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "owner-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Non-member cannot read uptime.
	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/uptime?range=24h", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+stranger)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	// Owner can.
	view := getUptime(t, env.Server.URL, owner, environmentID, "24h")
	assert.False(t, view.Settings.Enabled, "monitor starts disabled")
}

func TestGetUptimeSettingsNotFound(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/uptime", env.Server.URL, 999999), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetUptimeUnauthenticated(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")
	_ = token

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/uptime", env.Server.URL, environmentID), nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func getUptime(t *testing.T, serverURL, token string, environmentID int, rangeKey string) api.UptimeResponse {
	t.Helper()

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/uptime?range=%s", serverURL, environmentID, rangeKey), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var view api.UptimeResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	return view
}
