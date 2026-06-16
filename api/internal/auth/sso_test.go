package auth_test

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockOIDC holds a configurable mock OIDC identity provider for testing.
type mockOIDC struct {
	server     *httptest.Server
	privateKey *rsa.PrivateKey
	// signingKey is the key used to sign the ID token; defaults to privateKey.
	// Override to simulate a forged token.
	signingKey *rsa.PrivateKey
	// overrideClaims mutates the default ID token claims before signing.
	overrideClaims func(claims jwt.MapClaims)
	// currentNonce is embedded as the ID token nonce claim. Tests set it to the
	// nonce extracted from the authorization URL.
	currentNonce string
	// userinfoSub overrides the sub returned by the /userinfo endpoint. Defaults
	// to "sso-user-123" (matching the ID token sub) when empty.
	userinfoSub string
	// userinfoEmailVerified, when set, adds an email_verified field to the
	// /userinfo response.
	userinfoEmailVerified *bool
}

// jwkFromPublicKey encodes an RSA public key as a JWK for the JWKS endpoint.
func jwkFromPublicKey(pub *rsa.PublicKey, kid string) map[string]interface{} {
	eBytes := big.NewInt(int64(pub.E)).Bytes()
	return map[string]interface{}{
		"kty": "RSA",
		"use": "sig",
		"alg": "RS256",
		"kid": kid,
		"n":   base64.RawURLEncoding.EncodeToString(pub.N.Bytes()),
		"e":   base64.RawURLEncoding.EncodeToString(eBytes),
	}
}

// mockOIDCProvider creates a configurable mock OIDC identity provider.
func mockOIDCProvider(t *testing.T) *mockOIDC {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	m := &mockOIDC{privateKey: privateKey}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/.well-known/openid-configuration":
			baseURL := "http://" + r.Host
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"issuer":                 baseURL,
				"authorization_endpoint": baseURL + "/authorize",
				"token_endpoint":         baseURL + "/token",
				"jwks_uri":               baseURL + "/jwks",
				"userinfo_endpoint":      baseURL + "/userinfo",
				"scopes_supported":       []string{"openid", "profile", "email"},
			})

		case "/jwks":
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"keys": []interface{}{jwkFromPublicKey(&privateKey.PublicKey, "test-key")},
			})

		case "/token":
			code := r.FormValue("code")
			if code != "valid-code" {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid_grant"})
				return
			}

			claims := jwt.MapClaims{
				"iss":            "http://" + r.Host,
				"sub":            "sso-user-123",
				"aud":            "test-client-id",
				"email":          "ssouser@testcorp.com",
				"email_verified": true,
				"name":           "SSO User",
				"nonce":          m.currentNonce,
				"iat":            time.Now().Unix(),
				"exp":            time.Now().Add(time.Hour).Unix(),
			}
			if m.overrideClaims != nil {
				m.overrideClaims(claims)
			}

			token := jwt.NewWithClaims(jwt.SigningMethodRS256, claims)
			token.Header["kid"] = "test-key"
			signingKey := m.signingKey
			if signingKey == nil {
				signingKey = privateKey
			}
			idToken, _ := token.SignedString(signingKey)

			_ = json.NewEncoder(w).Encode(map[string]string{
				"access_token": "mock-access-token",
				"id_token":     idToken,
				"token_type":   "Bearer",
			})

		case "/userinfo":
			uiSub := m.userinfoSub
			if uiSub == "" {
				uiSub = "sso-user-123"
			}
			ui := map[string]interface{}{
				"sub":   uiSub,
				"email": "ssouser@testcorp.com",
				"name":  "SSO User",
			}
			if m.userinfoEmailVerified != nil {
				ui["email_verified"] = *m.userinfoEmailVerified
			}
			_ = json.NewEncoder(w).Encode(ui)

		default:
			http.NotFound(w, r)
		}
	}))

	return m
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

// initiateSSO performs the sign-in step and returns the state and nonce from
// the authorization URL.
func initiateSSO(t *testing.T, serverURL, email string) (state, nonce string) {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"email":       email,
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, serverURL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	var initResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&initResult))
	_ = resp.Body.Close()

	authURL, err := url.Parse(initResult["url"])
	require.NoError(t, err)
	return authURL.Query().Get("state"), authURL.Query().Get("nonce")
}

// createOrgWithSSO sets up an org owner, an organization, and an SSO provider.
func createOrgWithSSO(t *testing.T, env *testutil.TestEnv, idpURL string) (orgID, ownerCookie string) {
	t.Helper()
	ownerCookie = signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orgResult))
	_ = resp.Body.Close()
	orgID = orgResult["id"].(string)

	seedSSOProvider(t, env, "testcorp-oidc", "testcorp.com", orgID, idpURL)
	return orgID, ownerCookie
}

