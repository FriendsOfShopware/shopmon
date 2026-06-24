package auth

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
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

	// handle, when set, overrides the WebAuthn user handle. During discoverable
	// login go-webauthn requires WebAuthnID() to equal the userHandle the
	// authenticator returns. Passkeys registered by the legacy better-auth stack
	// carry a random handle that is not the user ID, so we echo the asserted
	// handle here while still resolving the real user via the credential ID.
	handle []byte
}

func (u *webauthnUser) WebAuthnID() []byte {
	if u.handle != nil {
		return u.handle
	}
	return []byte(u.id)
}
func (u *webauthnUser) WebAuthnName() string                       { return u.email }
func (u *webauthnUser) WebAuthnDisplayName() string                { return u.name }
func (u *webauthnUser) WebAuthnCredentials() []webauthn.Credential { return u.credentials }

// --- DB <-> library conversion ---

// decodeBase64Any decodes a base64 string that may be in either standard or
// URL-safe alphabet, with or without padding. Passkeys registered by the legacy
// better-auth stack stored public keys with standard base64, while passkeys
// registered by this service use raw URL-safe base64; both must decode here.
func decodeBase64Any(s string) ([]byte, error) {
	for _, enc := range []*base64.Encoding{
		base64.RawURLEncoding,
		base64.RawStdEncoding,
		base64.URLEncoding,
		base64.StdEncoding,
	} {
		if b, err := enc.DecodeString(s); err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("decode base64: no supported encoding matched")
}

// isMultiDeviceCredential reports whether the stored device_type marks a
// multi-device (backup-eligible) credential. The legacy better-auth stack
// stored SimpleWebAuthn's "multiDevice"/"singleDevice" values, while this
// service stores "multi_device"/"single_device".
func isMultiDeviceCredential(deviceType string) bool {
	return deviceType == "multi_device" || deviceType == "multiDevice"
}

func dbPasskeyToCredential(p queries.Passkey) (webauthn.Credential, error) {
	credID, err := decodeBase64Any(p.CredentialID)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("decode credential id for passkey %s: %w", p.ID, err)
	}
	pubKey, err := decodeBase64Any(p.PublicKey)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("decode public key for passkey %s: %w", p.ID, err)
	}

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
			BackupEligible: isMultiDeviceCredential(p.DeviceType),
			BackupState:    p.BackedUp,
			UserPresent:    true,
			UserVerified:   true,
		},
		Authenticator: webauthn.Authenticator{
			AAGUID:    aaguid,
			SignCount: uint32(p.Counter),
		},
	}, nil
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
