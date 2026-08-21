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

type oauthState struct {
	CallbackURL string `json:"callbackURL"`
}

// GitHub OAuth/API base URLs. Declared as vars (not consts) so tests can point
// the flow at a mock server; production always uses the real GitHub endpoints.
var (
	githubOAuthBaseURL = "https://github.com"
	githubAPIBaseURL   = "https://api.github.com"
)

func (h *AuthHandler) SignInSocial(ctx context.Context, input *socialSignInInput) (*urlOutput, error) {
	req := input.Body

	if req.Provider != "github" {
		return nil, huma.Error400BadRequest("unsupported provider")
	}

	if h.config.GitHubClientID == "" {
		return nil, huma.Error404NotFound("GitHub OAuth not configured")
	}

	if !h.validateCallbackURL(req.CallbackURL) {
		return nil, huma.Error400BadRequest("invalid callback URL")
	}

	state, err := generateToken()
	if err != nil {
		slog.ErrorContext(ctx, "failed to generate OAuth state", "error", err)
		return nil, huma.Error500InternalServerError("failed to generate OAuth state")
	}
	if err := h.challenges.Set(ctx, "oauth:"+state, oauthState{
		CallbackURL: req.CallbackURL,
	}, 10*time.Minute); err != nil {
		return nil, huma.Error500InternalServerError("failed to store OAuth state")
	}

	authURL := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&state=%s&scope=user:email",
		githubOAuthBaseURL, url.QueryEscape(h.config.GitHubClientID), url.QueryEscape(state))

	return &urlOutput{Body: urlResponse{URL: authURL}}, nil
}

func (h *AuthHandler) GithubCallback(ctx context.Context, input *githubCallbackInput) (*authCodeOutput, error) {
	if input.Code == "" || input.State == "" {
		return nil, huma.Error400BadRequest("missing or invalid parameters")
	}

	// Validate state (Get consumes the key)
	var stateData oauthState
	if err := h.challenges.Get(ctx, "oauth:"+input.State, &stateData); err != nil {
		return nil, huma.Error400BadRequest("invalid or expired state")
	}

	// Exchange code for access token
	tokenForm := url.Values{
		"client_id":     {h.config.GitHubClientID},
		"client_secret": {h.config.GitHubClientSecret},
		"code":          {input.Code},
	}
	tokenReq, err := http.NewRequestWithContext(ctx, http.MethodPost,
		githubOAuthBaseURL+"/login/oauth/access_token", strings.NewReader(tokenForm.Encode()))
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to build token request")
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := httputil.NewHTTPClient().Do(tokenReq)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to exchange code")
	}
	defer func() { _ = tokenResp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(tokenResp.Body, 1<<20))
	if err != nil {
		return nil, huma.Error502BadGateway("failed to read token response")
	}
	tokenParams, err := url.ParseQuery(string(body))
	if err != nil {
		return nil, huma.Error502BadGateway("failed to parse token response")
	}
	accessToken := tokenParams.Get("access_token")
	if accessToken == "" {
		return nil, huma.Error502BadGateway("failed to get access token")
	}

	// Fetch GitHub user info
	ghReq, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBaseURL+"/user", nil)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to build GitHub user request")
	}
	ghReq.Header.Set("Authorization", "Bearer "+accessToken)
	ghReq.Header.Set("Accept", "application/json")

	ghResp, err := httputil.NewHTTPClient().Do(ghReq)
	if err != nil {
		return nil, huma.Error502BadGateway("failed to fetch GitHub user")
	}
	defer func() { _ = ghResp.Body.Close() }()

	var ghUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	if err := json.NewDecoder(ghResp.Body).Decode(&ghUser); err != nil {
		return nil, huma.Error502BadGateway("failed to decode GitHub user response")
	}

	// If no public email, fetch from emails API
	if ghUser.Email == "" {
		emailReq, err := http.NewRequestWithContext(ctx, http.MethodGet, githubAPIBaseURL+"/user/emails", nil)
		if err != nil {
			return nil, huma.Error500InternalServerError("failed to build GitHub email request")
		}
		emailReq.Header.Set("Authorization", "Bearer "+accessToken)
		emailReq.Header.Set("Accept", "application/json")

		emailResp, err := httputil.NewHTTPClient().Do(emailReq)
		if err == nil {
			defer func() { _ = emailResp.Body.Close() }()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			if err := json.NewDecoder(emailResp.Body).Decode(&emails); err != nil {
				slog.Error("failed to decode GitHub emails response", "error", err)
			}
			for _, e := range emails {
				if e.Primary && e.Verified {
					ghUser.Email = e.Email
					break
				}
			}
		} else {
			slog.ErrorContext(ctx, "failed to fetch GitHub emails", "error", err)
		}
	}

	if ghUser.Email == "" {
		return nil, huma.Error400BadRequest("could not get email from GitHub")
	}

	if ghUser.Name == "" {
		ghUser.Name = ghUser.Login
	}

	userID, err := h.federated.Provision(ctx, identity.FederatedProfile{
		Provider: "github", AccountID: fmt.Sprintf("%d", ghUser.ID),
		Email: ghUser.Email, EmailVerified: true, Name: ghUser.Name, ImageURL: ghUser.AvatarURL,
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to provision GitHub identity", "error", err)
		return nil, huma.Error500InternalServerError("failed to create user")
	}

	// Create a one-time code (not the token itself, for security)
	authCode, err := h.createOneTimeCode(requestFromContext(ctx), userID)
	if err != nil {
		return nil, huma.Error500InternalServerError("failed to create session")
	}

	return &authCodeOutput{Body: authCodeResponse{Code: authCode}}, nil
}
