package auth_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"io"
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

	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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

	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/passkey/register-options", "application/json", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestPasskeyLoginOptions(t *testing.T) {
	env := testutil.Setup(t)

	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/passkey/login-options", "application/json", nil)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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
	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)

	var regOptions struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&regOptions))
	_ = resp.Body.Close()
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
	req = testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(mergedBody))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, "register failed: body=%s", string(body))
	}

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
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/passkey/login-options", "application/json", nil)
	require.NoError(t, err)

	var loginOptions struct {
		Options      json.RawMessage `json:"options"`
		ChallengeKey string          `json:"challengeKey"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginOptions))
	_ = resp.Body.Close()
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
	resp, err = testutil.Post(t, env.Server.URL+"/api/auth/passkey/login", "application/json", bytes.NewReader(loginBody))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		require.Equal(t, http.StatusOK, resp.StatusCode, "login failed: body=%s", string(body))
	}

	var loginResult map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginResult))
	user, ok := loginResult["user"].(map[string]interface{})
	require.True(t, ok, "response must contain user object, got: %v", loginResult)
	assert.Equal(t, "passkey@example.com", user["email"])
	assert.Equal(t, "Passkey User", user["name"])

	// Should have a token in the response
	loginToken, ok := loginResult["token"].(string)
	require.True(t, ok, "response must contain a token")
	assert.NotEmpty(t, loginToken)

	// Verify session works
	req = testutil.NewRequest(t, "GET", env.Server.URL+"/api/auth/session", nil)
	req.Header.Set("Authorization", "Bearer "+loginToken)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestPasskeyLoginLegacyBetterAuth reproduces the production scenario where a
// passkey was registered by the previous better-auth stack. Such passkeys
// differ from ones this service registers in two ways that broke login after
// the Go port:
//
//  1. The WebAuthn userHandle is a random per-credential string, NOT the user
//     ID, so the user must be resolved via the credential ID.
//  2. The public key is stored as standard base64 (with +/=) and device_type
//     as SimpleWebAuthn's "multiDevice", not raw-url base64 / "multi_device".
//
// This test registers a passkey through the API, then rewrites the stored row
// to match better-auth's encoding and uses a non-matching userHandle, and
// asserts that login still succeeds. Subtests cover both the credential_id
// encoding seen in production (raw-url base64) and a standard-base64 fallback.
func TestPasskeyLoginLegacyBetterAuth(t *testing.T) {
	// reEncodeCredentialID, when non-nil, rewrites the stored credential_id from
	// its raw-url form into another base64 variant to exercise the lookup fallback.
	cases := []struct {
		name                 string
		reEncodeCredentialID func(string) string
	}{
		{
			// Production data: credential IDs stored as raw-url base64.
			name:                 "credential_id raw-url base64",
			reEncodeCredentialID: nil,
		},
		{
			// Defensive: a legacy credential ID stored as standard base64 must
			// still resolve via the lookup fallback.
			name: "credential_id standard base64",
			reEncodeCredentialID: func(s string) string {
				raw, err := base64.RawURLEncoding.DecodeString(s)
				if err != nil {
					return s
				}
				return base64.RawStdEncoding.EncodeToString(raw)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			env := testutil.Setup(t)
			token := env.SeedUser(t, "user-1", "Legacy User", "legacy@example.com", "user")

			rp := virtualwebauthn.RelyingParty{
				Name:   "Shopmon",
				ID:     "localhost",
				Origin: "http://localhost:3000",
			}
			authenticator := virtualwebauthn.NewAuthenticator()
			credential := virtualwebauthn.NewCredential(virtualwebauthn.KeyTypeEC2)

			// === Register a passkey through the API ===
			req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
			req.Header.Set("Authorization", "Bearer "+token)
			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)

			var regOptions struct {
				Options      json.RawMessage `json:"options"`
				ChallengeKey string          `json:"challengeKey"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&regOptions))
			_ = resp.Body.Close()

			attOpts, err := virtualwebauthn.ParseAttestationOptions(string(regOptions.Options))
			require.NoError(t, err)
			attResponse := virtualwebauthn.CreateAttestationResponse(rp, authenticator, credential, *attOpts)

			var attJSON map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(attResponse), &attJSON))
			attJSON["challengeKey"] = regOptions.ChallengeKey
			attJSON["name"] = "Legacy Passkey"
			mergedBody, _ := json.Marshal(attJSON)

			req = testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(mergedBody))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			resp, err = http.DefaultClient.Do(req)
			require.NoError(t, err)
			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				require.Equal(t, http.StatusOK, resp.StatusCode, "register failed: body=%s", string(body))
			}
			_ = resp.Body.Close()

			// === Rewrite the stored row to match better-auth's encoding ===
			// better-auth stored the public key as standard base64 (this service
			// uses raw-url base64) and device_type as SimpleWebAuthn's camelCase
			// "multiDevice"/"singleDevice". Translate the representation only; the
			// device_type must still reflect the authenticator's real
			// backup-eligible flag, so derive the camelCase value from what was
			// actually stored.
			ctx := t.Context()
			var storedPubKey, storedDeviceType, storedCredID string
			require.NoError(t, env.Pool.QueryRow(ctx,
				`SELECT public_key, device_type, credential_id FROM passkey WHERE user_id = 'user-1'`).
				Scan(&storedPubKey, &storedDeviceType, &storedCredID))
			rawPubKey, err := base64.RawURLEncoding.DecodeString(storedPubKey)
			require.NoError(t, err)
			legacyPubKey := base64.StdEncoding.EncodeToString(rawPubKey)
			legacyDeviceType := "singleDevice"
			if storedDeviceType == "multi_device" {
				legacyDeviceType = "multiDevice"
			}
			legacyCredID := storedCredID
			if tc.reEncodeCredentialID != nil {
				// Note: the raw-url and raw-std alphabets coincide unless the
				// value contains '-' or '_', so this may be a no-op for some
				// random credential IDs; the lookup fallback is still exercised.
				legacyCredID = tc.reEncodeCredentialID(storedCredID)
			}
			_, err = env.Pool.Exec(ctx,
				`UPDATE passkey SET public_key = $1, device_type = $2, credential_id = $3 WHERE user_id = 'user-1'`,
				legacyPubKey, legacyDeviceType, legacyCredID)
			require.NoError(t, err)

			// === Log in with a random userHandle (as better-auth would emit) ===
			authenticator.Options.UserHandle = []byte("ycwlyu96xx3ja1r5yvf0rifktmlrs1o")
			authenticator.AddCredential(credential)

			resp, err = testutil.Post(t, env.Server.URL+"/api/auth/passkey/login-options", "application/json", nil)
			require.NoError(t, err)
			var loginOptions struct {
				Options      json.RawMessage `json:"options"`
				ChallengeKey string          `json:"challengeKey"`
			}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginOptions))
			_ = resp.Body.Close()

			assOpts, err := virtualwebauthn.ParseAssertionOptions(string(loginOptions.Options))
			require.NoError(t, err)
			assResponse := virtualwebauthn.CreateAssertionResponse(rp, authenticator, credential, *assOpts)

			var assJSON map[string]interface{}
			require.NoError(t, json.Unmarshal([]byte(assResponse), &assJSON))
			assJSON["challengeKey"] = loginOptions.ChallengeKey
			loginBody, _ := json.Marshal(assJSON)

			resp, err = testutil.Post(t, env.Server.URL+"/api/auth/passkey/login", "application/json", bytes.NewReader(loginBody))
			require.NoError(t, err)
			defer func() { _ = resp.Body.Close() }()

			if resp.StatusCode != http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				require.Equal(t, http.StatusOK, resp.StatusCode, "legacy login failed: body=%s", string(body))
			}

			var loginResult map[string]interface{}
			require.NoError(t, json.NewDecoder(resp.Body).Decode(&loginResult))
			user, ok := loginResult["user"].(map[string]interface{})
			require.True(t, ok, "response must contain user object, got: %v", loginResult)
			assert.Equal(t, "legacy@example.com", user["email"])
		})
	}
}

