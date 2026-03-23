package auth

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/google/uuid"
)

type oauthState struct {
	CallbackURL string `json:"callbackURL"`
}

func (h *AuthHandler) SignInSocial(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Provider    string `json:"provider"`
		CallbackURL string `json:"callbackURL"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Provider != "github" {
		httputil.WriteError(w, http.StatusBadRequest, "unsupported provider")
		return
	}

	if h.cfg.GithubClientID == "" {
		httputil.WriteError(w, http.StatusNotFound, "GitHub OAuth not configured")
		return
	}

	if !h.validateCallbackURL(req.CallbackURL) {
		httputil.WriteError(w, http.StatusBadRequest, "invalid callback URL")
		return
	}

	state := generateToken()
	h.challenges.Set(r.Context(), "oauth:"+state, oauthState{
		CallbackURL: req.CallbackURL,
	}, 10*time.Minute)

	authURL := fmt.Sprintf("https://github.com/login/oauth/authorize?client_id=%s&state=%s&scope=user:email",
		url.QueryEscape(h.cfg.GithubClientID), url.QueryEscape(state))

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"url": authURL})
}

func (h *AuthHandler) GithubCallback(w http.ResponseWriter, r *http.Request) {
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" || state == "" {
		httputil.WriteError(w, http.StatusBadRequest, "missing code or state")
		return
	}

	// Validate state (Get consumes the key)
	var stateData oauthState
	if err := h.challenges.Get(r.Context(), "oauth:"+state, &stateData); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid or expired state")
		return
	}

	// Exchange code for access token
	tokenResp, err := httputil.NewHTTPClient().PostForm("https://github.com/login/oauth/access_token", url.Values{
		"client_id":     {h.cfg.GithubClientID},
		"client_secret": {h.cfg.GithubClientSecret},
		"code":          {code},
	})
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to exchange code")
		return
	}
	defer tokenResp.Body.Close()

	body, _ := io.ReadAll(tokenResp.Body)
	params, _ := url.ParseQuery(string(body))
	accessToken := params.Get("access_token")
	if accessToken == "" {
		httputil.WriteError(w, http.StatusBadGateway, "failed to get access token")
		return
	}

	// Fetch GitHub user info
	ghReq, _ := http.NewRequest("GET", "https://api.github.com/user", nil)
	ghReq.Header.Set("Authorization", "Bearer "+accessToken)
	ghReq.Header.Set("Accept", "application/json")

	ghResp, err := httputil.NewHTTPClient().Do(ghReq)
	if err != nil {
		httputil.WriteError(w, http.StatusBadGateway, "failed to fetch GitHub user")
		return
	}
	defer ghResp.Body.Close()

	var ghUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Name      string `json:"name"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}
	json.NewDecoder(ghResp.Body).Decode(&ghUser)

	// If no public email, fetch from emails API
	if ghUser.Email == "" {
		emailReq, _ := http.NewRequest("GET", "https://api.github.com/user/emails", nil)
		emailReq.Header.Set("Authorization", "Bearer "+accessToken)
		emailReq.Header.Set("Accept", "application/json")

		emailResp, err := httputil.NewHTTPClient().Do(emailReq)
		if err == nil {
			defer emailResp.Body.Close()
			var emails []struct {
				Email    string `json:"email"`
				Primary  bool   `json:"primary"`
				Verified bool   `json:"verified"`
			}
			json.NewDecoder(emailResp.Body).Decode(&emails)
			for _, e := range emails {
				if e.Primary && e.Verified {
					ghUser.Email = e.Email
					break
				}
			}
		}
	}

	if ghUser.Email == "" {
		httputil.WriteError(w, http.StatusBadRequest, "could not get email from GitHub")
		return
	}

	if ghUser.Name == "" {
		ghUser.Name = ghUser.Login
	}

	githubAccountID := fmt.Sprintf("%d", ghUser.ID)

	existingAccount, err := h.queries.GetAccountByProviderAndAccountID(r.Context(), queries.GetAccountByProviderAndAccountIDParams{
		ProviderID: "github",
		AccountID:  githubAccountID,
	})

	var userID string

	if err == nil {
		// Account exists, use existing user
		userID = existingAccount.UserID
	} else {
		// Try to find user by email for account linking
		existingUser, err := h.queries.GetUserByEmail(r.Context(), ghUser.Email)
		if err == nil {
			// Link GitHub account to existing user
			userID = existingUser.ID
		} else {
			// Create new user
			userID = uuid.New().String()
			_, err = h.queries.CreateUser(r.Context(), queries.CreateUserParams{
				ID:    userID,
				Name:  ghUser.Name,
				Email: ghUser.Email,
			})
			if err != nil {
				slog.Error("failed to create user from GitHub", "error", err)
				httputil.WriteError(w, http.StatusInternalServerError, "failed to create user")
				return
			}
			// Mark email as verified (GitHub verified it)
			if err := h.queries.UpdateUserEmailVerified(r.Context(), userID); err != nil {
				slog.Warn("failed to mark email as verified for GitHub user", "error", err, "userID", userID)
			}
		}

		// Create GitHub account link
		if err := h.queries.CreateAccount(r.Context(), queries.CreateAccountParams{
			ID:         uuid.New().String(),
			AccountID:  githubAccountID,
			ProviderID: "github",
			UserID:     userID,
			Password:   nil,
		}); err != nil {
			slog.Warn("failed to link GitHub account", "error", err, "userID", userID)
		}
	}

	// Update user avatar if not set
	if ghUser.AvatarURL != "" {
		h.queries.UpdateUserProfile(r.Context(), queries.UpdateUserProfileParams{
			Name:  ghUser.Name,
			Image: &ghUser.AvatarURL,
			ID:    userID,
		})
	}

	// Create a one-time code (not the token itself, for security)
	authCode, err := h.createOneTimeCode(r, userID)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create session")
		return
	}

	// Redirect with short-lived code — frontend exchanges it for a token
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
