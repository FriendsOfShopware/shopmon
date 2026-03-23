package auth

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

type webauthnSessionEntry struct {
	SessionData *webauthn.SessionData `json:"sessionData"`
	UserID      string                `json:"userId"`
}

// webauthnUser adapts our User to the go-webauthn User interface.
type webauthnUser struct {
	id          string
	name        string
	email       string
	credentials []webauthn.Credential
}

func (u *webauthnUser) WebAuthnID() []byte                         { return []byte(u.id) }
func (u *webauthnUser) WebAuthnName() string                       { return u.email }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.name }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// --- DB <-> library conversion ---

func dbPasskeyToCredential(p queries.Passkey) webauthn.Credential {
	credID, _ := base64.RawURLEncoding.DecodeString(p.CredentialID)
	pubKey, _ := base64.RawURLEncoding.DecodeString(p.PublicKey)

	var aaguid []byte
	if p.Aaguid != nil {
		aaguid, _ = hex.DecodeString(*p.Aaguid)
	}

	var transports []protocol.AuthenticatorTransport
	if p.Transports != nil && *p.Transports != "" {
		for _, t := range strings.Split(*p.Transports, ",") {
			transports = append(transports, protocol.AuthenticatorTransport(strings.TrimSpace(t)))
		}
	}

	return webauthn.Credential{
		ID:        credID,
		PublicKey: pubKey,
		Transport: transports,
		Flags: webauthn.CredentialFlags{
			BackupEligible: p.DeviceType == "multi_device",
			BackupState:    p.BackedUp,
			UserPresent:    true,
			UserVerified:   true,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    aaguid,
			SignCount: uint32(p.Counter),
		},
	}
}

func credentialToDBFields(cred *webauthn.Credential) (credentialID, publicKey, deviceType string, counter int32, backedUp bool, transports, aaguid *string) {
	credentialID = base64.RawURLEncoding.EncodeToString(cred.ID)
	publicKey = base64.RawURLEncoding.EncodeToString(cred.PublicKey)

	if cred.Flags.BackupEligible {
		deviceType = "multi_device"
	} else {
		deviceType = "single_device"
	}

	counter = int32(cred.Authenticator.SignCount)
	backedUp = cred.Flags.BackupState

	if len(cred.Transport) > 0 {
		parts := make([]string, len(cred.Transport))
		for i, t := range cred.Transport {
			parts[i] = string(t)
		}
		s := strings.Join(parts, ",")
		transports = &s
	}

	if len(cred.Authenticator.AAGUID) > 0 {
		s := hex.EncodeToString(cred.Authenticator.AAGUID)
		aaguid = &s
	}

	return
}

// --- Registration ---

// PasskeyRegisterOptions begins passkey registration (user must be logged in).
func (h *AuthHandler) PasskeyRegisterOptions(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	su, err := ValidateSession(r.Context(), h.pool, token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	dbPasskeys, _ := h.queries.GetPasskeysByUserID(r.Context(), su.User.ID)
	var creds []webauthn.Credential
	for _, p := range dbPasskeys {
		creds = append(creds, dbPasskeyToCredential(p))
	}

	user := &webauthnUser{
		id:          su.User.ID,
		name:        su.User.Name,
		email:       su.User.Email,
		credentials: creds,
	}

	creation, sessionData, err := h.wan.BeginRegistration(user)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin registration")
		return
	}

	challengeKey := generateToken()
	h.challenges.Set(r.Context(), "webauthn:"+challengeKey, webauthnSessionEntry{
		SessionData: sessionData,
		UserID:      su.User.ID,
	}, 5*time.Minute)

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"options":      creation,
		"challengeKey": challengeKey,
	})
}

