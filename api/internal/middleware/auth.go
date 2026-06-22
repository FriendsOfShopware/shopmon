package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

type contextKey string

const (
	UserContextKey    contextKey = "user"
	SessionContextKey contextKey = "session"
)

func GetUser(ctx context.Context) *auth.User {
	if u, ok := ctx.Value(UserContextKey).(*auth.User); ok {
		return u
	}
	return nil
}

func GetSession(ctx context.Context) *auth.Session {
	if s, ok := ctx.Value(SessionContextKey).(*auth.Session); ok {
		return s
	}
	return nil
}

// OptionalAuthMiddleware sets user context if a valid token is present,
// but does not block requests without authentication.
func OptionalAuthMiddleware(q *queries.Queries) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := httputil.ExtractToken(r)
			if token == "" {
				next.ServeHTTP(w, r)
				return
			}

			su, err := auth.ValidateSession(r.Context(), q, token)
			if err != nil {
				slog.DebugContext(r.Context(), "optional auth: session validation failed", "error", err)
				next.ServeHTTP(w, r)
				return
			}

			if su.User.Banned != nil && *su.User.Banned {
				httputil.WriteError(w, http.StatusForbidden, "banned")
				return
			}

			ctx := context.WithValue(r.Context(), UserContextKey, &su.User)
			ctx = context.WithValue(ctx, SessionContextKey, &su.Session)
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
		if GetUser(r.Context()) == nil {
			httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}
