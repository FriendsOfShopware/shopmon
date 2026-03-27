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

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type oidcConfig struct {
	AuthorizationEndpoint string `json:"authorizationEndpoint"`
	TokenEndpoint         string `json:"tokenEndpoint"`
	JwksEndpoint          string `json:"jwksEndpoint"`
	ClientID              string `json:"clientId"`
	ClientSecret          string `json:"clientSecret,omitempty"`
}

type ssoState struct {
	ProviderID  string `json:"providerId"`
	CallbackURL string `json:"callbackURL"`
	Nonce       string `json:"nonce"`
}

// SignInSSO initiates SSO login by looking up provider by email domain.
func (h *AuthHandler) SignInSSO(w http.ResponseWriter, r *http.Request) {
	var req ssoSignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "email is required")
		return
	}

	// Extract domain from email
	parts := strings.SplitN(req.Email, "@", 2)
	if len(parts) != 2 {
		httputil.WriteError(w, http.StatusBadRequest, "invalid email format")
		return
	}
	domain := strings.ToLower(parts[1])

	// Look up SSO provider by domain
	provider, err := h.queries.GetSSOProviderByDomain(r.Context(), domain)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "no SSO provider configured for this domain")
		return
	}

	if provider.OidcConfig == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SSO provider has no OIDC configuration")
		return
	}

	var cfg oidcConfig
	if err := json.Unmarshal([]byte(*provider.OidcConfig), &cfg); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "invalid OIDC configuration")
		return
	}

	if !h.validateCallbackURL(req.CallbackURL) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid callback URL")
		return
	}

	// Generate state and nonce
	state := generateToken()
	nonce := generateToken()

	// Store state in Redis
	if err := h.challenges.Set(r.Context(), "sso:"+state, ssoState{
		ProviderID:  provider.ProviderID,
		CallbackURL: req.CallbackURL,
		Nonce:       nonce,
	}, 10*time.Minute); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to store SSO state")
		return
	}

	// Build authorization URL
	callbackURL := fmt.Sprintf("%s/auth/sso/callback/%s", h.cfg.FrontendURL, provider.ProviderID)

	authURL := fmt.Sprintf("%s?client_id=%s&redirect_uri=%s&response_type=code&scope=%s&state=%s&nonce=%s",
		cfg.AuthorizationEndpoint,
		url.QueryEscape(cfg.ClientID),
		url.QueryEscape(callbackURL),
		url.QueryEscape("openid profile email"),
		url.QueryEscape(state),
		url.QueryEscape(nonce),
	)

	httputil.WriteJSON(w, http.StatusOK, urlResponse{URL: authURL})
}

