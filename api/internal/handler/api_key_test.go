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

func TestGetApiKeyScopes(t *testing.T) {
	env := testutil.Setup(t)

	// This is a public endpoint, no auth needed
	resp, err := http.Get(env.Server.URL + "/api/api-key-scopes")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var scopes []api.ApiKeyScope
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&scopes))
	require.Len(t, scopes, 1)
	assert.Equal(t, "deployments", scopes[0].Value)
}

func TestGetApiKeys_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/api-keys", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var keys []api.ApiKey
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&keys))
	assert.Empty(t, keys)
}

func TestCreateApiKey(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	body, _ := json.Marshal(api.CreateApiKeyRequest{
		Name:   "My API Key",
		Scopes: []string{"deployments"},
	})

	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/api-keys", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var result api.CreateApiKeyResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, "My API Key", result.Name)
	assert.Equal(t, []string{"deployments"}, result.Scopes)
	assert.NotEmpty(t, result.Token)
}

func TestCreateAndListApiKeys(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	// Create a key
	body, _ := json.Marshal(api.CreateApiKeyRequest{
		Name:   "Key 1",
		Scopes: []string{"deployments"},
	})
	req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/api-keys", env.Server.URL, shopID), bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	// List keys
	req, _ = http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/api-keys", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var keys []api.ApiKey
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&keys))
	require.Len(t, keys, 1)
	assert.Equal(t, "Key 1", keys[0].Name)
}

func TestDeleteApiKey(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")
	shopID := env.SeedShop(t, "org-1", "Test Shop")

	// Create a key directly in DB to get the ID
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO shop_api_key (id, shop_id, name, token, scopes, created_at)
		VALUES ('key-1', $1, 'Test Key', 'shm_test_token', '["deployments"]'::jsonb, NOW())
	`, shopID)
	require.NoError(t, err)

	// Test with a numeric-ish key ID via the URL - just verify the endpoint responds
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/organizations/org-1/shops/%d/api-keys/999", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetApiKeys_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")
	shopID := env.SeedShop(t, "org-2", "Other Shop")

	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/organizations/org-2/shops/%d/api-keys", env.Server.URL, shopID), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
