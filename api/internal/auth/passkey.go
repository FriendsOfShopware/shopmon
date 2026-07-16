package auth

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	identitypasskey "github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
)

type passkeyRegisterMetadata struct {
	ChallengeKey string `json:"challengeKey"`
	Name         string `json:"name"`
}

type passkeyLoginMetadata struct {
	ChallengeKey string `json:"challengeKey"`
}

func (h *AuthHandler) PasskeyRegisterOptions(w http.ResponseWriter, r *http.Request) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}
	options, err := h.passkeys.BeginRegistration(r.Context(), identitypasskey.User{
		ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to begin passkey registration", "userID", principal.User.ID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin registration")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, passkeyOptionsResponse{Options: options.Options, ChallengeKey: options.ChallengeKey})
}

func (h *AuthHandler) PasskeyRegister(w http.ResponseWriter, r *http.Request) {
	principal := h.requireAuth(w, r)
	if principal == nil {
		return
	}
	var metadata passkeyRegisterMetadata
	if _, err := readJSONBody(r, &metadata); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if metadata.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}
	registered, err := h.passkeys.FinishRegistration(r.Context(), r, identitypasskey.User{
		ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
	}, metadata.ChallengeKey, metadata.Name)
	if err != nil {
		slog.ErrorContext(r.Context(), "passkey registration failed", "userID", principal.User.ID, "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "registration failed: "+err.Error())
		return
	}
	httputil.WriteJSON(w, http.StatusOK, passkeyRegisterResponse{ID: registered.ID, Name: registered.Name})
}

func (h *AuthHandler) PasskeyLoginOptions(w http.ResponseWriter, r *http.Request) {
	options, err := h.passkeys.BeginLogin(r.Context())
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to begin passkey login", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to begin login")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, passkeyOptionsResponse{Options: options.Options, ChallengeKey: options.ChallengeKey})
}

func (h *AuthHandler) PasskeyLogin(w http.ResponseWriter, r *http.Request) {
	var metadata passkeyLoginMetadata
	if _, err := readJSONBody(r, &metadata); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if metadata.ChallengeKey == "" {
		httputil.WriteError(w, http.StatusBadRequest, "challengeKey is required")
		return
	}
	authentication, err := h.passkeys.FinishLogin(r, metadata.ChallengeKey, sessionMetadata(r))
	if err != nil {
		if errors.Is(err, identitypasskey.ErrChallenge) {
			httputil.WriteError(w, http.StatusBadRequest, identitypasskey.ErrChallenge.Error())
			return
		}
		slog.ErrorContext(r.Context(), "passkey authentication failed", "error", err)
		httputil.WriteError(w, http.StatusUnauthorized, "authentication failed")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: authentication.User.ID, Name: authentication.User.Name, Email: authentication.User.Email,
		},
	})
}

func readJSONBody[T any](r *http.Request, destination *T) ([]byte, error) {
	body, err := io.ReadAll(io.LimitReader(r.Body, 1<<20))
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(body, destination); err != nil {
		return nil, err
	}
	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}