func TestSignInSSO(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	body, _ := json.Marshal(map[string]string{
		"email":       "user@testcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Contains(t, result["url"], "/authorize")
	assert.Contains(t, result["url"], "client_id=test-client-id")
	assert.Contains(t, result["url"], "response_type=code")
	assert.Contains(t, result["url"], "scope=openid")
	assert.Contains(t, result["url"], "nonce=")
}

func TestSignInSSO_UnknownDomain(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "user@unknown-domain.com",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestSignInSSO_InvalidEmail(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "not-an-email",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSSOCallback(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	require.NotEmpty(t, state)
	require.NotEmpty(t, nonce)
	idp.currentNonce = nonce

	callbackURL := fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state))

	resp, err := testutil.Get(t, callbackURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var callbackResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&callbackResult))
	code := callbackResult["code"]
	require.NotEmpty(t, code, "code must be in JSON response")

	exchangeBody, _ := json.Marshal(map[string]string{"code": code})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/exchange-code", "application/json", bytes.NewReader(exchangeBody))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var exchangeResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&exchangeResult))
	ssoToken := exchangeResult["token"]
	require.NotEmpty(t, ssoToken)

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+ssoToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var sessionResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sessionResult))
	user := sessionResult["user"].(map[string]interface{})
	assert.Equal(t, "ssouser@testcorp.com", user["email"])
	assert.Equal(t, "SSO User", user["name"])
}

func TestSSOCallback_OrgAutoProvisioning(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	orgID, _ := createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, _ := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	_ = resp.Body.Close()

	var memberCount int
	err := env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM member WHERE organization_id = $1`, orgID).Scan(&memberCount)
	require.NoError(t, err)
	assert.Equal(t, 2, memberCount) // owner + SSO user
}

func TestSSOCallback_InvalidState(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/some-provider?code=abc&state=invalid",
		env.Server.URL))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSSOCallback_InvalidCode(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, _ := initiateSSO(t, env.Server.URL, "user@testcorp.com")

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=invalid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

func TestSSOCallback_ExistingUserLinksAccount(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	// Pre-create the user with password auth (same email as SSO will return)
	signUp(t, env.Server.URL, "ssouser@testcorp.com", "password123!", "Existing User")

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, _ := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))

	var callbackResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&callbackResult))
	_ = resp.Body.Close()
	code := callbackResult["code"]
	require.NotEmpty(t, code)

	exchangeBody, _ := json.Marshal(map[string]string{"code": code})
	resp, _ = testutil.Post(t, env.Server.URL+"/api/auth/exchange-code", "application/json", bytes.NewReader(exchangeBody))
	var exchangeResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&exchangeResult))
	_ = resp.Body.Close()
	ssoToken := exchangeResult["token"]
	require.NotEmpty(t, ssoToken)

	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+ssoToken)
	resp, _ = http.DefaultClient.Do(req)

	var sessionResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sessionResult))
	_ = resp.Body.Close()
	user := sessionResult["user"].(map[string]interface{})
	assert.Equal(t, "ssouser@testcorp.com", user["email"])

	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount))
	assert.Equal(t, 1, userCount)
}

// TestSSOCallback_ForgedToken ensures an ID token signed with a key that is not
// published in the provider's JWKS is rejected.
func TestSSOCallback_ForgedToken(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	// Sign the ID token with an attacker-controlled key not in the JWKS.
	forgedKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)
	idp.signingKey = forgedKey

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	// No user should have been created from the forged token.
	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount))
	assert.Equal(t, 0, userCount)
}

// TestSSOCallback_AudienceMismatch ensures a token whose aud does not match the
// configured client id is rejected.
func TestSSOCallback_AudienceMismatch(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	idp.overrideClaims = func(claims jwt.MapClaims) {
		claims["aud"] = "some-other-client"
	}

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

// TestSSOCallback_NonceMismatch ensures a token whose nonce does not match the
// one issued during sign-in is rejected (replay protection).
func TestSSOCallback_NonceMismatch(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, _ := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	// Deliberately embed a nonce that does not match the one issued.
	idp.currentNonce = "attacker-controlled-nonce"

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

// TestSSOCallback_ExpiredToken ensures an expired ID token is rejected.
func TestSSOCallback_ExpiredToken(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	idp.overrideClaims = func(claims jwt.MapClaims) {
		claims["exp"] = time.Now().Add(-time.Hour).Unix()
		claims["iat"] = time.Now().Add(-2 * time.Hour).Unix()
	}

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}

// TestSSOCallback_MissingSub ensures an ID token without a sub claim is
// rejected, so an account is never keyed on an empty subject.
func TestSSOCallback_MissingSub(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	idp.overrideClaims = func(claims jwt.MapClaims) {
		delete(claims, "sub")
	}

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount))
	assert.Equal(t, 0, userCount)
}

// TestSSOCallback_UserinfoSubMismatch ensures that when the email is sourced
// from the userinfo endpoint, a userinfo sub that differs from the ID token sub
// is rejected (prevents binding a different identity).
func TestSSOCallback_UserinfoSubMismatch(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	// Force the userinfo fallback by omitting email from the ID token, and have
	// userinfo return a different subject.
	idp.overrideClaims = func(claims jwt.MapClaims) {
		delete(claims, "email")
	}
	idp.userinfoSub = "a-different-subject"

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount))
	assert.Equal(t, 0, userCount)
}

// TestSSOCallback_UserinfoEmailFallback ensures the userinfo fallback works when
// the userinfo sub matches the ID token sub.
func TestSSOCallback_UserinfoEmailFallback(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	// Omit email from the ID token; userinfo supplies it with a matching sub.
	idp.overrideClaims = func(claims jwt.MapClaims) {
		delete(claims, "email")
	}

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestSSOCallback_UserinfoEmailUnverified ensures that when the email is sourced
// from the userinfo endpoint, an email_verified:false there is honored and the
// login is rejected (not silently treated as verified).
func TestSSOCallback_UserinfoEmailUnverified(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	// Force the userinfo fallback and have userinfo report the email unverified.
	idp.overrideClaims = func(claims jwt.MapClaims) {
		delete(claims, "email")
	}
	unverified := false
	idp.userinfoEmailVerified = &unverified

	env := testutil.Setup(t)
	createOrgWithSSO(t, env, idp.server.URL)

	state, nonce := initiateSSO(t, env.Server.URL, "ssouser@testcorp.com")
	idp.currentNonce = nonce

	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/testcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'ssouser@testcorp.com'`).Scan(&userCount))
	assert.Equal(t, 0, userCount)
}

