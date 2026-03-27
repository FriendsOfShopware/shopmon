package middleware

import (
	"context"
	"net/http"
	"strconv"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const ShopIDContextKey contextKey = "shopId"

func GetShopID(ctx context.Context) int {
	if id, ok := ctx.Value(ShopIDContextKey).(int); ok {
		return id
	}
	return 0
}

func ShopMiddleware(pool *pgxpool.Pool) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			shopIDStr := chi.URLParam(r, "shopId")
			shopID, err := strconv.Atoi(shopIDStr)
			if err != nil {
				httputil.WriteError(w, http.StatusBadRequest, "invalid shop id")
				return
			}

			user := GetUser(r.Context())
			if user == nil {
				httputil.WriteError(w, http.StatusUnauthorized, "unauthorized")
				return
			}

			var orgID string
			err = pool.QueryRow(r.Context(),
				"SELECT organization_id FROM shop WHERE id = $1", shopID).Scan(&orgID)
			if err != nil {
				httputil.WriteError(w, http.StatusNotFound, "shop not found")
				return
			}

			var isMember bool
			err = pool.QueryRow(r.Context(),
				"SELECT COUNT(*) > 0 FROM member WHERE organization_id = $1 AND user_id = $2",
				orgID, user.ID).Scan(&isMember)
			if err != nil || !isMember {
				httputil.WriteError(w, http.StatusForbidden, "no access to this shop")
				return
			}

			ctx := context.WithValue(r.Context(), ShopIDContextKey, shopID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
