package auth_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// withGithub configures the test environment with GitHub OAuth credentials.
func withGithub(cfg *config.Config) {
	cfg.GithubClientID = "test-github-client-id"
	cfg.GithubClientSecret = "test-github-client-secret"
}

// TestGithubSignIn covers SignInSocial when GitHub OAuth is configured: it must
// return the GitHub authorize URL (as JSON) carrying the client_id and a state,
// and persist that state so the callback can later validate it.
func TestGithubSignIn(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	body, _ := json.Marshal(map[string]string{
		"provider":    "github",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	authURL, err := url.Parse(result["url"])
	require.NoError(t, err)
	assert.Equal(t, "github.com", authURL.Host)
	assert.Equal(t, "/login/oauth/authorize", authURL.Path)
	assert.Equal(t, "test-github-client-id", authURL.Query().Get("client_id"))
	assert.Equal(t, "user:email", authURL.Query().Get("scope"))
	assert.NotEmpty(t, authURL.Query().Get("state"), "authorize URL must carry a state")
}

// TestGithubSignIn_NotConfigured ensures that when no GitHub client ID is set,
// social sign-in is disabled with a 404.
func TestGithubSignIn_NotConfigured(t *testing.T) {
	env := testutil.Setup(t) // no GitHub credentials

	body, _ := json.Marshal(map[string]string{
		"provider":    "github",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestGithubSignIn_UnsupportedProvider ensures a non-github provider is rejected.
func TestGithubSignIn_UnsupportedProvider(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	body, _ := json.Marshal(map[string]string{
		"provider":    "gitlab",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestGithubSignIn_InvalidCallbackURL ensures a callback URL pointing at a host
// other than the configured frontend is rejected.
func TestGithubSignIn_InvalidCallbackURL(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	body, _ := json.Marshal(map[string]string{
		"provider":    "github",
		"callbackURL": "http://evil.example.com/steal",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestGithubCallback_InvalidState ensures the callback rejects a state value that
// was never issued (or has expired) before contacting GitHub.
func TestGithubCallback_InvalidState(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/callback/github?code=some-code&state=never-issued")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestGithubCallback_MissingCode ensures the generated handler rejects a callback
// without the required "code" query parameter.
func TestGithubCallback_MissingCode(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/callback/github?state=some-state")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// TestGithubCallback_MissingState ensures the generated handler rejects a
// callback without the required "state" query parameter.
func TestGithubCallback_MissingState(t *testing.T) {
	env := testutil.Setup(t, withGithub)

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/callback/github?code=some-code")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// mockGithub is a configurable mock GitHub OAuth + API server. It serves the
// token-exchange endpoint and the user / user-emails API endpoints used by
// GithubCallback.
type mockGithub struct {
	server *httptest.Server
	// accessToken returned by the token endpoint. When empty, the token
	// response omits access_token (simulating an exchange failure).
	accessToken string
	// user is the JSON object returned by GET /user.
	user map[string]interface{}
	// emails is the JSON array returned by GET /user/emails (the fallback used
	// when the user has no public email).
	emails []map[string]interface{}
}

func newMockGithub(t *testing.T) *mockGithub {
	t.Helper()
	m := &mockGithub{accessToken: "gho_mock_access_token"}

	m.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login/oauth/access_token":
			// GitHub returns the token as application/x-www-form-urlencoded.
			vals := url.Values{}
			if m.accessToken != "" {
				vals.Set("access_token", m.accessToken)
				vals.Set("token_type", "bearer")
			} else {
				vals.Set("error", "bad_verification_code")
			}
			w.Header().Set("Content-Type", "application/x-www-form-urlencoded")
			_, _ = w.Write([]byte(vals.Encode()))

		case "/user":
			_ = json.NewEncoder(w).Encode(m.user)

		case "/user/emails":
			_ = json.NewEncoder(w).Encode(m.emails)

		default:
			http.NotFound(w, r)
		}
	}))
	return m
}

// setupGithubEnv points the package-level GitHub base URLs at the mock server
// (restoring them on cleanup) and returns a configured test environment.
func setupGithubEnv(t *testing.T, mock *mockGithub) *testutil.TestEnv {
	t.Helper()
	restore := auth.SetGithubBaseURLsForTest(mock.server.URL, mock.server.URL)
	t.Cleanup(restore)

	return testutil.Setup(t, func(cfg *config.Config) {
		cfg.GithubClientID = "test-github-client-id"
		cfg.GithubClientSecret = "test-github-client-secret"
	})
}

// initiateGithubSignIn calls the social sign-in endpoint and returns the state
// stored in the challenge store (extracted from the authorize URL).
func initiateGithubSignIn(t *testing.T, serverURL string) string {
	t.Helper()
	body, _ := json.Marshal(map[string]string{
		"provider":    "github",
		"callbackURL": "http://localhost:3000/dashboard",
	})
	resp, err := testutil.Post(t, serverURL+"/api/auth/sign-in/social", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	authURL, err := url.Parse(result["url"])
	require.NoError(t, err)
	state := authURL.Query().Get("state")
	require.NotEmpty(t, state)
	return state
}

func githubCallback(t *testing.T, serverURL, code, state string) *http.Response {
	t.Helper()
	resp, err := testutil.Get(t, fmt.Sprintf("%s/api/auth/callback/github?code=%s&state=%s",
		serverURL, url.QueryEscape(code), url.QueryEscape(state)))
	require.NoError(t, err)
	return resp
}

// TestGithubCallback_HappyPath drives the full OAuth callback: a real state is
// minted via SignInSocial, the mock token + user endpoints respond, and a new
// user plus a linked github account row are created.
func TestGithubCallback_HappyPath(t *testing.T) {
	mock := newMockGithub(t)
	defer mock.server.Close()
	mock.user = map[string]interface{}{
		"id":         12345,
		"login":      "octocat",
		"name":       "The Octocat",
		"email":      "octocat@github.com",
		"avatar_url": "https://avatars.example/octocat.png",
	}

	env := setupGithubEnv(t, mock)
	state := initiateGithubSignIn(t, env.Server.URL)

	resp := githubCallback(t, env.Server.URL, "valid-code", state)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["code"], "callback must return a one-time code")

	// A user was created with the GitHub-provided details.
	var name, email string
	var verified bool
	err := env.Pool.QueryRow(t.Context(),
		`SELECT name, email, email_verified FROM "user" WHERE email = 'octocat@github.com'`).
		Scan(&name, &email, &verified)
	require.NoError(t, err)
	assert.Equal(t, "The Octocat", name)
	assert.True(t, verified, "GitHub-verified email must be marked verified")

	// A github account row links to that user.
	var accountID, providerID string
	err = env.Pool.QueryRow(t.Context(),
		`SELECT a.account_id, a.provider_id FROM account a
		 JOIN "user" u ON u.id = a.user_id
		 WHERE u.email = 'octocat@github.com' AND a.provider_id = 'github'`).
		Scan(&accountID, &providerID)
	require.NoError(t, err)
	assert.Equal(t, "12345", accountID)
	assert.Equal(t, "github", providerID)
}

// TestGithubCallback_EmailFallback covers the case where GET /user returns no
// email and the primary+verified address is sourced from GET /user/emails.
func TestGithubCallback_EmailFallback(t *testing.T) {
	mock := newMockGithub(t)
	defer mock.server.Close()
	mock.user = map[string]interface{}{
		"id":         67890,
		"login":      "hubot",
		"name":       "Hubot",
		"email":      "", // no public email
		"avatar_url": "",
	}
	mock.emails = []map[string]interface{}{
		{"email": "noreply@github.com", "primary": false, "verified": true},
		{"email": "hubot@github.com", "primary": true, "verified": true},
	}

	env := setupGithubEnv(t, mock)
	state := initiateGithubSignIn(t, env.Server.URL)

	resp := githubCallback(t, env.Server.URL, "valid-code", state)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var count int
	err := env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'hubot@github.com'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "user must be created with the primary+verified fallback email")
}

// TestGithubCallback_NoAccessToken ensures a token response without an
// access_token yields a 502 and creates no user.
func TestGithubCallback_NoAccessToken(t *testing.T) {
	mock := newMockGithub(t)
	defer mock.server.Close()
	mock.accessToken = "" // token endpoint returns an error instead of a token

	env := setupGithubEnv(t, mock)
	state := initiateGithubSignIn(t, env.Server.URL)

	resp := githubCallback(t, env.Server.URL, "bad-code", state)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadGateway, resp.StatusCode)

	var count int
	err := env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM "user"`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestGithubCallback_LinksExistingUser ensures that when a user with the same
// email already exists, the github account is linked to that user instead of
// creating a duplicate.
func TestGithubCallback_LinksExistingUser(t *testing.T) {
	mock := newMockGithub(t)
	defer mock.server.Close()
	mock.user = map[string]interface{}{
		"id":         24680,
		"login":      "existing",
		"name":       "Existing User",
		"email":      "existing@github.com",
		"avatar_url": "",
	}

	env := setupGithubEnv(t, mock)

	// Pre-create a user with the same email via password sign-up.
	signUpBody, _ := json.Marshal(map[string]string{
		"email": "existing@github.com", "password": "password123!", "name": "Existing User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(signUpBody))
	require.NoError(t, err)
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var existingUserID string
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT id FROM "user" WHERE email = 'existing@github.com'`).Scan(&existingUserID))

	state := initiateGithubSignIn(t, env.Server.URL)
	resp = githubCallback(t, env.Server.URL, "valid-code", state)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Still exactly one user with that email.
	var userCount int
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM "user" WHERE email = 'existing@github.com'`).Scan(&userCount))
	assert.Equal(t, 1, userCount)

	// The github account links to the pre-existing user.
	var linkedUserID string
	require.NoError(t, env.Pool.QueryRow(t.Context(),
		`SELECT user_id FROM account WHERE provider_id = 'github' AND account_id = '24680'`).
		Scan(&linkedUserID))
	assert.Equal(t, existingUserID, linkedUserID)
}
