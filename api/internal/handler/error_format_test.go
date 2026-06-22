package handler_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assertJSONError checks that the response is the standard JSON error shape
// ({"message": "..."}) with a JSON content type and a non-empty message.
func assertJSONError(t *testing.T, resp *http.Response, wantStatus int) {
	t.Helper()
	assert.Equal(t, wantStatus, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body map[string]json.RawMessage
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	// The shape is exactly {"message": "..."}.
	require.Contains(t, body, "message")
	assert.Len(t, body, 1, "error body should only contain a message field")

	var msg string
	require.NoError(t, json.Unmarshal(body["message"], &msg))
	assert.NotEmpty(t, msg)
}

// TestUnmatchedAPIRouteReturnsJSON404 verifies that an unknown /api path returns
// a JSON 404 instead of falling through to the SPA / HTML.
func TestUnmatchedAPIRouteReturnsJSON404(t *testing.T) {
	env := testutil.Setup(t)

	req := testutil.NewRequest(t, http.MethodGet, env.Server.URL+"/api/this-route-does-not-exist", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assertJSONError(t, resp, http.StatusNotFound)
}

// TestWrongMethodReturnsJSON405 verifies that hitting a real route with an
// unsupported method returns a JSON 405.
func TestWrongMethodReturnsJSON405(t *testing.T) {
	env := testutil.Setup(t)

	// /api/health exists for GET; DELETE is not allowed.
	req := testutil.NewRequest(t, http.MethodDelete, env.Server.URL+"/api/health", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assertJSONError(t, resp, http.StatusMethodNotAllowed)
}

// TestInvalidPathParamReturnsJSON400 verifies that oapi-codegen's parameter
// binding failures surface as a JSON 400, not plain text. A non-numeric
// environment id cannot bind to the typed path parameter.
func TestInvalidPathParamReturnsJSON400(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req := testutil.NewRequest(t, http.MethodGet, env.Server.URL+"/api/environments/not-a-number", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assertJSONError(t, resp, http.StatusBadRequest)
}

// TestUnauthorizedReturnsJSON401 sanity-checks that the existing auth-failure
// path also uses the standard JSON shape (regression guard for the contract).
func TestUnauthorizedReturnsJSON401(t *testing.T) {
	env := testutil.Setup(t)

	req := testutil.NewRequest(t, http.MethodGet, env.Server.URL+"/api/account/me", nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	// Some endpoints may 401 or 403 without a token; both must be JSON.
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")
	assert.True(t, resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden,
		"expected 401/403, got %d", resp.StatusCode)
}
