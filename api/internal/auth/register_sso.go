package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerSSOAuth(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("signInSSO", http.MethodPost, "/auth/sign-in/sso", "Initiate SSO sign-in", true), h.SignInSSO)
	huma.Register(api, authOperation("ssoCallback", http.MethodGet, "/auth/sso/callback/{providerId}", "Handle SSO OIDC callback", true), h.SsoCallback)
}
