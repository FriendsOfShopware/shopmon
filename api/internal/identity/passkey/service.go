package passkey

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/go-webauthn/webauthn/protocol"
	"github.com/go-webauthn/webauthn/protocol/webauthncose"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrNotConfigured = errors.New("passkeys are not configured")
	ErrChallenge     = errors.New("invalid or expired challenge")
	ErrNoPasskeys    = errors.New("no passkeys found")
	ErrBanned        = errors.New("account is banned")
	ErrNotFound      = errors.New("passkey not found")
)

const challengeLifetime = 5 * time.Minute

type User struct {
	ID    string
	Name  string
	Email string
	Ban   identity.Ban
}

type StoredPasskey struct {
	ID           string
	Name         *string
	PublicKey    string
	UserID       string
	CredentialID string
	Counter      int32
	DeviceType   string
	BackedUp     bool
	Transports   *string
	AAGUID       *string
}

type ChallengeStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, destination any) error
}

type SessionIssuer interface {
	Issue(ctx context.Context, userID string, metadata identity.SessionMetadata) (string, error)
}

type Config struct {
	RPID      string
	RPName    string
	RPOrigins []string
}

type Service struct {
	webauthn   *webauthn.WebAuthn
	queries    *queries.Queries
	challenges ChallengeStore
	sessions   SessionIssuer
}

func NewService(config Config, q *queries.Queries, challenges ChallengeStore, sessions SessionIssuer) (*Service, error) {
	if config.RPID == "" {
		return &Service{queries: q, challenges: challenges, sessions: sessions}, nil
	}
	wan, err := webauthn.New(&webauthn.Config{
		RPID: config.RPID, RPDisplayName: config.RPName, RPOrigins: config.RPOrigins,
	})
	if err != nil {
		return nil, fmt.Errorf("configure WebAuthn: %w", err)
	}
	return &Service{webauthn: wan, queries: q, challenges: challenges, sessions: sessions}, nil
}

type Options struct {
	Options      any
	ChallengeKey string
}

type Registered struct {
	ID   string
	Name string
}

type Authentication struct {
	Token string
	User  User
}

type sessionEntry struct {
	SessionData *webauthn.SessionData `json:"sessionData"`
	UserID      string                `json:"userId"`
}

func (s *Service) BeginRegistration(ctx context.Context, user User) (Options, error) {
	if s.webauthn == nil {
		return Options{}, ErrNotConfigured
	}
	webauthnUser, err := s.loadUser(ctx, user.ID)
	if err != nil {
		return Options{}, fmt.Errorf("load passkey registration user: %w", err)
	}
	creation, sessionData, err := s.webauthn.BeginRegistration(webauthnUser)
	if err != nil {
		return Options{}, fmt.Errorf("begin WebAuthn registration: %w", err)
	}
	challengeKey, err := generateToken()
	if err != nil {
		return Options{}, fmt.Errorf("generate registration challenge key: %w", err)
	}
	if err := s.challenges.Set(ctx, "webauthn:"+challengeKey, sessionEntry{SessionData: sessionData, UserID: user.ID}, challengeLifetime); err != nil {
		return Options{}, fmt.Errorf("store registration challenge: %w", err)
	}
	return Options{Options: creation, ChallengeKey: challengeKey}, nil
}

func (s *Service) FinishRegistration(ctx context.Context, request *http.Request, user User, challengeKey, name string) (Registered, error) {
	if s.webauthn == nil {
		return Registered{}, ErrNotConfigured
	}
	var entry sessionEntry
	if err := s.challenges.Get(ctx, "webauthn:"+challengeKey, &entry); err != nil || entry.UserID != user.ID {
		return Registered{}, ErrChallenge
	}
	credential, err := s.webauthn.FinishRegistration(&webauthnUser{id: user.ID, name: user.Name, email: user.Email}, *entry.SessionData, request)
	if err != nil {
		return Registered{}, fmt.Errorf("finish WebAuthn registration: %w", err)
	}
	stored, err := credentialToStored(user.ID, credential)
	if err != nil {
		return Registered{}, err
	}
	stored.ID = uuid.NewString()
	if name == "" {
		name = "Passkey"
	}
	stored.Name = &name
	if err := s.queries.CreatePasskey(ctx, queries.CreatePasskeyParams{
		ID: stored.ID, Name: stored.Name, PublicKey: stored.PublicKey, UserID: stored.UserID,
		CredentialID: stored.CredentialID, Counter: stored.Counter, DeviceType: stored.DeviceType,
		BackedUp: stored.BackedUp, Transports: stored.Transports, Aaguid: stored.AAGUID,
	}); err != nil {
		return Registered{}, fmt.Errorf("create passkey: %w", err)
	}
	return Registered{ID: stored.ID, Name: name}, nil
}

