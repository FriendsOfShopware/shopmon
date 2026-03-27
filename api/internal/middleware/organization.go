package middleware

import (
	"context"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const OrgIDContextKey contextKey = "orgId"

func GetOrgID(ctx context.Context) string {
	if id, ok := ctx.Value(OrgIDContextKey).(string); ok {
		return id
	}
	return ""
}

func OrganizationMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			orgID := chi.URLParam(r, "orgId")
			if orgID == "" {
				httputil.WriteError(w, http.StatusBadRequest, "organization id required")
				return
			}

			user := GetUser(r.Context())
			if user == nil {
				httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var isMember bool
			err := pool.QueryRow(r.Context(),
				"SELECT COUNT(*) > 0 FROM member WHERE organization_id = $1 AND user_id = $2",
				orgID, user.ID).Scan(&isMember)
			if err != nil || !isMember {
				httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
				return
			}

			ctx := context.WithValue(r.Context(), OrgIDContextKey, orgID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
