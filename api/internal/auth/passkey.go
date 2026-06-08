package auth

import (
	"encoding/base64"
	"encoding/hex"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/webauthn"
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
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	response, err := h.beginPasskeyRegistration(r.Context(), su)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin registration")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

// PasskeyRegister completes passkey registration.
func (h *AuthHandler) PasskeyRegister(w http.ResponseWriter, r *http.Request) {
	su := h.requireAuth(w, r)
	if su == nil {
		return
	}

	var meta passkeyRegisterMetadata
	if _, err := readJSONBody(r, &meta); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if meta.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}

	response, err := h.finishPasskeyRegistration(r.Context(), r, su, meta)
	if err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "registration failed: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

// --- Login ---

// PasskeyLoginOptions begins passkey authentication (no auth needed).
func (h *AuthHandler) PasskeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	response, err := h.beginPasskeyLogin(r.Context())
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin login")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

// PasskeyLogin completes passkey authentication.
func (h *AuthHandler) PasskeyLogin(w http.ResponseWriter, r *http.Request) {
	var meta passkeyLoginMetadata
	if _, err := readJSONBody(r, &meta); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if meta.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}

	response, err := h.finishPasskeyLogin(r, meta.ChallengeKey)
	if err != nil {
		if err == errChallenge {
			httputil.WriteError(w, http.StatusBadRequest, err.Error())
			return
		}
		httputil.WriteError(w, http.StatusUnauthorized, "authentication failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, response)
}

var (
	errBanned     = &passkeyError{"account is banned"}
	errNoPasskeys = &passkeyError{"no passkeys found"}
	errChallenge  = &passkeyError{"invalid or expired challenge"}
)

type passkeyError struct{ msg string }

func (e *passkeyError) Error() string { return e.msg }
