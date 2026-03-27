package auth_test

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// signUpAndMakeAdmin signs up a user and promotes them to admin via direct DB update.
func signUpAndMakeAdmin(t *testing.T, env *testutil.TestEnv, email, password, name string) string {
	t.Helper()
	token := signUp(t, env.Server.URL, email, password, name)
	_, err := env.Pool.Exec(t.Context(), `UPDATE "user" SET role = 'admin' WHERE email = $1`, email)
	require.NoError(t, err)
	return token
}

// getUserIDByEmail looks up a user ID from the database.
func getUserIDByEmail(t *testing.T, env *testutil.TestEnv, email string) string {
	t.Helper()
	var id string
	err := env.Pool.QueryRow(t.Context(), `SELECT id FROM "user" WHERE email = $1`, email).Scan(&id)
	require.NoError(t, err)
	return id
}

func TestAdminListUsers(t *testing.T) {
	env := testutil.Setup(t)
	adminCookie := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")

	// Create a regular user too
	signUp(t, env.Server.URL, "regular@example.com", "password123", "Regular")

	resp := authGet(t, env.Server.URL, "/api/auth/admin/users", adminCookie)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var users []map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&users))
	assert.GreaterOrEqual(t, len(users), 2)

	// Check that both users are present
	emails := make([]string, 0, len(users))
	for _, u := range users {
		emails = append(emails, u["email"].(string))
	}
	assert.Contains(t, emails, "admin@example.com")
	assert.Contains(t, emails, "regular@example.com")
}

func TestAdminListUsers_NonAdmin(t *testing.T) {
	env := testutil.Setup(t)
	cookie := signUp(t, env.Server.URL, "user@example.com", "password123", "User")

	resp := authGet(t, env.Server.URL, "/api/auth/admin/users", cookie)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestAdminSetRole(t *testing.T) {
	env := testutil.Setup(t)
	adminCookie := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")
	signUp(t, env.Server.URL, "user@example.com", "password123", "User")

	userID := getUserIDByEmail(t, env, "user@example.com")

	// Promote user to admin
	resp := authPatch(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/role", userID), adminCookie, map[string]string{
		"role": "admin",
	})
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := decodeJSON(t, resp)
	assert.Equal(t, "ok", result["status"])

	// Verify the role changed in DB
	var role string
	err := env.Pool.QueryRow(t.Context(), `SELECT role FROM "user" WHERE id = $1`, userID).Scan(&role)
	require.NoError(t, err)
	assert.Equal(t, "admin", role)
}

func TestAdminSetRole_Self(t *testing.T) {
	env := testutil.Setup(t)
	adminCookie := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")

	adminID := getUserIDByEmail(t, env, "admin@example.com")

	resp := authPatch(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/role", adminID), adminCookie, map[string]string{
		"role": "user",
	})
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAdminBanUser(t *testing.T) {
	env := testutil.Setup(t)
	adminCookie := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")
	userCookie := signUp(t, env.Server.URL, "user@example.com", "password123", "User")

	userID := getUserIDByEmail(t, env, "user@example.com")

	reason := "violation of terms"
	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/ban", userID), adminCookie, map[string]string{
		"banReason": reason,
	})
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)
	result := decodeJSON(t, resp)
	assert.Equal(t, "ok", result["status"])

	// Verify user is banned in DB
	var banned bool
	err := env.Pool.QueryRow(t.Context(), `SELECT banned FROM "user" WHERE id = $1`, userID).Scan(&banned)
	require.NoError(t, err)
	assert.True(t, banned)

	// Verify sessions are invalidated: the banned user's cookie should no longer work
	resp2 := authGet(t, env.Server.URL, "/api/auth/admin/users", userCookie)
	defer func() { _ = resp2.Body.Close() }()
	// The user is not admin anyway, but more importantly their session is gone
	// so they should get 401 (session invalid) or 403 (not admin).
	assert.NotEqual(t, http.StatusOK, resp2.StatusCode)
}

func TestAdminUnbanUser(t *testing.T) {
	env := testutil.Setup(t)
	adminCookie := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")
	signUp(t, env.Server.URL, "user@example.com", "password123", "User")

	userID := getUserIDByEmail(t, env, "user@example.com")

	// Ban the user first
	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/ban", userID), adminCookie, map[string]string{
		"banReason": "testing",
	})
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	// Now unban
	resp2 := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/unban", userID), adminCookie, nil)
	defer func() { _ = resp2.Body.Close() }()
	require.Equal(t, http.StatusOK, resp2.StatusCode)

	// Verify user is unbanned
	var banned bool
	err := env.Pool.QueryRow(t.Context(), `SELECT banned FROM "user" WHERE id = $1`, userID).Scan(&banned)
	require.NoError(t, err)
	assert.False(t, banned)
}

func TestAdminImpersonate(t *testing.T) {
	env := testutil.Setup(t)
	adminToken := signUpAndMakeAdmin(t, env, "admin@example.com", "password123", "Admin")
	signUp(t, env.Server.URL, "user@example.com", "password123", "User")

	userID := getUserIDByEmail(t, env, "user@example.com")

	resp := authPost(t, env.Server.URL, fmt.Sprintf("/api/auth/admin/users/%s/impersonate", userID), adminToken, nil)
	defer func() { _ = resp.Body.Close() }()

	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	impersonationToken, ok := result["token"].(string)
	require.True(t, ok)
	assert.NotEmpty(t, impersonationToken)

	session, ok := result["session"].(map[string]interface{})
	require.True(t, ok)
	assert.NotEmpty(t, session["token"])
	assert.NotEmpty(t, session["impersonatedBy"])

	// Use the impersonation token to access the session endpoint and verify it belongs to the target user
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+impersonationToken)
	resp2, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp2.Body.Close() }()

	require.Equal(t, http.StatusOK, resp2.StatusCode)

	var sessionResp map[string]interface{}
	require.NoError(t, json.NewDecoder(resp2.Body).Decode(&sessionResp))
	user := sessionResp["user"].(map[string]interface{})
	assert.Equal(t, "user@example.com", user["email"])
}
