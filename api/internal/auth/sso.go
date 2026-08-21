package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
)

// isTrue interprets an OIDC claim value (which may be a bool or a string) as a
// boolean. OIDC providers sometimes encode email_verified as the string "true".
func isTrue(v interface{}) bool {
	switch val := v.(type) {
	case bool:
		return val
	case string:
		return val == "true"
	default:
		return false
	}
}

type ssoState struct {
	ProviderID  string `json:"providerId"`
	CallbackURL string `json:"callbackURL"`
	Nonce       string `json:"nonce"`
}

// SignInSSO initiates SSO login by looking up provider by email domain.
func (h *AuthHandler) SignInSSO(ctx context.Context, input *ssoSignInInput) (*urlOutput, error) {
	req := input.Body

	if req.Email == "" {
		return nil, huma.Error400BadRequest("email is required")
	}

	// Extract domain from email
	parts := strings.SplitN(req.Email, "@", 2)
	if len(parts) != 2 {
		return nil, huma.Error400BadRequest("invalid email format")
	}
	domain := strings.ToLower(parts[1])

	// Look up SSO provider by domain
	provider, err := h.sso.FindByDomain(ctx, domain)
	if err != nil {
		return nil, huma.Error404NotFound("no SSO provider configured for this domain")
	}
	cfg := provider.Configuration

	if !h.validateCallbackURL(req.CallbackURL) {
		return nil, huma.Error400BadRequest("invalid callback URL")
	}

	// Generate state and nonce
	state, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate SSO state", "error", err)
		return nil, huma.Error500InternalServerError("failed to generate SSO state")
	}
	nonce, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate SSO nonce", "error", err)
		return nil, huma.Error500InternalServerError("failed to generate SSO nonce")
	}

	// Store state in Redis
	if err := h.challenges.Set(ctx, "sso:"+state, ssoState{
		ProviderID:  provider.ID,
		CallbackURL: req.CallbackURL,
		Nonce:       nonce,
	}, 10*time.Minute); err != nil {
		return nil, huma.Error500InternalServerError("failed to store SSO state")
	}

	// Build authorization URL
	callbackURL := fmt.Sprintf("%s/auth/sso/callback/%s", h.config.FrontendURL, provider.ID)

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&nonce=%s",
		cfg.AuthorizationEndpoint,
		url.QueryEscape(cfg.ClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("openid profile email"),
		url.QueryEscape(state),
		url.QueryEscape(nonce),
	)

	return &urlOutput{Body: urlResponse{URL: authURL}}, nil
}

