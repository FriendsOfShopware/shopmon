package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/go-webauthn/webauthn/webauthn"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Compile-time check that AuthHandler implements the generated ServerInterface.
var _ authapi.ServerInterface = (*AuthHandler)(nil)

type AuthHandler struct {
	pool       *pgxpool.Pool
	queries    *queries.Queries
	cfg        *config.Config
	mail       *mail.Service
	wan        *webauthn.WebAuthn
	challenges *ChallengeStore
}

func NewAuthHandler(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, mail *mail.Service) *AuthHandler {
	var wan *webauthn.WebAuthn
	if cfg.WebAuthnRPID != "" {
		var err error
		wan, err = webauthn.New(&webauthn.Config{
			RPID:          cfg.WebAuthnRPID,
			RPDisplayName: cfg.WebAuthnRPName,
			RPOrigins:     cfg.WebAuthnRPOrigins,
		})
		if err != nil {
			panic("failed to create webauthn: " + err.Error())
		}
	}

	challenges := NewChallengeStore(cfg.RedisURL, "shopmon:challenge:")

	return &AuthHandler{
		pool:       pool,
		queries:    q,
		cfg:        cfg,
		mail:       mail,
		wan:        wan,
		challenges: challenges,
	}
}

// requireAuth extracts and validates the session token from the request.
// Returns nil and writes an error response if authentication fails.
func (h *AuthHandler) requireAuth(w http.ResponseWriter, r *http.Request) *SessionUser {
	token := httputil.ExtractToken(r)
	if token == "" {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return nil
	}
	su, err := ValidateSession(r.Context(), h.pool, token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
		return nil
	}
	return su
}

// requireAdmin extracts and validates the session token, then checks for admin role.
// Returns nil and writes an error response if authentication or authorization fails.
func (h *AuthHandler) requireAdmin(w http.ResponseWriter, r *http.Request) *SessionUser {
	su := h.requireAuth(w, r)
	if su == nil {
		return nil
	}
	if su.User.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, "admin access required")
		return nil
	}
	return su
}

func (h *AuthHandler) requireOrgRole(w http.ResponseWriter, r *http.Request, userID, organizationID string, allowedRoles ...string) string {
	role, err := h.queries.GetMemberRole(r.Context(), queries.GetMemberRoleParams{
		OrganizationID: organizationID,
		UserID:         userID,
	})
	if err != nil {
		httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
		return ""
	}

	for _, allowedRole := range allowedRoles {
		if role == allowedRole {
			return role
		}
	}

	httputil.WriteError(w, http.StatusForbidden, "insufficient permissions")
	return ""
}

func (h *AuthHandler) requireOrgMembership(w http.ResponseWriter, r *http.Request, userID, organizationID string) bool {
	err := h.requireOrganization(r.Context(), organizationID)
	if err != nil {
		if errors.Is(err, errOrganizationNotFound) {
			httputil.WriteError(w, http.StatusNotFound, "organization not found")
			return false
		}
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load organization")
		return false
	}

	role := h.requireOrgRole(w, r, userID, organizationID, "owner", "admin", "member")
	return role != ""
}

var errOrganizationNotFound = errors.New("organization not found")

func (h *AuthHandler) requireOrganization(ctx context.Context, organizationID string) error {
	_, err := h.queries.GetOrganizationByID(ctx, organizationID)
	if err != nil {
		return errOrganizationNotFound
	}

	return nil
}

// validateCallbackURL checks that the callback URL is either empty (will default
// to FrontendURL) or belongs to the same host as the configured frontend.
func (h *AuthHandler) validateCallbackURL(callbackURL string) bool {
	if callbackURL == "" {
		return true // will default to FrontendURL
	}
	parsed, err := url.Parse(callbackURL)
	if err != nil {
		return false
	}
	frontendParsed, err := url.Parse(h.cfg.FrontendURL)
	if err != nil {
		return false
	}
	return parsed.Host == frontendParsed.Host
}

// NewRateLimiter creates a new rate limiter for use in middleware setup.
func NewRateLimiter(ctx context.Context, window time.Duration, max int) *rateLimiter {
	return newRateLimiter(ctx, window, max)
}

func generateToken() string {
	b := make([]byte, 32)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// createOneTimeCode creates a session and stores the token behind a short-lived
// one-time code in Redis. Used by OAuth/SSO callbacks to avoid putting tokens in URLs.
func (h *AuthHandler) createOneTimeCode(r *http.Request, userID string) (string, error) {
	sessionToken, err := h.createSession(r, userID)
	if err != nil {
		return "", err
	}

	code := generateToken()
	if err = h.challenges.Set(r.Context(), "auth-code:"+code, sessionToken, 60*time.Second); err != nil {
		return "", err
	}
	return code, nil
}

// ExchangeCode exchanges a one-time authorization code for a session token.
func (h *AuthHandler) ExchangeCode(w http.ResponseWriter, r *http.Request) {
	var req exchangeCodeRequest
	if err := httputil.DecodeBody(r, &req); err != nil || req.Code == "" {
		httputil.WriteError(w, http.StatusBadRequest, "code is required")
		return
	}

	var sessionToken string
	if err := h.challenges.Get(r.Context(), "auth-code:"+req.Code, &sessionToken); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired code")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, exchangeCodeResponse{Token: sessionToken})
}

// createSession creates a new session with a 30-day expiry.
// TODO: Session tokens currently have a fixed 30-day expiry with no rotation.
func (h *AuthHandler) createSession(r *http.Request, userID string) (string, error) {
	token := generateToken()
	sessionID := uuid.New().String()

	ipAddress := r.RemoteAddr
	userAgent := r.UserAgent()

	_, err := h.queries.CreateSession(r.Context(), queries.CreateSessionParams{
		ID:        sessionID,
		ExpiresAt: pgtype.Timestamp{Time: time.Now().Add(30 * 24 * time.Hour), Valid: true},
		Token:     token,
		IpAddress: &ipAddress,
		UserAgent: &userAgent,
		UserID:    userID,
	})
	if err != nil {
		return "", err
	}

	return token, nil
}
