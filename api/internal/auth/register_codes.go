package auth

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerCodes(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("exchangeCode", http.MethodPost, "/auth/exchange-code", "Exchange one-time authorization code for session token", true), h.ExchangeCode)
}