// SsoCallback handles the OIDC callback after IdP authentication.
func (h *AuthHandler) SsoCallback(ctx context.Context, input *ssoCallbackInput) (*authCodeOutput, error) {
	if input.Code == "" || input.State == "" {
		return nil, huma.Error400BadRequest("missing or invalid parameters")
	}

	providerId := input.ProviderID

	// Validate state
	var stateData ssoState
	if err := h.challenges.Get(ctx, "sso:"+input.State, &stateData); err != nil {
		return nil, huma.Error400BadRequest("invalid or expired state")
	}

	if stateData.ProviderID != providerId {
		return nil, huma.Error400BadRequest("provider mismatch")
	}

	provider, err := h.sso.FindByID(ctx, providerId)
	if err != nil {
		return nil, huma.Error404NotFound("SSO provider not found")
	}
	cfg := provider.Configuration

	// Re-validate the stored endpoints before issuing any server-side request.
	// Create/update validation can be bypassed by pre-existing or out-of-band
	// rows, so a row with an http:// non-loopback or private/loopback endpoint
	// must be rejected here (loopback dev URLs are still allowed).
	for _, endpoint := range []string{cfg.TokenEndpoint, cfg.JWKSEndpoint, provider.Issuer} {
		if err := httputil.ValidateSSOEndpoint(endpoint); err != nil {
			slog.Error("stored SSO endpoint failed validation", "error", err, "provider", providerId)
			return nil, huma.Error502BadGateway("SSO provider has an invalid endpoint configured")
		}
	}

	// Exchange code for tokens
	callbackURL := fmt.Sprintf("%s/auth/sso/callback/%s", h.config.FrontendURL, providerId)

	clientSecret := ""
	if cfg.ClientSecret != nil {
		clientSecret = *cfg.ClientSecret
	}
	tokenForm := url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {clientSecret},
		"code":          {input.Code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {callbackURL},
	}
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost, cfg.TokenEndpoint, strings.NewReader(tokenForm.Encode()))
	if err != nil {
		slog.Error("failed to build SSO token request", "error", err)
		return nil, huma.Error500InternalServerError("failed to build token request")
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	tokenReq.Header.Set("Accept", "application/json")

	// Each provider-controlled endpoint picks its own client based on its own
	// URL, so a loopback endpoint never downgrades requests to other hosts.
	tokenResp, err := httputil.ClientForURL(cfg.TokenEndpoint, 15*time.Second).Do(tokenReq)
	if err != nil {
		slog.Error("SSO token exchange failed", "error", err)
		return nil, huma.Error502BadGateway("failed to exchange code for token")
	}
	defer func() { _ = tokenResp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(tokenResp.Body, 1<<20))
	if err != nil {
		slog.ErrorContext(ctx, "failed to read SSO token response", "provider", providerId, "error", err)
		return nil, huma.Error502BadGateway("failed to read token response")
	}
	if tokenResp.StatusCode != http.StatusOK {
		slog.Error("SSO token endpoint error", "status", tokenResp.StatusCode, "body", string(body))
		return nil, huma.Error502BadGateway("token endpoint returned error")
	}

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenData); err != nil {
		return nil, huma.Error502BadGateway("invalid token response")
	}

	if tokenData.IDToken == "" {
		return nil, huma.Error502BadGateway("token response missing id_token")
	}

	if cfg.JWKSEndpoint == "" {
		slog.Error("SSO provider has no JWKS endpoint configured", "provider", providerId)
		return nil, huma.Error500InternalServerError("SSO provider has no JWKS endpoint configured")
	}

	// Verify the ID token's RS256 signature against the provider's JWKS and
	// validate the issuer, audience (client id), and expiry. Also enforces that
	// sub and iat are present.
	jwksClient := httputil.ClientForURL(cfg.JWKSEndpoint, 15*time.Second)
	idToken, claims, err := verifyIDToken(ctx, jwksClient, tokenData.IDToken, cfg.JWKSEndpoint, provider.Issuer, cfg.ClientID)
	if err != nil {
		slog.Error("failed to verify ID token", "error", err, "provider", providerId)
		return nil, huma.Error502BadGateway("failed to verify ID token")
	}

	// Validate the nonce matches the one we issued to prevent token replay.
	// go-oidc surfaces the nonce but does not verify it — that is our job.
	if stateData.Nonce == "" || claims.Nonce != stateData.Nonce {
		slog.Error("SSO ID token nonce mismatch", "provider", providerId)
		return nil, huma.Error502BadGateway("invalid token nonce")
	}

	// Extract user info. sub is guaranteed non-empty by verifyIDToken; it is the
	// stable subject we key the account on.
	sub := idToken.Subject
	email := claims.Email
	name := claims.Name
	picture := claims.Picture

	// Track email_verified from whichever endpoint supplied the email. The
	// claim may be a bool or the string "true"; a present-but-falsey value
	// means the email is NOT verified. Default to "verified" only when the
	// claim is entirely absent (some providers omit it for trusted directories).
	emailVerified := true
	if claims.EmailVerified != nil {
		emailVerified = isTrue(claims.EmailVerified)
	}

	if email == "" {
		// Fall back to the userinfo endpoint for the email/profile, but never
		// for sub. Verify the userinfo subject matches the ID token subject so a
		// compromised/mismatched userinfo response cannot bind a different user.
		userInfo, err := h.fetchUserInfo(ctx, tokenData.AccessToken, provider.Issuer)
		if err != nil {
			return nil, huma.Error502BadGateway("could not get user info from SSO provider")
		}
		if userInfo.Sub == "" || userInfo.Sub != sub {
			slog.Error("SSO userinfo subject mismatch", "provider", providerId)
			return nil, huma.Error502BadGateway("userinfo subject does not match ID token")
		}
		if email == "" {
			email = userInfo.Email
			// The email now comes from userinfo, so its verification status must
			// come from userinfo too — don't trust the (absent) ID token claim.
			if userInfo.EmailVerified != nil {
				emailVerified = isTrue(userInfo.EmailVerified)
			}
		}
		if name == "" {
			name = userInfo.Name
		}
		if picture == "" {
			picture = userInfo.Picture
		}
	}

	if email == "" {
		return nil, huma.Error400BadRequest("SSO provider did not return an email")
	}

	// We treat SSO emails as verified identities and must not trust an
	// unverified address, regardless of whether it came from the ID token or the
	// userinfo endpoint.
	if !emailVerified {
		slog.Error("SSO provider returned unverified email", "provider", providerId,
			"emailVerifiedClaim", fmt.Sprintf("%#v", claims.EmailVerified))
		return nil, huma.Error400BadRequest("SSO provider returned an unverified email")
	}

	if name == "" {
		name = email
	}

	userID, err := h.federated.Provision(ctx, identity.FederatedProfile{
		Provider: "sso-" + providerId, AccountID: fmt.Sprintf("sso-%s:%s", providerId, sub),
		Email: email, EmailVerified: emailVerified, Name: name, ImageURL: picture,
		OrganizationID: provider.OrganizationID,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to provision SSO identity", "provider", providerId, "error", err)
		return nil, huma.Error500InternalServerError("failed to create user")
	}

	// Create a one-time code (not the token itself, for security)
	authCode, err := h.createOneTimeCode(requestFromContext(ctx), userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create session")
	}

	return &authCodeOutput{Body: authCodeResponse{Code: authCode}}, nil
}

