package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetSsoProviders_Empty(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "user-1")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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

	req := testutil.NewRequest(t, "PUT", env.Server.URL+"/api/organizations/org-1/sso-providers/provider-1", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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

	req := testutil.NewRequest(t, "DELETE", env.Server.URL+"/api/organizations/org-1/sso-providers/provider-1", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	// Verify deleted
	req = testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-1/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	var providers []api.SsoProvider
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&providers))
	assert.Empty(t, providers)
}

func TestGetSsoProviders_NotMember(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")
	env.SeedUser(t, "user-2", "Other User", "other@example.com", "user")
	env.SeedOrganization(t, "org-2", "Other Org", "other-org", "user-2")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/organizations/org-2/sso-providers", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestDiscoverSso uses a mock HTTP server to test OIDC discovery
func TestDiscoverSso(t *testing.T) {
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/.well-known/openid-configuration" {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                 "https://idp.example.com",
				"authorization_endpoint": "https://idp.example.com/authorize",
				"token_endpoint":         "https://idp.example.com/token",
				"jwks_uri":               "https://idp.example.com/jwks",
				"userinfo_endpoint":      "https://idp.example.com/userinfo",
				"scopes_supported":       []string{"openid", "profile", "email"},
			})
			return
		}
		http.NotFound(w, r)
	}))
	defer mockServer.Close()

	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	// Use the mock server URL (127.0.0.1) as the issuer; discovery against a
	// local address is allowed for dev workflows.
	issuerURL := mockServer.URL
	req := testutil.NewRequest(t, "GET", fmt.Sprintf("%s/api/sso/discover?issuer=%s", env.Server.URL, issuerURL), nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var discovery api.SsoDiscovery
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&discovery))
	assert.Equal(t, "https://idp.example.com", discovery.Issuer)
	assert.Equal(t, "https://idp.example.com/authorize", discovery.AuthorizationEndpoint)
	assert.Equal(t, "https://idp.example.com/token", discovery.TokenEndpoint)
	assert.Contains(t, discovery.Scopes, "openid")
}

// TestDiscoverSso_Unauthenticated ensures discovery requires authentication,
// reducing the SSRF attack surface to logged-in accounts.
func TestDiscoverSso_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/sso/discover?issuer=https://idp.example.com", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

// TestDiscoverSso_RejectsNonHTTPS ensures non-HTTPS, non-local issuers are
// rejected before any outbound request is made.
func TestDiscoverSso_RejectsNonHTTPS(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/sso/discover?issuer=http://idp.example.com", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// seedOrgMember adds a user to an org with the given role and returns a token.
func seedOrgMember(t *testing.T, env *testutil.TestEnv, memberID, orgID, role string) string {
	t.Helper()
	token := env.SeedUser(t, memberID, "Member "+memberID, memberID+"@example.com", "user")
	_, err := env.Pool.Exec(t.Context(),
		`INSERT INTO member (id, organization_id, user_id, role, created_at)
		 VALUES ($1, $2, $3, $4, NOW())`,
		"mem-"+memberID, orgID, memberID, role)
	require.NoError(t, err)
	return token
}

func TestUpdateSsoProvider_MemberDenied(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "owner-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "owner-1")
	memberToken := seedOrgMember(t, env, "member-1", "org-1", "member")

	body, _ := json.Marshal(api.UpdateSsoProviderRequest{
		Domain:                "example.com",
		Issuer:                "https://idp.example.com",
		ClientId:              "client-123",
		AuthorizationEndpoint: "https://idp.example.com/authorize",
		TokenEndpoint:         "https://idp.example.com/token",
		JwksEndpoint:          "https://idp.example.com/jwks",
	})

	req := testutil.NewRequest(t, "PUT", env.Server.URL+"/api/organizations/org-1/sso-providers/some-provider", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+memberToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestDeleteSsoProvider_MemberDenied(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "owner-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "owner-1")
	memberToken := seedOrgMember(t, env, "member-1", "org-1", "member")

	req := testutil.NewRequest(t, "DELETE", env.Server.URL+"/api/organizations/org-1/sso-providers/some-provider", nil)
	req.Header.Set("Authorization", "Bearer "+memberToken)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestUpdateSsoProvider_AdminAllowed(t *testing.T) {
	env := testutil.Setup(t)
	env.SeedUser(t, "owner-1", "Owner", "owner@example.com", "user")
	env.SeedOrganization(t, "org-1", "Test Org", "test-org", "owner-1")
	adminToken := seedOrgMember(t, env, "admin-1", "org-1", "admin")

	body, _ := json.Marshal(api.UpdateSsoProviderRequest{
		Domain:                "example.com",
		Issuer:                "https://idp.example.com",
		ClientId:              "client-123",
		AuthorizationEndpoint: "https://idp.example.com/authorize",
		TokenEndpoint:         "https://idp.example.com/token",
		JwksEndpoint:          "https://idp.example.com/jwks",
	})

	req := testutil.NewRequest(t, "PUT", env.Server.URL+"/api/organizations/org-1/sso-providers/some-provider", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+adminToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