// TestSSOCallback_RejectsInsecureStoredEndpoint ensures a provider row with an
// http:// non-loopback endpoint (e.g. predating endpoint validation or inserted
// out-of-band) is rejected at callback time before any server-side request is
// made — even though such a row could never be created through the API.
func TestSSOCallback_RejectsInsecureStoredEndpoint(t *testing.T) {
	idp := mockOIDCProvider(t)
	defer idp.server.Close()

	env := testutil.Setup(t)
	ownerCookie := signUp(t, env.Server.URL, "owner@testcorp.com", "password123!", "Owner")
	resp := authPost(t, env.Server.URL, "/api/auth/organizations", ownerCookie, map[string]string{
		"name": "Test Corp",
	})
	var orgResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&orgResult))
	_ = resp.Body.Close()
	orgID := orgResult["id"].(string)

	// Seed a provider whose token endpoint is plain HTTP against a non-loopback
	// host. The issuer points at the real mock IdP so sign-in (which only reads
	// the authorization endpoint) succeeds and we reach the callback.
	oidcConfig := fmt.Sprintf(`{
		"authorizationEndpoint": "%s/authorize",
		"tokenEndpoint": "http://evil.internal.example/token",
		"jwksEndpoint": "%s/jwks",
		"clientId": "test-client-id",
		"clientSecret": "test-client-secret"
	}`, idp.server.URL, idp.server.URL)
	_, err := env.Pool.Exec(t.Context(), `
		INSERT INTO sso_provider (id, issuer, oidc_config, provider_id, organization_id, domain)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, "sso-evilcorp-oidc", idp.server.URL, oidcConfig, "evilcorp-oidc", orgID, "evilcorp.com")
	require.NoError(t, err)

	body, _ := json.Marshal(map[string]string{
		"email":       "user@evilcorp.com",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/sso", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	var initResult map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&initResult))
	_ = resp.Body.Close()
	authURL, err := url.Parse(initResult["url"])
	require.NoError(t, err)
	state := authURL.Query().Get("state")

	resp, err = testutil.Get(t, fmt.Sprintf("%s/api/auth/sso/callback/evilcorp-oidc?code=valid-code&state=%s",
		env.Server.URL, url.QueryEscape(state)))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Rejected before any request to the insecure endpoint.
	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)
}