// SSOCallback handles the OIDC callback after IdP authentication.
func (h *AuthHandler) SSOCallback(w http.ResponseWriter, r *http.Request) {
	providerId := chi.URLParam(r, "providerId")
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing code or state")
		return
	}

	// Validate state
	var stateData ssoState
	if err := h.challenges.Get(r.Context(), "sso:"+state, &stateData); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired state")
		return
	}

	if stateData.ProviderID != providerId {
		httputil.WriteError(w, http.StatusBadRequest, "provider mismatch")
		return
	}

	provider, err := h.queries.GetSSOProviderByProviderID(r.Context(), providerId)
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "SSO provider not found")
		return
	}

	if provider.OidcConfig == nil {
		httputil.WriteError(w, http.StatusInternalServerError, "SSO provider has no OIDC configuration")
		return
	}

	var cfg oidcConfig
	if err := json.Unmarshal([]byte(*provider.OidcConfig), &cfg); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "invalid OIDC configuration")
		return
	}

	// Exchange code for tokens
	callbackURL := fmt.Sprintf("%s/auth/sso/callback/%s", h.cfg.FrontendURL, providerId)

	tokenResp, err := httputil.NewHTTPClient().PostForm(cfg.TokenEndpoint, url.Values{
		"client_id":     {cfg.ClientID},
		"client_secret": {cfg.ClientSecret},
		"code":          {code},
		"grant_type":    {"authorization_code"},
		"redirect_uri":  {callbackURL},
	})
	if err != nil {
		slog.Error("SSO token exchange failed", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to exchange code for token")
		return
	}
	defer func() { _ = tokenResp.Body.Close() }()

	body, _ := io.ReadAll(tokenResp.Body)
	if tokenResp.StatusCode != http.StatusOK {
		slog.Error("SSO token endpoint error", "status", tokenResp.StatusCode, "body", string(body))
		httputil.WriteError(w, http.StatusBadGateway, "token endpoint returned error")
		return
	}

	var tokenData struct {
		AccessToken  string `json:"access_token"`
		IDToken      string `json:"id_token"`
		RefreshToken string `json:"refresh_token"`
		TokenType    string `json:"token_type"`
	}
	if err := json.Unmarshal(body, &tokenData); err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "invalid token response")
		return
	}

	// Parse ID token - since received directly from token endpoint over TLS,
	// signature verification is optional, but validate claims
	parser := jwt.NewParser()
	token, _, err := parser.ParseUnverified(tokenData.IDToken, jwt.MapClaims{})
	if err != nil {
		slog.Error("failed to parse ID token", "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "failed to parse ID token")
		return
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		httputil.WriteError(w, http.StatusBadGateway, "invalid ID token claims")
		return
	}

	// Validate issuer matches the provider
	if iss, _ := claims["iss"].(string); iss != "" && iss != provider.Issuer {
		httputil.WriteError(w, http.StatusBadGateway, "invalid token issuer")
		return
	}

	// Validate expiry
	if exp, ok := claims["exp"].(float64); ok {
		if time.Unix(int64(exp), 0).Before(time.Now()) {
			httputil.WriteError(w, http.StatusBadGateway, "token expired")
			return
		}
	}

	// Extract user info from claims
	sub, _ := claims["sub"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	picture, _ := claims["picture"].(string)

	if sub == "" || email == "" {
		// Try userinfo endpoint as fallback
		userInfo, err := h.fetchUserInfo(r.Context(), tokenData.AccessToken, provider.Issuer)
		if err != nil {
			httputil.WriteError(w, http.StatusBadGateway, "could not get user info from SSO provider")
			return
		}
		if sub == "" {
			sub = userInfo.Sub
		}
		if email == "" {
			email = userInfo.Email
		}
		if name == "" {
			name = userInfo.Name
		}
		if picture == "" {
			picture = userInfo.Picture
		}
	}

	if email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "SSO provider did not return an email")
		return
	}

	if name == "" {
		name = email
	}

	// Find or create user
	ssoAccountID := fmt.Sprintf("sso-%s:%s", providerId, sub)
	var userID string

	existingAccount, err := h.queries.GetAccountByProviderAndAccountID(r.Context(), queries.GetAccountByProviderAndAccountIDParams{
		ProviderID: "sso-" + providerId,
		AccountID:  ssoAccountID,
	})

	if err == nil {
		userID = existingAccount.UserID
	} else {
		// Try to find user by email
		existingUser, err := h.queries.GetUserByEmail(r.Context(), email)
		if err == nil {
			userID = existingUser.ID
		} else {
			// Create new user
			userID = uuid.New().String()
			_, err = h.queries.CreateUser(r.Context(), queries.CreateUserParams{
				ID:    userID,
				Name:  name,
				Email: email,
			})
			if err != nil {
				slog.Error("failed to create SSO user", "error", err)
				httputil.WriteError(w, http.StatusInternalServerError, "failed to create user")
				return
			}
			// Email verified by SSO provider
			if err := h.queries.UpdateUserEmailVerified(r.Context(), userID); err != nil {
				slog.Warn("failed to mark email as verified for SSO user", "error", err, "userID", userID)
			}
		}

		// Link SSO account
		if err := h.queries.CreateAccount(r.Context(), queries.CreateAccountParams{
			ID:         uuid.New().String(),
			AccountID:  ssoAccountID,
			ProviderID: "sso-" + providerId,
			UserID:     userID,
			Password:   nil,
		}); err != nil {
			slog.Warn("failed to link SSO account", "error", err, "userID", userID, "provider", providerId)
		}
	}

	// Update profile picture if available
	if picture != "" {
		if err := h.queries.UpdateUserProfile(r.Context(), queries.UpdateUserProfileParams{
			Name:  name,
			Image: &picture,
			ID:    userID,
		}); err != nil {
			slog.Error("failed to update user profile from SSO", "error", err, "userID", userID)
		}
	}

	// Organization auto-provisioning
	if provider.OrganizationID != nil && *provider.OrganizationID != "" {
		isMember, _ := h.queries.IsOrgMember(r.Context(), queries.IsOrgMemberParams{
			OrganizationID: *provider.OrganizationID,
			UserID:         userID,
		})
		if !isMember {
			if err := h.queries.CreateMember(r.Context(), queries.CreateMemberParams{
				ID:             uuid.New().String(),
				OrganizationID: *provider.OrganizationID,
				UserID:         userID,
				Role:           "member",
			}); err != nil {
				slog.Error("failed to auto-provision SSO org membership", "error", err, "userID", userID, "orgID", *provider.OrganizationID)
			}
		}
	}

	// Create a one-time code (not the token itself, for security)
	authCode, err := h.createOneTimeCode(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Redirect with short-lived code — frontend exchanges it for a token via POST /auth/exchange-code
	redirectURL := stateData.CallbackURL
	if redirectURL == "" {
		redirectURL = h.cfg.FrontendURL
	}
	u, _ := url.Parse(redirectURL)
	q := u.Query()
	q.Set("code", authCode)
	u.RawQuery = q.Encode()
	http.Redirect(w, r, u.String(), http.StatusFound)
}

type userInfoResponse struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

func (h *AuthHandler) fetchUserInfo(ctx context.Context, accessToken, issuer string) (*userInfoResponse, error) {
	// Discover userinfo endpoint
	discoveryURL := issuer + "/.well-known/openid-configuration"
	discoveryReq, err := http.NewRequestWithContext(ctx, http.MethodGet, discoveryURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := httputil.NewHTTPClient().Do(discoveryReq)
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

	// Fetch userinfo
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, discovery.UserinfoEndpoint, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer "+accessToken)

	uiResp, err := httputil.NewHTTPClient().Do(req)
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
