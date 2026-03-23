package auth_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/descope/virtualwebauthn"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPasskeyRegisterOptions(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.ChallengeKey)
	assert.NotEmpty(t, result.Options)
}

func TestPasskeyRegisterOptions_Unauthenticated(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Post(env.Server.URL+"/api/auth/passkey/register-options", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPasskeyLoginOptions(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := http.Post(env.Server.URL+"/api/auth/passkey/login-options", "application/json", nil)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result.ChallengeKey)
	assert.NotEmpty(t, result.Options)
}

func TestPasskeyFullRegistrationAndLogin(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Passkey User", "passkey@example.com", "user")

	rp := virtualwebauthn.RelyingParty{
		Name:   "Shopmon",
		ID:     "localhost",
		Origin: "http://localhost:3000",
	}
	authenticator := virtualwebauthn.NewAuthenticator()
	credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

	// === Step 1: Get registration options ===
	req, _ := http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var regOptions struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&regOptions))
	resp.Body.Close()
	require.NotEmpty(t, regOptions.ChallengeKey)

	// === Step 2: Create attestation response using virtual authenticator ===
	attOpts, err := virtualwebauthn.ParseAttestationOptions(string(regOptions.Options))
	require.NoError(t, err)

	attResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attOpts)

	// Merge challengeKey and name into the attestation response
	var attJSON map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(attResponse), &attJSON))
	attJSON["challengeKey"] = regOptions.ChallengeKey
	attJSON["name"] = "My Test Passkey"
	mergedBody, _ := json.Marshal(attJSON)

	// === Step 3: Complete registration ===
	req, _ = http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(mergedBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var regResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&regResult))
	assert.NotEmpty(t, regResult["id"])
	assert.Equal(t, "My Test Passkey", regResult["name"])

	// Verify passkey was stored in DB
	var passkeyCount int
	err = env.Pool.QueryRow(t.Context(),
		`SELECT COUNT(*) FROM passkey WHERE user_id = 'user-1'`).Scan(&passkeyCount)
	require.NoError(t, err)
	assert.Equal(t, 1, passkeyCount)

	// === Step 4: Now test login with the passkey ===
	authenticator.Options.UserHandle = []byte("user-1")
	authenticator.AddCredential(credential)

	// Get login options
	resp, err = http.Post(env.Server.URL+"/api/auth/passkey/login-options", "application/json", nil)
	require.NoError(t, err)

	var loginOptions struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginOptions))
	resp.Body.Close()
	require.NotEmpty(t, loginOptions.ChallengeKey)

	// Create assertion response
	assOpts, err := virtualwebauthn.ParseAssertionOptions(string(loginOptions.Options))
	require.NoError(t, err)

	assResponse := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *assOpts)

	// Merge challengeKey into assertion
	var assJSON map[string]interface{}
	require.NoError(t, json.Unmarshal([]byte(assResponse), &assJSON))
	assJSON["challengeKey"] = loginOptions.ChallengeKey
	loginBody, _ := json.Marshal(assJSON)

	// Complete login
	resp, err = http.Post(env.Server.URL+"/api/auth/passkey/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var loginResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginResult))
	user := loginResult["user"].(map[string]interface{})
	assert.Equal(t, "passkey@example.com", user["email"])
	assert.Equal(t, "Passkey User", user["name"])

	// Should have a token in the response
	loginToken, ok := loginResult["token"].(string)
	require.True(t, ok, "response must contain a token")
	assert.NotEmpty(t, loginToken)

	// Verify session works
	req, _ = http.NewRequest("GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+loginToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestPasskeyRegister_InvalidChallenge(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	body, _ := json.Marshal(map[string]string{
		"challengeKey": "invalid-key",
	})

	req, _ := http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPasskeyLogin_InvalidChallenge(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"challengeKey": "invalid-key",
	})

	resp, err := http.Post(env.Server.URL+"/api/auth/passkey/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPasskeyRegisterMultiple(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Multi Key User", "multi@example.com", "user")

	rp := virtualwebauthn.RelyingParty{
		Name:   "Shopmon",
		ID:     "localhost",
		Origin: "http://localhost:3000",
	}
	authenticator := virtualwebauthn.NewAuthenticator()

	// Register two passkeys
	for i, name := range []string{"Laptop Key", "Phone Key"} {
		_ = i
		cred := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

		// Get options
		req, _ := http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		var opts struct {
			Options      json.RawMessage `json:"options"`
			ChallengeKey string          `json:"challengeKey"`
		}
		json.NewDecoder(resp.Body).Decode(&opts)
		resp.Body.Close()

		// Create attestation
		attOpts, _ := virtualwebauthn.ParseAttestationOptions(string(opts.Options))
		attResp := virtualwebauthn.CreateAttestationResponse(rp, authenticator, cred, *attOpts)

		var attJSON map[string]interface{}
		json.Unmarshal([]byte(attResp), &attJSON)
		attJSON["challengeKey"] = opts.ChallengeKey
		attJSON["name"] = name
		body, _ := json.Marshal(attJSON)

		// Register
		req, _ = http.NewRequest("POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		resp.Body.Close()
	}

	// Verify two passkeys stored
	var count int
	env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM passkey WHERE user_id = 'user-1'`).Scan(&count)
	assert.Equal(t, 2, count)
}
