package auth

import (
	"context"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
)

type httpRequestContextKey struct{}

func attachHTTPRequest(ctx huma.Context, next func(huma.Context)) {
	r, _ := humachi.Unwrap(ctx)
	next(huma.WithValue(ctx, httpRequestContextKey{}, r))
}

func requestFromContext(ctx context.Context) *http.Request {
	if r, ok := ctx.Value(httpRequestContextKey{}).(*http.Request); ok && r != nil {
		return r
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
	if err != nil {
		return &http.Request{}
	}
	return req
}

func authOperation(id, method, path, summary string, public bool) huma.Operation {
	op := huma.Operation{
		OperationID:      id,
		Method:           method,
		Path:             path,
		Summary:          summary,
		Tags:             []string{"Auth"},
		Middlewares:      huma.Middlewares{attachHTTPRequest},
		SkipValidateBody: true,
	}
	if public {
		op.Security = []map[string][]string{}
	}
	return op
}

func registerCredentials(api huma.API, h *AuthHandler) {
	huma.Register(api, authOperation("signUpEmail", http.MethodPost, "/auth/sign-up/email", "Register with email and password", true), h.SignUpEmail)
	huma.Register(api, authOperation("signInEmail", http.MethodPost, "/auth/sign-in/email", "Sign in with email and password", true), h.SignInEmail)
	huma.Register(api, authOperation("signOut", http.MethodPost, "/auth/sign-out", "Sign out", false), h.SignOut)
	huma.Register(api, authOperation("getSession", http.MethodGet, "/auth/session", "Get current session", false), h.GetSession)
	huma.Register(api, authOperation("verifyEmail", http.MethodGet, "/auth/verify-email", "Verify email address", true), h.VerifyEmail)
	huma.Register(api, authOperation("forgetPassword", http.MethodPost, "/auth/forget-password", "Request password reset email", true), h.ForgetPassword)
	huma.Register(api, authOperation("resetPassword", http.MethodPost, "/auth/reset-password", "Reset password with token", true), h.ResetPassword)
}