type userInfoResponse struct {
	Sub           string      `json:"sub"`
	Email         string      `json:"email"`
	EmailVerified interface{} `json:"email_verified"`
	Name          string      `json:"name"`
	Picture       string      `json:"picture"`
}

func (h *AuthHandler) fetchUserInfo(ctx context.Context, accessToken, issuer string) (*userInfoResponse, error) {
	// Discover userinfo endpoint. The client is chosen per-URL so a loopback
	// issuer never downgrades the (separately-chosen) userinfo request.
	discoveryURL := issuer + "/.well-known/openid-configuration"
	discoveryReq, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httputil.ClientForURL(discoveryURL, 15*time.Second).Do(discoveryReq)
	if err != nil {
		return nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	var discovery struct {
		UserinfoEndpoint string `json:"userinfo_endpoint"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&discovery); err != nil {
		return nil, fmt.Errorf("failed to decode OIDC discovery response: %w", err)
	}

	if discovery.UserinfoEndpoint == "" {
		return nil, fmt.Errorf("no userinfo endpoint found")
	}

	// The userinfo endpoint comes from the (provider-controlled) discovery
	// document, so validate it before use: public HTTPS, or loopback for dev.
	if err := httputil.ValidateSSOEndpoint(discovery.UserinfoEndpoint); err != nil {
		return nil, fmt.Errorf("invalid userinfo endpoint: %w", err)
	}

	// Fetch userinfo using a client selected for the userinfo URL specifically.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	uiResp, err := httputil.ClientForURL(discovery.UserinfoEndpoint, 15*time.Second).Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = uiResp.Body.Close() }()

	var userInfo userInfoResponse
	if err = json.NewDecoder(uiResp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %w", err)
	}
	return &userInfo, nil
}
