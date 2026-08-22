package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/audit"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	identitypasskey "github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
)

type AuthHandler struct {
	config        HandlerConfig
	challenges    StateStore
	organizations *organization.Service
	sso           *organizationsso.Service
	accounts      *identity.AccountService
	adminUsers    *identity.AdminService
	credentials   *identity.CredentialService
	sessions      *identity.SessionService
	federated     *identity.FederatedService
	passkeys      *identitypasskey.Service
	audit         *audit.Service
}

type StateStore interface {
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Get(ctx context.Context, key string, destination any) error
}

type HandlerConfig struct {
	FrontendURL        string
	GitHubClientID     string
	GitHubClientSecret string
}

type Dependencies struct {
	Organizations *organization.Service
	SSO           *organizationsso.Service
	Accounts      *identity.AccountService
	AdminUsers    *identity.AdminService
	Credentials   *identity.CredentialService
	Sessions      *identity.SessionService
	Federated     *identity.FederatedService
	Passkeys      *identitypasskey.Service
	Audit         *audit.Service
	State         StateStore
	Config        HandlerConfig
}

func NewAuthHandler(dependencies Dependencies) *AuthHandler {
	return &AuthHandler{
		config:        dependencies.Config,
		challenges:    dependencies.State,
		organizations: dependencies.Organizations,
		sso:           dependencies.SSO,
		accounts:      dependencies.Accounts,
		adminUsers:    dependencies.AdminUsers,
		credentials:   dependencies.Credentials,
		sessions:      dependencies.Sessions,
		federated:     dependencies.Federated,
		passkeys:      dependencies.Passkeys,
		audit:         dependencies.Audit,
	}
}

func (h *AuthHandler) requireAuth(ctx context.Context) (*access.Principal, error) {
	principal := access.PrincipalFromContext(ctx)
	if principal == nil {
		return nil, huma.Error401Unauthorized("unauthorized")
	}
	return principal, nil
}

func (h *AuthHandler) requireAdmin(ctx context.Context) (*access.Principal, error) {
	principal, err := h.requireAuth(ctx)
	if err != nil {
		return nil, err
	}
	if principal.User.Role != "admin" {
		return nil, huma.Error403Forbidden(httputil.MsgAdminRequired)
	}
	return principal, nil
}

// requireAuthHTTP is used by unconverted http.Handler methods.
func (h *AuthHandler) requireAuthHTTP(w http.ResponseWriter, r *http.Request) *access.Principal {
	principal, err := h.requireAuth(r.Context())
	if err != nil {
		httputil.WriteStatusError(w, err)
		return nil
	}
	return principal
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
	frontendParsed, err := url.Parse(h.config.FrontendURL)
	if err != nil {
		return false
	}
	return parsed.Host == frontendParsed.Host
}

// NewRateLimiter creates a new rate limiter for use in middleware setup.
func NewRateLimiter(ctx context.Context, window time.Duration, max int) *rateLimiter {
	return newRateLimiter(ctx, window, max)
}

func generateToken() (string, error) {
	value := make([]byte, 32)
	if _, err := rand.Read(value); err != nil {
		return "", fmt.Errorf("generate secure token: %w", err)
	}
	return hex.EncodeToString(value), nil
}

// createOneTimeCode creates a session and stores the token behind a short-lived
// one-time code in Redis. Used by OAuth/SSO callbacks to avoid putting tokens in URLs.
func (h *AuthHandler) createOneTimeCode(r *http.Request, userID string) (string, error) {
	sessionToken, err := h.createSession(r, userID)
	if err != nil {
		return "", err
	}

	code, err := generateToken()
	if err != nil {
		return "", err
	}
	if err = h.challenges.Set(r.Context(), "auth-code:"+code, sessionToken, 60*time.Second); err != nil {
		return "", err
	}
	return code, nil
}

// ExchangeCode exchanges a one-time authorization code for a session token.
func (h *AuthHandler) ExchangeCode(ctx context.Context, input *exchangeCodeInput) (*exchangeCodeOutput, error) {
	if input.Body.Code == "" {
		return nil, huma.Error400BadRequest("code is required")
	}

	var sessionToken string
	if err := h.challenges.Get(ctx, "auth-code:"+input.Body.Code, &sessionToken); err != nil {
		return nil, huma.Error400BadRequest("invalid or expired code")
	}

	return &exchangeCodeOutput{Body: exchangeCodeResponse{Token: sessionToken}}, nil
}

// createSession captures transport metadata and delegates session issuance to
// the identity capability. OAuth, SSO and passkey callbacks share this path.
func (h *AuthHandler) createSession(r *http.Request, userID string) (string, error) {
	return h.sessions.Issue(r.Context(), userID, sessionMetadata(r))
}

func sessionMetadata(r *http.Request) identity.SessionMetadata {
	userAgent := r.UserAgent()
	ipAddress := chimiddleware.GetClientIP(r.Context())
	if ipAddress == "" {
		ipAddress = r.RemoteAddr
	}
	return identity.SessionMetadata{
		IPAddress: ipAddress,
		UserAgent: userAgent,
	}
}
