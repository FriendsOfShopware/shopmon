package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/authapi"
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

func (h *AuthHandler) SignInSocial(w http.ResponseWriter, r *http.Request) {
	var req socialSignInRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Provider != "github" {
		httputil.WriteError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	if h.config.GitHubClientID == "" {
		httputil.WriteError(w, http.StatusNotFound, "GitHub OAuth not configured")
		return
	}

	if !h.validateCallbackURL(req.CallbackURL) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid callback URL")
		return
	}

	state, err := generateToken()
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to generate OAuth state", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate OAuth state")
		return
	}
	if err := h.challenges.Set(r.Context(), "oauth:"+state, oauthState{
		CallbackURL: req.CallbackURL,
	}, 10*time.Minute); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to store OAuth state")
		return
	}

	authURL := fmt.Sprintf("%s/login/oauth/authorize?client_id=%s&state=%s&scope=user:email",
		githubOAuthBaseURL, url.QueryEscape(h.config.GitHubClientID), url.QueryEscape(state))

	httputil.WriteJSON(w, http.StatusOK, urlResponse{URL: authURL})
}

func (h *AuthHandler) GithubCallback(w http.ResponseWriter, r *http.Request, params authapi.GithubCallbackParams) {
	// Validate state (Get consumes the key)
	var stateData oauthState
	if err := h.challenges.Get(r.Context(), "oauth:"+params.State, &stateData); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired state")
		return
	}

	// Exchange code for access token
	tokenForm := url.Values{
		"client_id":     {h.config.GitHubClientID},
		"client_secret": {h.config.GitHubClientSecret},
		"code":          {params.Code},
	}
	tokenReq, err := http.NewRequestWithContext(r.Context(), http.MethodPost,
		githubOAuthBaseURL+"/login/oauth/access_token", strings.NewReader(tokenForm.Encode()))
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build token request")
		return
	}
	tokenReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	tokenResp, err := httputil.NewHTTPClient().Do(tokenReq)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to exchange code")
		return
	}
	defer func() { _ = tokenResp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(tokenResp.Body, 1<<20))
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to read token response")
		return
	}
	tokenParams, err := url.ParseQuery(string(body))
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to parse token response")
		return
	}
	accessToken := tokenParams.Get("access_token")
	if accessToken == "" {
		httputil.WriteError(w, http.StatusBadGateway, "failed to get access token")
		return
	}

	// Fetch GitHub user info
	ghReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, githubAPIBaseURL+"/user", nil)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to build GitHub user request")
		return
	}
	ghReq.Header.Set("Authorization", "Bearer "+accessToken)
	ghReq.Header.Set("Accept", "application/json")

	ghResp, err := httputil.NewHTTPClient().Do(ghReq)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch GitHub user")
		return
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
		httputil.WriteError(w, http.StatusBadGateway, "failed to decode GitHub user response")
		return
	}

	// If no public email, fetch from emails API
	if ghUser.Email == "" {
		emailReq, err := http.NewRequestWithContext(r.Context(), http.MethodGet, githubAPIBaseURL+"/user/emails", nil)
		if err != nil {
			httputil.WriteError(w, http.StatusInternalServerError, "failed to build GitHub email request")
			return
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
			slog.ErrorContext(r.Context(), "failed to fetch GitHub emails", "error", err)
		}
	}

	if ghUser.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "could not get email from GitHub")
		return
	}

	if ghUser.Name == "" {
		ghUser.Name = ghUser.Login
	}

	userID, err := h.federated.Provision(r.Context(), identity.FederatedProfile{
		Provider: "github", AccountID: fmt.Sprintf("%d", ghUser.ID),
		Email: ghUser.Email, EmailVerified: true, Name: ghUser.Name, ImageURL: ghUser.AvatarURL,
	})
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to provision GitHub identity", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create user")
		return
	}

	// Create a one-time code (not the token itself, for security)
	authCode, err := h.createOneTimeCode(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"code": authCode})
}
