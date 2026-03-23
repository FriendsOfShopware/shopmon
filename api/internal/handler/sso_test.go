package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSsoProviders_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var providers []api.SsoProvider
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&providers))
	assert.Empty(t, providers)
}

func TestGetSsoProviders_WithData(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	// Seed an SSO provider
	oidcConfig := `{"authorizationEndpoint":"https://idp.example.com/authorize","tokenEndpoint":"https://idp.example.com/token","jwksEndpoint":"https://idp.example.com/jwks","clientId":"my-client"}`
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO sso_provider (id, issuer, oidc_config, provider_id, organization_id, domain)
		VALUES ('sso-1', 'https://idp.example.com', $1, 'provider-1', 'org-1', 'example.com')
	`, oidcConfig)
	require.NoError(t, err)

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var providers []api.SsoProvider
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&providers))
	require.Len(t, providers, 1)
	assert.Equal(t, "provider-1", providers[0].Id)
	assert.Equal(t, "https://idp.example.com", providers[0].Issuer)
	assert.Equal(t, "example.com", providers[0].Domain)
}

func TestUpdateSsoProvider(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	// Seed provider
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO sso_provider (id, issuer, provider_id, organization_id, domain)
		VALUES ('sso-1', 'https://old.example.com', 'provider-1', 'org-1', 'old.com')
	`)
	require.NoError(t, err)

	body, _ := json.Marshal(api.UpdateSsoProviderRequest{
		Domain:                "new.example.com",
		Issuer:                "https://new.example.com",
		ClientId:              "new-client-id",
		AuthorizationEndpoint: "https://new.example.com/authorize",
		TokenEndpoint:         "https://new.example.com/token",
		JwksEndpoint:          "https://new.example.com/jwks",
	})

	req, _ := http.NewRequest("PUT", env.Server.URL+"/api/organizations/org-1/sso-providers/provider-1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteSsoProvider(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO sso_provider (id, issuer, provider_id, organization_id, domain)
		VALUES ('sso-1', 'https://idp.example.com', 'provider-1', 'org-1', 'example.com')
	`)
	require.NoError(t, err)

	req, _ := http.NewRequest("DELETE", env.Server.URL+"/api/organizations/org-1/sso-providers/provider-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify deleted
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	var providers []api.SsoProvider
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&providers))
	assert.Empty(t, providers)
}

func TestGetSsoProviders_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")

	req, _ := http.NewRequest("GET", env.Server.URL+"/api/organizations/org-2/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestDiscoverSso uses a mock HTTP server to test OIDC discovery
func TestDiscoverSso(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                 "https://idp.example.com",
				"authorization_endpoint": "https://idp.example.com/authorize",
				"token_endpoint":         "https://idp.example.com/token",
				"jwks_uri":              "https://idp.example.com/jwks",
				"userinfo_endpoint":      "https://idp.example.com/userinfo",
				"scopes_supported":       []string{"openid", "profile", "email"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	env := testutil.Setup(t)

	// Use the mock server URL as the issuer
	issuerURL := mockServer.URL
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/sso/discover?issuer=%s", env.Server.URL, issuerURL), nil)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var discovery api.SsoDiscovery
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&discovery))
	assert.Equal(t, "https://idp.example.com", discovery.Issuer)
	assert.Equal(t, "https://idp.example.com/authorize", discovery.AuthorizationEndpoint)
	assert.Equal(t, "https://idp.example.com/token", discovery.TokenEndpoint)
	assert.Contains(t, discovery.Scopes, "openid")
}