// PasskeyRegister completes passkey registration.
func (h *AuthHandler) PasskeyRegister(w http.ResponseWriter, r *http.Request) {
	token := httputil.ExtractToken(r)
	su, err := ValidateSession(r.Context(), h.pool, token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "not authenticated")
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var meta struct {
		ChallengeKey string `json:"challengeKey"`
		Name         string `json:"name"`
	}
	json.Unmarshal(body, &meta)

	if meta.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}

	var entry webauthnSessionEntry
	if err := h.challenges.Get(r.Context(), "webauthn:"+meta.ChallengeKey, &entry); err != nil || entry.UserID != su.User.ID {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired challenge")
		return
	}

	// Reconstruct body for the library
	r.Body = io.NopCloser(bytes.NewReader(body))

	user := &webauthnUser{
		id:    su.User.ID,
		name:  su.User.Name,
		email: su.User.Email,
	}

	credential, err := h.wan.FinishRegistration(user, *entry.SessionData, r)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "registration failed: "+err.Error())
		return
	}

	credentialID, publicKey, deviceType, counter, backedUp, transports, aaguid := credentialToDBFields(credential)

	passkeyName := meta.Name
	if passkeyName == "" {
		passkeyName = "Passkey"
	}

	passkeyID := uuid.New().String()
	if err := h.queries.CreatePasskey(r.Context(), queries.CreatePasskeyParams{
		ID:           passkeyID,
		Name:         &passkeyName,
		PublicKey:    publicKey,
		UserID:       su.User.ID,
		CredentialID: credentialID,
		Counter:      counter,
		DeviceType:   deviceType,
		BackedUp:     backedUp,
		Transports:   transports,
		Aaguid:       aaguid,
	}); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save passkey")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"id":   passkeyID,
		"name": passkeyName,
	})
}

// --- Login ---

// PasskeyLoginOptions begins passkey authentication (no auth needed).
func (h *AuthHandler) PasskeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	assertion, sessionData, err := h.wan.BeginDiscoverableLogin()
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin login")
		return
	}

	challengeKey := generateToken()
	h.challenges.Set(r.Context(), "webauthn:"+challengeKey, webauthnSessionEntry{
		SessionData: sessionData,
	}, 5*time.Minute)

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"options":      assertion,
		"challengeKey": challengeKey,
	})
}

// PasskeyLogin completes passkey authentication.
func (h *AuthHandler) PasskeyLogin(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "failed to read request body")
		return
	}

	var meta struct {
		ChallengeKey string `json:"challengeKey"`
	}
	json.Unmarshal(body, &meta)

	if meta.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}

	var entry webauthnSessionEntry
	if err := h.challenges.Get(r.Context(), "webauthn:"+meta.ChallengeKey, &entry); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired challenge")
		return
	}

	r.Body = io.NopCloser(bytes.NewReader(body))

	// Discoverable login handler — resolves user from the assertion's userHandle
	var resolvedUserID string
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userID := string(userHandle)
		resolvedUserID = userID

		dbUser, err := h.queries.GetUserByID(r.Context(), userID)
		if err != nil {
			return nil, err
		}

		if dbUser.Banned != nil && *dbUser.Banned {
			return nil, errBanned
		}

		dbPasskeys, err := h.queries.GetPasskeysByUserID(r.Context(), userID)
		if err != nil || len(dbPasskeys) == 0 {
			return nil, errNoPasskeys
		}

		var creds []webauthn.Credential
		for _, p := range dbPasskeys {
			creds = append(creds, dbPasskeyToCredential(p))
		}

		return &webauthnUser{
			id:          dbUser.ID,
			name:        dbUser.Name,
			email:       dbUser.Email,
			credentials: creds,
		}, nil
	}

	credential, err := h.wan.FinishDiscoverableLogin(handler, *entry.SessionData, r)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "authentication failed")
		return
	}

	userID := resolvedUserID

	// Update sign counter
	credIDEncoded := base64.RawURLEncoding.EncodeToString(credential.ID)
	dbPasskey, err := h.queries.GetPasskeyByCredentialID(r.Context(), credIDEncoded)
	if err == nil {
		h.queries.UpdatePasskeyCounter(r.Context(), queries.UpdatePasskeyCounterParams{
			Counter: int32(credential.Authenticator.SignCount),
			ID:      dbPasskey.ID,
		})
	}
	sessionToken, err := h.createSession(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Fetch user info for the response
	dbUser, err := h.queries.GetUserByID(r.Context(), userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get user")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"token": sessionToken,
		"user": map[string]interface{}{
			"id":    dbUser.ID,
			"name":  dbUser.Name,
			"email": dbUser.Email,
		},
	})
}

var (
	errBanned     = &passkeyError{"account is banned"}
	errNoPasskeys = &passkeyError{"no passkeys found"}
)

type passkeyError struct{ msg string }

func (e *passkeyError) Error() string { return e.msg }
