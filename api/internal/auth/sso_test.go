package auth_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOIDCProvider creates a mock OIDC identity provider for testing.
func mockOIDCProvider(t *testing.T) (*httptest.Server, *rsa.PrivateKey) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			baseURL := "http://" + r.Host
			json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                 baseURL,
				"authorization_endpoint": baseURL + "/authorize",
				"token_endpoint":         baseURL + "/token",
				"jwks_uri":               baseURL + "/jwks",
				"userinfo_endpoint":      baseURL + "/userinfo",
				"scopes_supported":       []string{"openid", "profile", "email"},
			})

		case "/token":
			// Exchange code for token
			code := r.FormValue("code")
			if code != "valid-code" {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
				return
			}

			// Create a JWT ID token
			token := jwt.NewWithClaims(jwt.SigningMethodRS256, jwt.MapClaims{
				"iss":   "http://" + r.Host,
				"sub":   "sso-user-123",
				"email": "ssouser@testcorp.com",
				"name":  "SSO User",
				"iat":   time.Now().Unix(),
				"exp":   time.Now().Add(time.Hour).Unix(),
			})
			idToken, _ := token.SignedString(privateKey)

			json.NewEncoder(w).Encode(map[string]string{
				"access_token": "mock-access-token",
				"id_token":     idToken,
				"token_type":   "Bearer",
			})

		case "/userinfo":
			json.NewEncoder(w).Encode(map[string]string{
				"sub":   "sso-user-123",
				"email": "ssouser@testcorp.com",
				"name":  "SSO User",
			})

		default:
			http.NotFound(w, r)
		}
	}))

	return server, privateKey
}

// seedSSOProvider creates an SSO provider in the DB for testing.
func seedSSOProvider(t *testing.T, env *testutil.TestEnv, providerID, domain, orgID, idpURL string) {
	t.Helper()
	oidcConfig := fmt.Sprintf(`{
		"authorizationEndpoint": "%s/authorize",
		"tokenEndpoint": "%s/token",
		"jwksEndpoint": "%s/jwks",
		"clientId": "test-client-id",
		"clientSecret": "test-client-secret"
	}`, idpURL, idpURL, idpURL)

	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO sso_provider (id, issuer, oidc_config, provider_id, organization_id, domain)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "sso-"+providerID, idpURL, oidcConfig, providerID, orgID, domain)
	require.NoError(t, err)
}

func TestSignInSSO(t *testing.T) {
	idp, _ := mockOIDCProvider(t)
	defer idp.Close()

	env := testutil.Setup(t)

	// Create org and SSO provider
	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&orgResult)
	resp.Body.Close()
	orgID := orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idp.URL)

	// Initiate SSO sign-in
	body, _ := json.Marshal(map[string]string{
		"email":       "user@testcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Contains(t, result["url"], "/authorize")
	assert.Contains(t, result["url"], "client_id=test-client-id")
	assert.Contains(t, result["url"], "response_type=code")
	assert.Contains(t, result["url"], "scope=openid")
}

func TestSignInSSO_UnknownDomain(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "user@unknown-domain.com",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSignInSSO_InvalidEmail(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSSOCallback(t *testing.T) {
	idp, _ := mockOIDCProvider(t)
	defer idp.Close()

	env := testutil.Setup(t)

	// Create org and SSO provider
	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&orgResult)
	resp.Body.Close()
	orgID := orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idp.URL)

	// Initiate SSO to get the state
	body, _ := json.Marshal(map[string]string{
		"email":       "ssouser@testcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	var initResult map[string]string
	json.NewDecoder(resp.Body).Decode(&initResult)
	resp.Body.Close()

	// Extract state from the authorization URL
	authURL, err := url.Parse(initResult["url"])
	require.NoError(t, err)
	state := authURL.Query().Get("state")
	require.NotEmpty(t, state)

	// Simulate the callback with the code and state
	callbackURL := fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state))

	// Don't follow redirects so we can check the response
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err = client.Get(callbackURL)
	require.NoError(t, err)
	defer resp.Body.Close()

	// Should redirect to the callback URL with a one-time code
	assert.Equal(t, http.StatusFound, resp.StatusCode)
	location := resp.Header.Get("Location")
	locationURL, err := url.Parse(location)
	require.NoError(t, err)
	assert.Equal(t, "localhost:3000", locationURL.Host)
	assert.Equal(t, "/dashboard", locationURL.Path)

	// Code should be in the redirect URL (not the token)
	code := locationURL.Query().Get("code")
	require.NotEmpty(t, code, "code must be in redirect URL query params")
	assert.Empty(t, locationURL.Query().Get("token"), "token must NOT be in redirect URL")

	// Exchange code for token
	exchangeBody, _ := json.Marshal(map[string]string{"code": code})
	resp, err = http.Post(env.Server.URL+"/api/auth/exchange-code", "application/json", bytes.NewReader(exchangeBody))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var exchangeResult map[string]string
	json.NewDecoder(resp.Body).Decode(&exchangeResult)
	ssoToken := exchangeResult["token"]
	require.NotEmpty(t, ssoToken)

	// Verify session works
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+ssoToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var sessionResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessionResult)
	user := sessionResult["user"].(map[string]interface{})
	assert.Equal(t, "ssouser@testcorp.com", user["email"])
	assert.Equal(t, "SSO User", user["name"])
}

