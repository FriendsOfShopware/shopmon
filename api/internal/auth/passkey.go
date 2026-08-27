package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	identitypasskey "github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
)

type passkeyRegisterMetadata struct {
	ChallengeKey string `json:"challengeKey"`
	Name         string `json:"name"`
}

type passkeyLoginMetadata struct {
	ChallengeKey string `json:"challengeKey"`
}

func (h *AuthHandler) PasskeyRegisterOptions(ctx context.Context, _ *struct{}) (*passkeyOptionsOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	options, err := h.passkeys.BeginRegistration(ctx, identitypasskey.User{
		ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin passkey registration", "userID", principal.User.ID, "error", err)
		return nil, huma.Error500InternalServerError("failed to begin registration")
	}
	return &passkeyOptionsOutput{Body: passkeyOptionsResponse{Options: options.Options, ChallengeKey: options.ChallengeKey}}, nil
}

func (h *AuthHandler) PasskeyRegister(ctx context.Context, input *passkeyRegisterInput) (*passkeyRegisterOutput, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	var metadata passkeyRegisterMetadata
	if err := json.Unmarshal(input.Body, &metadata); err != nil {
		return nil, huma.Error400BadRequest("invalid request body")
	}
	if metadata.ChallengeKey == "" {
		return nil, huma.Error400BadRequest("challengeKey is required")
	}
	registered, err := h.passkeys.FinishRegistration(ctx, requestWithRawBody(ctx, input.Body), identitypasskey.User{
		ID: principal.User.ID, Name: principal.User.Name, Email: principal.User.Email,
	}, metadata.ChallengeKey, metadata.Name)
	if err != nil {
		slog.ErrorContext(ctx, "passkey registration failed", "userID", principal.User.ID, "error", err)
		return nil, huma.Error400BadRequest("registration failed: " + err.Error())
	}
	return &passkeyRegisterOutput{Body: passkeyRegisterResponse{ID: registered.ID, Name: registered.Name}}, nil
}

func (h *AuthHandler) PasskeyLoginOptions(ctx context.Context, _ *struct{}) (*passkeyOptionsOutput, error) {
	options, err := h.passkeys.BeginLogin(ctx)
	if err != nil {
		slog.ErrorContext(ctx, "failed to begin passkey login", "error", err)
		return nil, huma.Error500InternalServerError("failed to begin login")
	}
	return &passkeyOptionsOutput{Body: passkeyOptionsResponse{Options: options.Options, ChallengeKey: options.ChallengeKey}}, nil
}

func (h *AuthHandler) PasskeyLogin(ctx context.Context, input *passkeyLoginInput) (*tokenUserOutput, error) {
	var metadata passkeyLoginMetadata
	if err := json.Unmarshal(input.Body, &metadata); err != nil {
		return nil, huma.Error400BadRequest("invalid request body")
	}
	if metadata.ChallengeKey == "" {
		return nil, huma.Error400BadRequest("challengeKey is required")
	}
	r := requestWithRawBody(ctx, input.Body)
	authentication, err := h.passkeys.FinishLogin(r, metadata.ChallengeKey, sessionMetadata(r))
	if err != nil {
		if errors.Is(err, identitypasskey.ErrChallenge) {
			return nil, huma.Error400BadRequest(identitypasskey.ErrChallenge.Error())
		}
		slog.ErrorContext(ctx, "passkey authentication failed", "error", err)
		return nil, huma.Error401Unauthorized("authentication failed")
	}
	return &tokenUserOutput{Body: tokenUserResponse{
		Token: authentication.Token,
		User: authenticatedUserResponse{
			ID: authentication.User.ID, Name: authentication.User.Name, Email: authentication.User.Email,
		},
	}}, nil
}

func requestWithRawBody(ctx context.Context, raw []byte) *http.Request {
	r := requestFromContext(ctx)
	clone := r.Clone(ctx)
	if clone.Header == nil {
		clone.Header = make(http.Header)
	}
	clone.Body = io.NopCloser(bytes.NewReader(raw))
	clone.ContentLength = int64(len(raw))
	if clone.Header.Get("Content-Type") == "" {
		clone.Header.Set("Content-Type", "application/json")
	}
	return clone
}
