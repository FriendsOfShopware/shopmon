package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSignUpEmail(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email":    "new@example.com",
		"password": "password123!",
		"name":     "New User",
	})

	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	assert.Equal(t, "ok", result["status"])
	assert.Empty(t, result["token"])
	var sessions int
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM session`).Scan(&sessions))
	assert.Zero(t, sessions)

	body, err = json.Marshal(map[string]string{"email": "new@example.com", "password": "password123!"})
	require.NoError(t, err)
	login, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = login.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, login.StatusCode)
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM session`).Scan(&sessions))
	assert.Zero(t, sessions)
	verifyEmail(t, env, "new@example.com")
	assert.NotEmpty(t, signIn(t, env, "new@example.com", "password123!"))
}

func TestSignUpEmail_DuplicateEmail(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "dupe@example.com", "password": "password123!", "name": "User 1",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Try again with same email
	body, _ = json.Marshal(map[string]string{
		"email": "dupe@example.com", "password": "password456!", "name": "User 2",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestSignUpEmail_ShortPassword(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "test@example.com", "password": "short", "name": "Test",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignUpEmail_MissingFields(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignInEmail(t *testing.T) {
	env := testutil.Setup(t)

	// First register
	body, _ := json.Marshal(map[string]string{
		"email": "login@example.com", "password": "password123!", "name": "Login User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	verifyEmail(t, env, "login@example.com")

	// Now login
	body, _ = json.Marshal(map[string]string{
		"email": "login@example.com", "password": "password123!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	user := result["user"].(map[string]any)
	assert.Equal(t, "Login User", user["name"])
	assert.Equal(t, "login@example.com", user["email"])

	// Should have a token in the response
	token, ok := result["token"].(string)
	require.True(t, ok, "response must contain a token")
	assert.NotEmpty(t, token)
}

func TestSignInEmail_Ban(t *testing.T) {
	tests := []struct {
		name       string
		email      string
		banSQL     string
		wantStatus int
	}{
		{
			name:       "active ban without expiry blocks login",
			email:      "active-ban@example.com",
			banSQL:     `UPDATE "user" SET banned = true, ban_reason = 'test ban' WHERE email = $1`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "active ban with future expiry blocks login",
			email:      "future-ban@example.com",
			banSQL:     `UPDATE "user" SET banned = true, ban_reason = 'test ban', ban_expires = NOW() + INTERVAL '1 hour' WHERE email = $1`,
			wantStatus: http.StatusForbidden,
		},
		{
			name:       "expired ban allows login",
			email:      "expired-ban@example.com",
			banSQL:     `UPDATE "user" SET banned = true, ban_reason = 'test ban', ban_expires = NOW() - INTERVAL '1 hour' WHERE email = $1`,
			wantStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			env := testutil.Setup(t)

			body, _ := json.Marshal(map[string]string{
				"email": tt.email, "password": "password123!", "name": "Ban User",
			})
			resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
			require.NoError(t, err)
			_ = resp.Body.Close()

			verifyEmail(t, env, tt.email)

			_, err = env.Pool.Exec(t.Context(), tt.banSQL, tt.email)
			require.NoError(t, err)

			body, _ = json.Marshal(map[string]string{
				"email": tt.email, "password": "password123!",
			})
			resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			assert.Equal(t, tt.wantStatus, resp.StatusCode)
		})
	}
}

func TestSignInEmail_WrongPassword(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "wrong@example.com", "password": "password123!", "name": "User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	body, _ = json.Marshal(map[string]string{
		"email": "wrong@example.com", "password": "wrongpassword!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSignInEmail_NonexistentUser(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "nobody@example.com", "password": "password123!",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSession(t *testing.T) {
	env := testutil.Setup(t)

	// Register to get a session
	body, _ := json.Marshal(map[string]string{
		"email": "session@example.com", "password": "password123!", "name": "Session User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	_ = resp.Body.Close()
	verifyEmail(t, env, "session@example.com")
	token := signIn(t, env, "session@example.com", "password123!")
	require.NotEmpty(t, token)

	// Get session using the Bearer token
	req := testutil.NewRequest(t, "GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	user := result["user"].(map[string]any)
	assert.Equal(t, "Session User", user["name"])
	assert.Equal(t, "session@example.com", user["email"])
	assert.NotNil(t, result["session"])
}

func TestGetSession_NoToken(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/session")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSession_ExpiredBan(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "expired-session-ban@example.com", "password": "password123!", "name": "Expired Ban User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	_ = resp.Body.Close()
	verifyEmail(t, env, "expired-session-ban@example.com")
	token := signIn(t, env, "expired-session-ban@example.com", "password123!")
	require.NotEmpty(t, token)

	_, err = env.Pool.Exec(t.Context(), `
		UPDATE "user"
		SET banned = true, ban_reason = 'expired test ban', ban_expires = NOW() - INTERVAL '1 hour'
		WHERE email = $1
	`, "expired-session-ban@example.com")
	require.NoError(t, err)

	req := testutil.NewRequest(t, http.MethodGet, env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestSignOut(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "signout@example.com", "password": "password123!", "name": "Signout User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	_ = resp.Body.Close()
	verifyEmail(t, env, "signout@example.com")
	token := signIn(t, env, "signout@example.com", "password123!")
	require.NotEmpty(t, token)

	// Sign out
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/sign-out", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Session should be invalid now
	req = testutil.NewRequest(t, "GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestVerifyEmail(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "verify@example.com", "password": "password123!", "name": "Verify User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	// Get the verification token from DB
	var verificationToken string
	err = env.Pool.QueryRow(t.Context(),
		`SELECT value FROM verification WHERE identifier LIKE 'verify-email:%:verify@example.com'`).Scan(&verificationToken)
	require.NoError(t, err)

	// Verify email
	resp, err = testutil.Get(t, env.Server.URL+"/api/auth/verify-email?token="+verificationToken)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check user is now verified
	var verified bool
	err = env.Pool.QueryRow(t.Context(),
		`SELECT email_verified FROM "user" WHERE email = 'verify@example.com'`).Scan(&verified)
	require.NoError(t, err)
	assert.True(t, verified)
	// Confirmation links are single use.
	replay, err := testutil.Get(t, env.Server.URL+"/api/auth/verify-email?token="+verificationToken)
	require.NoError(t, err)
	defer func() { _ = replay.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, replay.StatusCode)

}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/verify-email?token=invalid-token")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestForgetPassword(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "forgot@example.com", "password": "password123!", "name": "Forgot User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	// Request password reset
	body, _ = json.Marshal(map[string]string{"email": "forgot@example.com"})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Verify a token was created (might have email verification token too)
	var count int
	err = env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM verification`).Scan(&count)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)
}

func TestForgetPassword_NonexistentEmail(t *testing.T) {
	env := testutil.Setup(t)

	// Should return 200 even for non-existent email (anti-enumeration)
	body, _ := json.Marshal(map[string]string{"email": "nobody@example.com"})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResetPassword(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "oldpassword1!", "name": "Reset User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	verifyEmail(t, env, "reset@example.com")

	// Request reset
	body, _ = json.Marshal(map[string]string{"email": "reset@example.com"})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	// Get the reset token (forget-password uses user.ID as identifier)
	var resetToken string
	err = env.Pool.QueryRow(t.Context(),
		`SELECT v.value FROM verification v
		 JOIN "user" u ON u.id = v.identifier
		 WHERE u.email = 'reset@example.com'
		 ORDER BY v.created_at DESC LIMIT 1`).Scan(&resetToken)
	require.NoError(t, err)

	// Reset password
	body, _ = json.Marshal(map[string]string{
		"token": resetToken, "newPassword": "newpassword1!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/reset-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login with new password should work
	body, _ = json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "newpassword1!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login with old password should fail
	body, _ = json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "oldpassword1!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestResetPassword_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"token": "invalid", "newPassword": "newpassword1!",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/reset-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignInEmail_BannedUser(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "banned@example.com", "password": "password123!", "name": "Banned User",
	})
	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()

	// Ban the user directly in DB
	_, err = env.Pool.Exec(t.Context(),
		`UPDATE "user" SET banned = true, ban_reason = 'test ban' WHERE email = 'banned@example.com'`)
	require.NoError(t, err)

	// Try to login
	body, _ = json.Marshal(map[string]string{
		"email": "banned@example.com", "password": "password123!",
	})
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestChangeEmail_RequiresNewMailboxVerification(t *testing.T) {
	env := testutil.Setup(t)
	token := signUp(t, env, "old@example.com", "password123!", "User")
	secondToken := signIn(t, env, "old@example.com", "password123!")

	// An outstanding link for the old address must not verify the new address.
	_, err := env.Pool.Exec(t.Context(), `INSERT INTO verification (id, identifier, value, expires_at, created_at, updated_at)
		SELECT 'old-link', 'verify-email:' || id || ':' || email, 'old-token', NOW() + INTERVAL '1 hour', NOW(), NOW()
		FROM "user" WHERE email = 'old@example.com'`)
	require.NoError(t, err)

	resp := authRequest(t, http.MethodPost, env.Server.URL+"/api/auth/change-email", token,
		map[string]string{"newEmail": "new@example.com", "currentPassword": "password123!"})
	_ = resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	sent, ok := env.Mail.Last()
	require.True(t, ok)
	assert.Contains(t, string(sent.Bytes()), "To: new@example.com")

	var verified bool
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT email_verified FROM "user" WHERE email = 'new@example.com'`).Scan(&verified))
	assert.False(t, verified)
	for _, oldToken := range []string{token, secondToken} {
		resp = authRequest(t, http.MethodGet, env.Server.URL+"/api/auth/session", oldToken, nil)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	}
	resp, err = testutil.Get(t, env.Server.URL+"/api/auth/verify-email?token=old-token")
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	body, err := json.Marshal(map[string]string{"email": "new@example.com", "password": "password123!"})
	require.NoError(t, err)
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	verifyEmail(t, env, "new@example.com")
	assert.NotEmpty(t, signIn(t, env, "new@example.com", "password123!"))
}

func TestGetSession_UnverifiedLegacySession(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "legacy", "Legacy", "legacy@example.com", "user")
	_, err := env.Pool.Exec(t.Context(), `UPDATE "user" SET email_verified = false WHERE id = 'legacy'`)
	require.NoError(t, err)
	resp := authRequest(t, http.MethodGet, env.Server.URL+"/api/auth/session", token, nil)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	_, err = env.Pool.Exec(t.Context(), `INSERT INTO verification (id, identifier, value, expires_at, created_at, updated_at)
		VALUES ('legacy-link', 'verify-email:legacy:legacy@example.com', 'legacy-confirmation', NOW() + INTERVAL '1 hour', NOW(), NOW())`)
	require.NoError(t, err)
	verifyEmail(t, env, "legacy@example.com")
	// Old sessions must not become usable when the owner confirms ownership.
	after := authRequest(t, http.MethodGet, env.Server.URL+"/api/auth/session", token, nil)
	defer func() { _ = after.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, after.StatusCode)

}
