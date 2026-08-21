package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerGithub(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("signInSocial", http.MethodPost, "/auth/sign-in/social", "Initiate OAuth sign-in (GitHub)", true), h.SignInSocial)
	huma.Register(api, authOperation("githubCallback", http.MethodGet, "/auth/callback/github", "Handle GitHub OAuth callback", true), h.GithubCallback)
}