func TestSSOCallback_OrgAutoProvisioning(t *testing.T) {
	idp, _ := mockOIDCProvider(t)
	defer idp.Close()

	env := testutil.Setup(t)

	// Create org and SSO provider
	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&orgResult)
	resp.Body.Close()
	orgID := orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idp.URL)

	// Initiate + callback SSO
	body, _ := json.Marshal(map[string]string{
		"email":       "ssouser@testcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, _ = http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	var initResult map[string]string
	json.NewDecoder(resp.Body).Decode(&initResult)
	resp.Body.Close()

	authURL, _ := url.Parse(initResult["url"])
	state := authURL.Query().Get("state")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, _ = client.Get(fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	resp.Body.Close()

	// Verify the SSO user was auto-added to the organization
	var memberCount int
	err := env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM member WHERE organization_id = $1`, orgID).Scan(&memberCount)
	require.NoError(t, err)
	assert.Equal(t, 2, memberCount) // owner + SSO user
}

func TestSSOCallback_InvalidState(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(fmt.Sprintf("%s/api/auth/sso/callback/some-provider?code=abc&state=invalid",
		env.Server.URL))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSSOCallback_InvalidCode(t *testing.T) {
	idp, _ := mockOIDCProvider(t)
	defer idp.Close()

	env := testutil.Setup(t)

	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&orgResult)
	resp.Body.Close()
	orgID := orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idp.URL)

	// Initiate SSO
	body, _ := json.Marshal(map[string]string{
		"email": "user@testcorp.com",
	})
	resp, _ = http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	var initResult map[string]string
	json.NewDecoder(resp.Body).Decode(&initResult)
	resp.Body.Close()

	authURL, _ := url.Parse(initResult["url"])
	state := authURL.Query().Get("state")

	// Use invalid code
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, err := client.Get(fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=invalid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestSSOCallback_ExistingUserLinksAccount(t *testing.T) {
	idp, _ := mockOIDCProvider(t)
	defer idp.Close()

	env := testutil.Setup(t)

	// Create org
	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&orgResult)
	resp.Body.Close()
	orgID := orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idp.URL)

	// Pre-create the user with password auth (same email as SSO will return)
	signUp(t, env.Server.URL, "ssouser@testcorp.com", "password123!", "Existing User")

	// Now do SSO login — should link to existing user, not create a new one
	body, _ := json.Marshal(map[string]string{
		"email":       "ssouser@testcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, _ = http.Post(env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	var initResult map[string]string
	json.NewDecoder(resp.Body).Decode(&initResult)
	resp.Body.Close()

	authURL, _ := url.Parse(initResult["url"])
	state := authURL.Query().Get("state")

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	resp, _ = client.Get(fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	resp.Body.Close()

	// Extract code from redirect URL and exchange for token
	location := resp.Header.Get("Location")
	locationURL, err := url.Parse(location)
	require.NoError(t, err)
	code := locationURL.Query().Get("code")
	require.NotEmpty(t, code)

	exchangeBody, _ := json.Marshal(map[string]string{"code": code})
	resp, _ = http.Post(env.Server.URL+"/api/auth/exchange-code", "application/json", bytes.NewReader(exchangeBody))
	var exchangeResult map[string]string
	json.NewDecoder(resp.Body).Decode(&exchangeResult)
	resp.Body.Close()
	ssoToken := exchangeResult["token"]
	require.NotEmpty(t, ssoToken)

	// Verify it's the same user
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+ssoToken)
	resp, _ = http.DefaultClient.Do(req)

	var sessionResult map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&sessionResult)
	resp.Body.Close()
	user := sessionResult["user"].(map[string]interface{})
	assert.Equal(t, "ssouser@testcorp.com", user["email"])

	// Should only be 1 user with that email
	var userCount int
	env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount)
	assert.Equal(t, 1, userCount)
}
