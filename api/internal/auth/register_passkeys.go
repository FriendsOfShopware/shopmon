package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerPasskeys(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("passkeyRegisterOptions", http.MethodPost, "/auth/passkey/register-options", "Begin passkey registration", false), h.PasskeyRegisterOptions)
	huma.Register(api, authOperation("passkeyRegister", http.MethodPost, "/auth/passkey/register", "Complete passkey registration", false), h.PasskeyRegister)
	huma.Register(api, authOperation("passkeyLoginOptions", http.MethodPost, "/auth/passkey/login-options", "Begin passkey login", true), h.PasskeyLoginOptions)
	huma.Register(api, authOperation("passkeyLogin", http.MethodPost, "/auth/passkey/login", "Complete passkey login", true), h.PasskeyLogin)
}
