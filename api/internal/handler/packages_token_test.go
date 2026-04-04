package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetPackagesTokenConfiguration_NotConfigured(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/packages-token/configuration", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cfg api.PackagesTokenConfiguration
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&cfg))
	assert.False(t, cfg.Configured)
	assert.Nil(t, cfg.ComposerUrl)
}

func TestGetPackagesTokenConfiguration_Configured(t *testing.T) {
	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = "https://packages.example.com"
		cfg.PackagesAPIToken = "secret-token"
	})
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/packages-token/configuration", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var cfg2 api.PackagesTokenConfiguration
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&cfg2))
	assert.True(t, cfg2.Configured)
	require.NotNil(t, cfg2.ComposerUrl)
	assert.Equal(t, "https://packages.example.com/composer", *cfg2.ComposerUrl)
}

func TestGetPackagesTokens_NotConfigured(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/packages-tokens", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetPackagesTokens_WithMockAPI(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify auth header
		assert.Equal(t, "Bearer mock-packages-token", r.Header.Get("Authorization"))

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"id": 1, "source": "shopware", "lastSyncedAt": 1700000000},
			{"id": 2, "source": "custom", "lastSyncedAt": nil},
		})
	}))
	defer mockAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = mockAPI.URL
		cfg.PackagesAPIToken = "mock-packages-token"
	})

	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/packages-tokens", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var tokens []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&tokens))
	assert.Len(t, tokens, 2)
}

func TestCreatePackagesToken_WithMockAPI(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Equal(t, "Bearer mock-packages-token", r.Header.Get("Authorization"))

		// Verify request body was proxied
		var body map[string]string
		_ = json.NewDecoder(r.Body).Decode(&body)
		assert.Equal(t, "my-shopware-token", body["token"])

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           3,
			"source":       "shopware",
			"lastSyncedAt": nil,
		})
	}))
	defer mockAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = mockAPI.URL
		cfg.PackagesAPIToken = "mock-packages-token"
	})

	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	body, _ := json.Marshal(map[string]string{"token": "my-shopware-token"})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/packages-tokens", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestDeletePackagesToken_WithMockAPI(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "DELETE", r.Method)
		assert.Contains(t, r.URL.Path, "/tokens/42")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer mockAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = mockAPI.URL
		cfg.PackagesAPIToken = "mock-packages-token"
	})

	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/packages-tokens/42", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestSyncPackagesToken_WithMockAPI(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "POST", r.Method)
		assert.Contains(t, r.URL.Path, "/tokens/42/sync")
		w.WriteHeader(http.StatusOK)
	}))
	defer mockAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = mockAPI.URL
		cfg.PackagesAPIToken = "mock-packages-token"
	})

	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/packages-tokens/42/sync", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetPackagesTokens_NotMember(t *testing.T) {
	mockAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("should not reach the packages API")
	}))
	defer mockAPI.Close()

	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.PackagesAPIURL = mockAPI.URL
		cfg.PackagesAPIToken = "mock-packages-token"
	})

	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-2/shops/%d/packages-tokens", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