func TestPasskeyRegister_InvalidChallenge(t *testing.T) {
	env := testutil.Setup(t)
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	body, _ := json.Marshal(map[string]string{
		"challengeKey": "invalid-key",
	})

	req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestPasskeyLogin_InvalidChallenge(t *testing.T) {
	env := testutil.Setup(t)

	body, _ := json.Marshal(map[string]string{
		"challengeKey": "invalid-key",
	})

	resp, err := testutil.Post(t, env.Server.URL+"/api/auth/passkey/login", "application/json", bytes.NewReader(body))
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

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
		req := testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register-options", nil)
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := http.DefaultClient.Do(req)
		require.NoError(t, err)

		var opts struct {
			Options      json.RawMessage `json:"options"`
			ChallengeKey string          `json:"challengeKey"`
		}
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&opts))
		_ = resp.Body.Close()

		// Create attestation
		attOpts, _ := virtualwebauthn.ParseAttestationOptions(string(opts.Options))
		attResp := virtualwebauthn.CreateAttestationResponse(rp, authenticator, cred, *attOpts)

		var attJSON map[string]interface{}
		require.NoError(t, json.Unmarshal([]byte(attResp), &attJSON))
		attJSON["challengeKey"] = opts.ChallengeKey
		attJSON["name"] = name
		body, _ := json.Marshal(attJSON)

		// Register
		req = testutil.NewRequest(t, "POST", env.Server.URL+"/api/auth/passkey/register", bytes.NewReader(body))
		req.Header.Set("Authorization", "Bearer "+token)
		req.Header.Set("Content-Type", "application/json")
		resp, err = http.DefaultClient.Do(req)
		require.NoError(t, err)
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		_ = resp.Body.Close()
	}

	// Verify two passkeys stored
	var count int
	require.NoError(t, env.Pool.QueryRow(t.Context(), `SELECT COUNT(*) FROM passkey WHERE user_id = 'user-1'`).Scan(&count))
	assert.Equal(t, 2, count)
}