func (s *Service) BeginLogin(ctx context.Context) (Options, error) {
	if s.webauthn == nil {
		return Options{}, ErrNotConfigured
	}
	assertion, sessionData, err := s.webauthn.BeginDiscoverableLogin()
	if err != nil {
		return Options{}, fmt.Errorf("begin WebAuthn login: %w", err)
	}
	challengeKey, err := generateToken()
	if err != nil {
		return Options{}, fmt.Errorf("generate login challenge key: %w", err)
	}
	if err := s.challenges.Set(ctx, "webauthn:"+challengeKey, sessionEntry{SessionData: sessionData}, challengeLifetime); err != nil {
		return Options{}, fmt.Errorf("store login challenge: %w", err)
	}
	return Options{Options: assertion, ChallengeKey: challengeKey}, nil
}

func (s *Service) FinishLogin(request *http.Request, challengeKey string, metadata identity.SessionMetadata) (Authentication, error) {
	if s.webauthn == nil {
		return Authentication{}, ErrNotConfigured
	}
	ctx := request.Context()
	var entry sessionEntry
	if err := s.challenges.Get(ctx, "webauthn:"+challengeKey, &entry); err != nil {
		return Authentication{}, ErrChallenge
	}

	var resolved User
	resolve := func(rawID, userHandle []byte) (webauthn.User, error) {
		stored, err := s.findByRawCredentialID(ctx, rawID)
		if err != nil {
			return nil, ErrNoPasskeys
		}
		resolved, err = s.findUser(ctx, stored.UserID)
		if err != nil {
			return nil, err
		}
		if resolved.Ban.ActiveAt(time.Now()) {
			return nil, ErrBanned
		}
		user, err := s.loadUser(ctx, stored.UserID)
		if err != nil || len(user.credentials) == 0 {
			return nil, ErrNoPasskeys
		}
		user.handle = userHandle
		return user, nil
	}

	credential, err := s.webauthn.FinishDiscoverableLogin(resolve, *entry.SessionData, request)
	if err != nil {
		return Authentication{}, fmt.Errorf("finish WebAuthn login: %w", err)
	}
	stored, err := s.findByRawCredentialID(ctx, credential.ID)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return Authentication{}, fmt.Errorf("reload authenticated passkey: %w", err)
	}
	if err == nil {
		if err := s.queries.UpdatePasskeyCounter(ctx, queries.UpdatePasskeyCounterParams{Counter: int32(credential.Authenticator.SignCount), ID: stored.ID}); err != nil {
			return Authentication{}, fmt.Errorf("update passkey counter: %w", err)
		}
	}
	token, err := s.sessions.Issue(ctx, resolved.ID, metadata)
	if err != nil {
		return Authentication{}, fmt.Errorf("issue passkey session: %w", err)
	}
	return Authentication{Token: token, User: resolved}, nil
}

func (s *Service) loadUser(ctx context.Context, userID string) (*webauthnUser, error) {
	user, err := s.findUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	stored, err := s.listByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	credentials := make([]webauthn.Credential, 0, len(stored))
	for _, passkey := range stored {
		credential, err := storedToCredential(passkey)
		if err != nil {
			return nil, err
		}
		credentials = append(credentials, credential)
	}
	return &webauthnUser{id: user.ID, name: user.Name, email: user.Email, credentials: credentials}, nil
}

func (s *Service) findByRawCredentialID(ctx context.Context, rawID []byte) (StoredPasskey, error) {
	var lastErr error
	for _, encoding := range []*base64.Encoding{base64.RawURLEncoding, base64.RawStdEncoding} {
		passkey, err := s.findByCredentialID(ctx, encoding.EncodeToString(rawID))
		if err == nil {
			return passkey, nil
		}
		lastErr = err
	}
	return StoredPasskey{}, lastErr
}

func (s *Service) findUser(ctx context.Context, userID string) (User, error) {
	row, err := s.queries.GetUserByID(ctx, userID)
	if err != nil {
		return User{}, fmt.Errorf("query passkey user: %w", err)
	}
	var banExpires *time.Time
	if row.BanExpires.Valid {
		value := row.BanExpires.Time
		banExpires = &value
	}
	return User{
		ID: row.ID, Name: row.Name, Email: row.Email,
		Ban: identity.Ban{Enabled: row.Banned != nil && *row.Banned, ExpiresAt: banExpires},
	}, nil
}

func (s *Service) listByUser(ctx context.Context, userID string) ([]StoredPasskey, error) {
	rows, err := s.queries.GetPasskeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("query user passkeys: %w", err)
	}
	result := make([]StoredPasskey, 0, len(rows))
	for _, row := range rows {
		result = append(result, mapPasskey(row))
	}
	return result, nil
}

func (s *Service) findByCredentialID(ctx context.Context, credentialID string) (StoredPasskey, error) {
	row, err := s.queries.GetPasskeyByCredentialID(ctx, credentialID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return StoredPasskey{}, ErrNotFound
		}
		return StoredPasskey{}, fmt.Errorf("query passkey by credential ID: %w", err)
	}
	return StoredPasskey{
		ID: row.ID, Name: row.Name, PublicKey: row.PublicKey, UserID: row.UserID,
		CredentialID: row.CredentialID, Counter: row.Counter, DeviceType: row.DeviceType,
		BackedUp: row.BackedUp, Transports: row.Transports, AAGUID: row.Aaguid,
	}, nil
}

