package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// encryptEnvironmentSecret replaces an environment's stored client_secret with an
// AES-GCM encrypted value so that newShopwareClientFromCredentials can decrypt it.
// SeedEnvironment stores a plaintext secret, which the Shopware-calling handlers
// cannot decrypt, so tests that actually reach Shopware must call this.
func encryptEnvironmentSecret(t *testing.T, env *testutil.TestEnv, environmentID int, secret string) {
	t.Helper()
	encrypted, err := crypto.Encrypt(secret, env.Cfg.AppSecret)
	require.NoError(t, err)
	_, err = env.Pool.Exec(t.Context(),
		`UPDATE environment SET client_secret = $1 WHERE id = $2`, encrypted, environmentID)
	require.NoError(t, err)
}

func TestGetOrganizationEnvironments(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	env.SeedEnvironment(t, "org-1", shopID, "Environment A", "https://a.example.com")
	env.SeedEnvironment(t, "org-1", shopID, "Environment B", "https://b.example.com")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-1/environments", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environments []json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environments))
	assert.Len(t, environments, 2)
}

func TestGetEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.Equal(t, "My Environment", environment.Name)
	assert.Equal(t, "https://env.example.com", environment.Url)
	assert.Equal(t, "6.5.0.0", environment.ShopwareVersion)
	assert.Equal(t, "green", environment.Status)
	assert.Equal(t, "Test Org", environment.OrganizationName)
	assert.NotNil(t, environment.Extensions)
	assert.NotNil(t, environment.ScheduledTasks)
	assert.NotNil(t, environment.Queues)
	assert.NotNil(t, environment.Checks)
}

func TestGetEnvironment_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	// Create another user's org and environment
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "To Delete", "https://del.example.com")

	req := testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify it's gone
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Environment is deleted, so org membership check on a non-existent environment returns not found
	assert.True(t, resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden)
}

func TestGetEnvironmentSubscription(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Subscription status is exposed on the environment detail payload, and a
	// freshly seeded environment is not subscribed.
	assert.False(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)
}

func TestUpdateEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "Old Name", "https://old.example.com")

	newName := "New Name"
	body, _ := json.Marshal(api.UpdateEnvironmentRequest{
		Name:   &newName,
		ShopId: shopID,
	})

	req := testutil.NewRequest(t, "PATCH", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify the name changed by fetching the environment
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.Equal(t, "New Name", environment.Name)
}

func TestSubscribeAndUnsubscribeEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Initially not subscribed
	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var result map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])

	// Subscribe
	req = testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify subscribed
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result["subscribed"])

	// The environment detail payload reflects the subscription too, so the
	// frontend doesn't need a separate request.
	assert.True(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)

	// Unsubscribe
	req = testutil.NewRequest(t, "DELETE", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify unsubscribed
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d/subscribe", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.False(t, result["subscribed"])

	assert.False(t, getEnvironmentDetail(t, env.Server.URL, token, environmentID).Subscribed)
}

// getEnvironmentDetail fetches the full environment detail payload for the given
// environment as the authenticated user.
func getEnvironmentDetail(t *testing.T, serverURL, token string, environmentID int) api.EnvironmentDetail {
	t.Helper()

	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", serverURL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	return environment
}

func TestUpdateSitespeedSettings(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	urls := []string{"https://env.example.com/", "https://env.example.com/products"}
	body, _ := json.Marshal(api.SitespeedSettingsRequest{
		Enabled: true,
		Urls:    &urls,
	})

	req := testutil.NewRequest(t, "PUT", fmt.Sprintf("%s/api/environments/%d/sitespeed-settings", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify settings by fetching the environment
	req = testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/environments/%d", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var environment api.EnvironmentDetail
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&environment))
	assert.True(t, environment.SitespeedEnabled)
	require.NotNil(t, environment.SitespeedUrls)
	assert.Len(t, *environment.SitespeedUrls, 2)
	assert.Contains(t, *environment.SitespeedUrls, "https://env.example.com/")
	assert.Contains(t, *environment.SitespeedUrls, "https://env.example.com/products")
}

func TestGetAccountExtensions_WithData(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	// Seed extensions
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO environment_extension (environment_id, name, label, active, version, latest_version, installed)
		VALUES ($1, 'SwagPayPal', 'PayPal', true, '5.0.0', '5.1.0', true),
		       ($1, 'SwagCmsExtensions', 'CMS Extensions', true, '3.2.0', '3.2.0', true)
	`, environmentID)
	require.NoError(t, err)

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/account/extensions", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var extensions []api.AccountExtension
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&extensions))
	require.Len(t, extensions, 2)

	// Find the PayPal extension
	var paypal *api.AccountExtension
	for i := range extensions {
		if extensions[i].Name == "SwagPayPal" {
			paypal = &extensions[i]
			break
		}
	}
	require.NotNil(t, paypal, "SwagPayPal extension should be present")
	assert.Equal(t, "PayPal", paypal.Label)
	assert.Equal(t, "5.0.0", paypal.Version)
	assert.Equal(t, "5.1.0", paypal.LatestVersion)
	assert.True(t, paypal.Active)
	assert.True(t, paypal.Installed)
	require.Len(t, paypal.Environments, 1)
	assert.Equal(t, environmentID, paypal.Environments[0].EnvironmentId)
	assert.Equal(t, "My Environment", paypal.Environments[0].EnvironmentName)
}

func TestRefreshEnvironment(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Only enqueues a job on the in-memory bus; does not call Shopware.
	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestRefreshEnvironment_WithSitespeed(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	sitespeed := true
	body, _ := json.Marshal(api.RefreshEnvironmentJSONRequestBody{Sitespeed: &sitespeed})

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusAccepted, resp.StatusCode)
}

func TestRefreshEnvironment_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRefreshEnvironment_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", "https://env.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/refresh", env.Server.URL, environmentID), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestClearEnvironmentCache(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	var cacheCleared bool
	mux.HandleFunc("/api/_action/cache", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		cacheCleared = true
		w.WriteHeader(http.StatusNoContent)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, cacheCleared, "expected DELETE /_action/cache to be called on Shopware")
}

func TestClearEnvironmentCache_ShopwareError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	mux.HandleFunc("/api/_action/cache", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Shopware failure is surfaced as a bad gateway.
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestClearEnvironmentCache_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/clear-cache", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestRescheduleTask(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	var patched, runTriggered bool
	mux.HandleFunc("/api/scheduled-task/task-123", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "PATCH", r.Method)
		patched = true
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/api/_action/scheduled-task/run", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		runTriggered = true
		w.WriteHeader(http.StatusOK)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
	assert.True(t, patched, "expected PATCH /scheduled-task/task-123 to be called")
	assert.True(t, runTriggered, "expected POST /_action/scheduled-task/run to be called")
}

func TestRescheduleTask_ShopwareError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})
	mux.HandleFunc("/api/scheduled-task/task-123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})
	shopware := httptest.NewServer(mux)
	defer shopware.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")
	environmentID := env.SeedEnvironment(t, "org-1", shopID, "My Environment", shopware.URL)
	encryptEnvironmentSecret(t, env, environmentID, "test-secret")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestRescheduleTask_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")
	environmentID := env.SeedEnvironment(t, "org-2", shopID, "Other Environment", "https://other.example.com")

	req := testutil.NewRequest(t, "POST", fmt.Sprintf("%s/api/environments/%d/tasks/task-123/reschedule", env.Server.URL, environmentID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
