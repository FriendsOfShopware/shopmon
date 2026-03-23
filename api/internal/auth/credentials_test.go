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

	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))

	user := result["user"].(map[string]interface{})
	assert.Equal(t, "New User", user["name"])
	assert.Equal(t, "new@example.com", user["email"])
	assert.NotEmpty(t, user["id"])

	// Should have a token in the response
	token, ok := result["token"].(string)
	require.True(t, ok, "response must contain a token")
	assert.NotEmpty(t, token)
}

func TestSignUpEmail_DuplicateEmail(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "dupe@example.com", "password": "password123!", "name": "User 1",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Try again with same email
	body, _ = json.Marshal(map[string]string{
		"email": "dupe@example.com", "password": "password456!", "name": "User 2",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestSignUpEmail_ShortPassword(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "test@example.com", "password": "short", "name": "Test",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignUpEmail_MissingFields(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{"email": "test@example.com"})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignInEmail(t *testing.T) {
	env := testutil.Setup(t)

	// First register
	body, _ := json.Marshal(map[string]string{
		"email": "login@example.com", "password": "password123!", "name": "Login User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Now login
	body, _ = json.Marshal(map[string]string{
		"email": "login@example.com", "password": "password123!",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	user := result["user"].(map[string]interface{})
	assert.Equal(t, "Login User", user["name"])
	assert.Equal(t, "login@example.com", user["email"])

	// Should have a token in the response
	token, ok := result["token"].(string)
	require.True(t, ok, "response must contain a token")
	assert.NotEmpty(t, token)
}

func TestSignInEmail_WrongPassword(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "wrong@example.com", "password": "password123!", "name": "User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	body, _ = json.Marshal(map[string]string{
		"email": "wrong@example.com", "password": "wrongpassword!",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSignInEmail_NonexistentUser(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"email": "nobody@example.com", "password": "password123!",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestGetSession(t *testing.T) {
	env := testutil.Setup(t)

	// Register to get a session
	body, _ := json.Marshal(map[string]string{
		"email": "session@example.com", "password": "password123!", "name": "Session User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	var signUpResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signUpResult))
	resp.Body.Close()
	token := signUpResult["token"].(string)
	require.NotEmpty(t, token)

	// Get session using the Bearer token
	req, _ := http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	user := result["user"].(map[string]interface{})
	assert.Equal(t, "Session User", user["name"])
	assert.Equal(t, "session@example.com", user["email"])
	assert.NotNil(t, result["session"])
}

func TestGetSession_NoToken(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(env.Server.URL + "/api/auth/session")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestSignOut(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "signout@example.com", "password": "password123!", "name": "Signout User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)

	var signUpResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&signUpResult))
	resp.Body.Close()
	token := signUpResult["token"].(string)
	require.NotEmpty(t, token)

	// Sign out
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/auth/sign-out", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Session should be invalid now
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestVerifyEmail(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "verify@example.com", "password": "password123!", "name": "Verify User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Get the verification token from DB
	var verificationToken string
	err = env.Pool.QueryRow(t.Context(),
		`SELECT value FROM verification WHERE identifier = 'verify@example.com'`).Scan(&verificationToken)
	require.NoError(t, err)

	// Verify email
	resp, err = http.Get(env.Server.URL + "/api/auth/verify-email?token=" + verificationToken)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Check user is now verified
	var verified bool
	err = env.Pool.QueryRow(t.Context(),
		`SELECT email_verified FROM "user" WHERE email = 'verify@example.com'`).Scan(&verified)
	require.NoError(t, err)
	assert.True(t, verified)
}

func TestVerifyEmail_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Get(env.Server.URL + "/api/auth/verify-email?token=invalid-token")
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestForgetPassword(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "forgot@example.com", "password": "password123!", "name": "Forgot User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Request password reset
	body, _ = json.Marshal(map[string]string{"email": "forgot@example.com"})
	resp, err = http.Post(env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
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
	resp, err := http.Post(env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestResetPassword(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "oldpassword1!", "name": "Reset User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Request reset
	body, _ = json.Marshal(map[string]string{"email": "reset@example.com"})
	resp, err = http.Post(env.Server.URL+"/api/auth/forget-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

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
	resp, err = http.Post(env.Server.URL+"/api/auth/reset-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login with new password should work
	body, _ = json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "newpassword1!",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	// Login with old password should fail
	body, _ = json.Marshal(map[string]string{
		"email": "reset@example.com", "password": "oldpassword1!",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestResetPassword_InvalidToken(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"token": "invalid", "newPassword": "newpassword1!",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/reset-password", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestSignInEmail_BannedUser(t *testing.T) {
	env := testutil.Setup(t)

	// Register
	body, _ := json.Marshal(map[string]string{
		"email": "banned@example.com", "password": "password123!", "name": "Banned User",
	})
	resp, err := http.Post(env.Server.URL+"/api/auth/sign-up/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	resp.Body.Close()

	// Ban the user directly in DB
	_, err = env.Pool.Exec(t.Context(),
		`UPDATE "user" SET banned = true, ban_reason = 'test ban' WHERE email = 'banned@example.com'`)
	require.NoError(t, err)

	// Try to login
	body, _ = json.Marshal(map[string]string{
		"email": "banned@example.com", "password": "password123!",
	})
	resp, err = http.Post(env.Server.URL+"/api/auth/sign-in/email", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