func mapPasskey(row queries.Passkey) StoredPasskey {
	return StoredPasskey{
		ID: row.ID, Name: row.Name, PublicKey: row.PublicKey, UserID: row.UserID,
		CredentialID: row.CredentialID, Counter: row.Counter, DeviceType: row.DeviceType,
		BackedUp: row.BackedUp, Transports: row.Transports, AAGUID: row.Aaguid,
	}
}

type webauthnUser struct {
	id          string
	name        string
	email       string
	credentials []webauthn.Credential
	handle      []byte
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

func storedToCredential(stored StoredPasskey) (webauthn.Credential, error) {
	credentialID, err := decodeBase64Any(stored.CredentialID)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("decode credential ID for passkey %s: %w", stored.ID, err)
	}
	publicKey, err := decodeCOSEPublicKey(stored.PublicKey)
	if err != nil {
		return webauthn.Credential{}, fmt.Errorf("decode public key for passkey %s: %w", stored.ID, err)
	}
	var aaguid []byte
	if stored.AAGUID != nil {
		aaguid, err = hex.DecodeString(*stored.AAGUID)
		if err != nil {
			return webauthn.Credential{}, fmt.Errorf("decode AAGUID for passkey %s: %w", stored.ID, err)
		}
	}
	var transports []protocol.AuthenticatorTransport
	if stored.Transports != nil && *stored.Transports != "" {
		for transport := range strings.SplitSeq(*stored.Transports, ",") {
			transports = append(transports, protocol.AuthenticatorTransport(strings.TrimSpace(transport)))
		}
	}
	return webauthn.Credential{
		ID: credentialID, PublicKey: publicKey, Transport: transports,
		Flags: webauthn.CredentialFlags{
			BackupEligible: stored.DeviceType == "multi_device" || stored.DeviceType == "multiDevice",
			BackupState:    stored.BackedUp, UserPresent: true, UserVerified: true,
		},
		Authenticator: webauthn.Authenticator{AAGUID: aaguid, SignCount: uint32(stored.Counter)},
	}, nil
}

func credentialToStored(userID string, credential *webauthn.Credential) (StoredPasskey, error) {
	stored := StoredPasskey{
		UserID: userID, CredentialID: base64.RawURLEncoding.EncodeToString(credential.ID),
		PublicKey: base64.RawURLEncoding.EncodeToString(credential.PublicKey),
		Counter:   int32(credential.Authenticator.SignCount), BackedUp: credential.Flags.BackupState,
		DeviceType: "single_device",
	}
	if credential.Flags.BackupEligible {
		stored.DeviceType = "multi_device"
	}
	if len(credential.Transport) > 0 {
		values := make([]string, len(credential.Transport))
		for index, transport := range credential.Transport {
			values[index] = string(transport)
		}
		value := strings.Join(values, ",")
		stored.Transports = &value
	}
	if len(credential.Authenticator.AAGUID) > 0 {
		value := hex.EncodeToString(credential.Authenticator.AAGUID)
		stored.AAGUID = &value
	}
	return stored, nil
}

func decodeBase64Any(value string) ([]byte, error) {
	for _, encoding := range base64Encodings {
		if decoded, err := encoding.DecodeString(value); err == nil {
			return decoded, nil
		}
	}
	return nil, fmt.Errorf("no supported base64 encoding matched")
}

// decodeCOSEPublicKey decodes a stored WebAuthn public key and returns the
// first candidate that parses as a COSE key. Trying encodings blindly can
// succeed with the wrong variant and yield "Unsupported Public Key Type" at
// assertion time — the production symptom for some better-auth rows.
func decodeCOSEPublicKey(value string) ([]byte, error) {
	var candidates [][]byte
	var lastErr error
	for _, encoding := range base64Encodings {
		decoded, err := encoding.DecodeString(value)
		if err != nil {
			lastErr = err
			continue
		}
		candidates = append(candidates, decoded)
	}
	if decoded, err := hex.DecodeString(value); err == nil {
		candidates = append(candidates, decoded)
	}
	for _, decoded := range candidates {
		if _, err := webauthncose.ParsePublicKey(decoded); err != nil {
			lastErr = err
			continue
		}
		return decoded, nil
	}
	if lastErr != nil {
		return nil, fmt.Errorf("no supported COSE public key encoding matched: %w", lastErr)
	}
	return nil, fmt.Errorf("no supported COSE public key encoding matched")
}

var base64Encodings = []*base64.Encoding{
	base64.RawURLEncoding, base64.RawStdEncoding, base64.URLEncoding, base64.StdEncoding,
}

func generateToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", err
	}
	return hex.EncodeToString(value), nil
}
