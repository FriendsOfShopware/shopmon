package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const EnvironmentIDContextKey contextKey = "environmentId"

func GetEnvironmentID(ctx context.Context) int {
	if id, ok := ctx.Value(EnvironmentIDContextKey).(int); ok {
		return id
	}
	return 0
}

func EnvironmentMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			environmentIDStr := chi.URLParam(r, "environmentId")
			environmentID, err := strconv.Atoi(environmentIDStr)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "invalid environment id")
				return
			}

			user := GetUser(r.Context())
			if user == nil {
				httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var orgID string
			err = pool.QueryRow(r.Context(),
				"SELECT organization_id FROM environment WHERE id = $1", environmentID).Scan(&orgID)
			if err != nil {
				httputil.WriteError(w, http.StatusNotFound, "environment not found")
				return
			}

			var isMember bool
			err = pool.QueryRow(r.Context(),
				"SELECT COUNT(*) > 0 FROM member WHERE organization_id = $1 AND user_id = $2",
				orgID, user.ID).Scan(&isMember)
			if err != nil || !isMember {
				httputil.WriteError(w, http.StatusForbidden, "no access to this environment")
				return
			}

			ctx := context.WithValue(r.Context(), EnvironmentIDContextKey, environmentID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
