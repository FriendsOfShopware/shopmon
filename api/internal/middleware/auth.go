package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// SessionAuthenticator is the application boundary used by HTTP middleware.
type SessionAuthenticator interface {
	Authenticate(ctx context.Context, token string) (*access.Principal, error)
}

// OptionalAuthMiddleware sets the request principal if a valid token is present,
// but does not block requests without authentication.
func OptionalAuthMiddleware(authenticator SessionAuthenticator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := httputil.ExtractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			principal, err := authenticator.Authenticate(r.Context(), token)
			if err != nil {
				slog.DebugContext(r.Context(), "optional auth: session validation failed", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			ctx := access.WithPrincipal(r.Context(), principal)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAuth rejects requests with 401 when no authenticated user is present
// in the request context. It complements OptionalAuthMiddleware: place
// OptionalAuthMiddleware first to populate the user (and session) context, then
// RequireAuth on a route group to declaratively enforce that authentication is
// mandatory for those routes, removing the need for per-handler requireUser
// checks.
func RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if access.PrincipalFromContext(r.Context()) == nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
