package auth

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
)

type passkeyRegisterMetadata struct {
	ChallengeKey string `json:"challengeKey"`
	Name         string `json:"name"`
}

type passkeyLoginMetadata struct {
	ChallengeKey string `json:"challengeKey"`
}

func readJSONBody[T any](r *http.Request, dest *T) ([]byte, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(body, dest); err != nil {
		return nil, err
	}

	r.Body = io.NopCloser(bytes.NewReader(body))
	return body, nil
}

func (h *AuthHandler) loadWebauthnUser(ctx context.Context, userID string) (*webauthnUser, error) {
	dbUser, err := h.queries.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	dbPasskeys, err := h.queries.GetPasskeysByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}

	creds := make([]webauthn.Credential, 0, len(dbPasskeys))
	for _, passkey := range dbPasskeys {
		creds = append(creds, dbPasskeyToCredential(passkey))
	}

	return &webauthnUser{
		id:          dbUser.ID,
		name:        dbUser.Name,
		email:       dbUser.Email,
		credentials: creds,
	}, nil
}

func (h *AuthHandler) beginPasskeyRegistration(ctx context.Context, user *SessionUser) (*passkeyOptionsResponse, error) {
	webauthnUser, err := h.loadWebauthnUser(ctx, user.User.ID)
	if err != nil {
		return nil, err
	}

	creation, sessionData, err := h.wan.BeginRegistration(webauthnUser)
	if err != nil {
		return nil, err
	}

	challengeKey := generateToken()
	if err := h.challenges.Set(ctx, "webauthn:"+challengeKey, webauthnSessionEntry{
		SessionData: sessionData,
		UserID:      user.User.ID,
	}, 5*time.Minute); err != nil {
		return nil, err
	}

	return &passkeyOptionsResponse{
		Options:      creation,
		ChallengeKey: challengeKey,
	}, nil
}

func (h *AuthHandler) finishPasskeyRegistration(ctx context.Context, r *http.Request, user *SessionUser, meta passkeyRegisterMetadata) (*passkeyRegisterResponse, error) {
	var entry webauthnSessionEntry
	if err := h.challenges.Get(ctx, "webauthn:"+meta.ChallengeKey, &entry); err != nil || entry.UserID != user.User.ID {
		return nil, fmt.Errorf("invalid or expired challenge")
	}

	webauthnUser := &webauthnUser{
		id:    user.User.ID,
		name:  user.User.Name,
		email: user.User.Email,
	}

	credential, err := h.wan.FinishRegistration(webauthnUser, *entry.SessionData, r)
	if err != nil {
		return nil, err
	}

	credentialID, publicKey, deviceType, counter, backedUp, transports, aaguid := credentialToDBFields(credential)
	passkeyName := meta.Name
	if passkeyName == "" {
		passkeyName = "Passkey"
	}

	passkeyID := uuid.New().String()
	if err := h.queries.CreatePasskey(ctx, queries.CreatePasskeyParams{
		ID:           passkeyID,
		Name:         &passkeyName,
		PublicKey:    publicKey,
		UserID:       user.User.ID,
		CredentialID: credentialID,
		Counter:      counter,
		DeviceType:   deviceType,
		BackedUp:     backedUp,
		Transports:   transports,
		Aaguid:       aaguid,
	}); err != nil {
		return nil, err
	}

	return &passkeyRegisterResponse{
		ID:   passkeyID,
		Name: passkeyName,
	}, nil
}

func (h *AuthHandler) beginPasskeyLogin(ctx context.Context) (*passkeyOptionsResponse, error) {
	assertion, sessionData, err := h.wan.BeginDiscoverableLogin()
	if err != nil {
		return nil, err
	}

	challengeKey := generateToken()
	if err := h.challenges.Set(ctx, "webauthn:"+challengeKey, webauthnSessionEntry{
		SessionData: sessionData,
	}, 5*time.Minute); err != nil {
		return nil, err
	}

	return &passkeyOptionsResponse{
		Options:      assertion,
		ChallengeKey: challengeKey,
	}, nil
}

func (h *AuthHandler) finishPasskeyLogin(r *http.Request, challengeKey string) (*tokenUserResponse, error) {
	ctx := r.Context()
	var entry webauthnSessionEntry
	if err := h.challenges.Get(ctx, "webauthn:"+challengeKey, &entry); err != nil {
		return nil, errChallenge
	}

	var resolvedUserID string
	handler := func(rawID, userHandle []byte) (webauthn.User, error) {
		userID := string(userHandle)
		resolvedUserID = userID

		dbUser, err := h.queries.GetUserByID(ctx, userID)
		if err != nil {
			return nil, err
		}

		if dbUser.Banned != nil && *dbUser.Banned {
			return nil, errBanned
		}

		webauthnUser, err := h.loadWebauthnUser(ctx, userID)
		if err != nil || len(webauthnUser.credentials) == 0 {
			return nil, errNoPasskeys
		}

		return webauthnUser, nil
	}

	credential, err := h.wan.FinishDiscoverableLogin(handler, *entry.SessionData, r)
	if err != nil {
		return nil, err
	}

	credIDEncoded := base64.RawURLEncoding.EncodeToString(credential.ID)
	dbPasskey, err := h.queries.GetPasskeyByCredentialID(ctx, credIDEncoded)
	if err == nil {
		if err := h.queries.UpdatePasskeyCounter(ctx, queries.UpdatePasskeyCounterParams{
			Counter: int32(credential.Authenticator.SignCount),
			ID:      dbPasskey.ID,
		}); err != nil {
			return nil, err
		}
	}

	userRecord, err := h.queries.GetUserByID(ctx, resolvedUserID)
	if err != nil {
		return nil, err
	}

	sessionToken, err := h.createSession(r, resolvedUserID)
	if err != nil {
		return nil, err
	}

	return &tokenUserResponse{
		Token: sessionToken,
		User: authenticatedUserResponse{
			ID:    userRecord.ID,
			Name:  userRecord.Name,
			Email: userRecord.Email,
		},
	}, nil
}
