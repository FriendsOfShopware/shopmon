// Package apirouter wires the Huma API onto an /api sub-router with consistent
// JSON error handling. It is the single source of truth shared by the
// production server (server.go) and the test harness (internal/testutil) so the
// two cannot drift apart.
package apirouter

import (
	"net/http"
	"reflect"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
)

func init() {
	// oapi-codegen treated Go slices as required non-null arrays. Huma's
	// default marks every slice nullable, which makes generated frontend
	// types `T[] | null` and breaks existing callers.
	huma.DefaultArrayNullable = false
}

// schemaNameAliases map Huma's Go-type-derived names onto the historical
// OpenAPI schema names so generated frontend types stay stable.
var schemaNameAliases = map[string]string{
	"AdminUserResponse":         "AdminUser",
	"AdminListUsersOutputBody":  "AdminUsersResponse",
	"AuthAdminUserDetail":       "AdminUserDetail",
	"AuthAdminUserAuthProvider": "AdminUserAuthProvider",
	"AuthAdminUserMembership":   "AdminUserMembership",
	"AuthAdminUserSession":      "AdminUserSession",
	"JSONStatusError":           "ErrorResponse",
}

func schemaNamer(t reflect.Type, hint string) string {
	name := huma.DefaultSchemaNamer(t, hint)
	if alias, ok := schemaNameAliases[name]; ok {
		return alias
	}
	return name
}

// Options configures Mount.
type Options struct {
	// AuthRateLimit is applied to paths under /auth (e.g. /api/auth/...).
	// May be nil.
	AuthRateLimit func(http.Handler) http.Handler
}

// NewAPI constructs the Huma API on apiRouter (already scoped to /api) with
// Shopmon OpenAPI metadata. Operations are not registered.
func NewAPI(apiRouter chi.Router) huma.API {
	config := huma.DefaultConfig("Shopmon API", "1.0.0")
	config.CreateHooks = nil
	config.Transformers = nil
	config.SchemasPath = ""
	config.OpenAPIPath = "/openapi"
	config.DocsPath = "/docs"
	config.DocsRenderer = huma.DocsRendererStoplightElements
	if config.Info != nil {
		config.Info.Description = "Shopware monitoring application API"
	}
	config.Servers = []*huma.Server{{
		URL:         "/api",
		Description: "API base path",
	}}
	if config.Components == nil {
		config.Components = &huma.Components{}
	}
	config.Components.Schemas = huma.NewMapRegistry("#/components/schemas/", schemaNamer)
	config.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"bearerAuth": {
			Type:        "http",
			Scheme:      "bearer",
			Description: "Bearer token authentication",
		},
	}
	config.Security = []map[string][]string{
		{"bearerAuth": {}},
	}

	return humachi.New(apiRouter, config)
}

// Mount registers auth and API operations on apiRouter (already scoped to /api)
// and installs JSON 404 / 405 responders so every error leaving the /api
// surface uses the standard {"message": "..."} shape.
func Mount(apiRouter chi.Router, apiHandler *handler.Handler, authHandler *auth.AuthHandler, opts Options) huma.API {
	if opts.AuthRateLimit != nil {
		apiRouter.Use(authPathMiddleware(opts.AuthRateLimit))
	}

	api := NewAPI(apiRouter)
	handler.Register(api, apiHandler)
	auth.Register(api, authHandler)

	apiRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
	apiRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	})
	return api
}

func authPathMiddleware(limit func(http.Handler) http.Handler) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		limited := limit(next)
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if isAuthPath(r.URL.Path) {
				limited.ServeHTTP(w, r)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func isAuthPath(path string) bool {
	return path == "/auth" || strings.HasPrefix(path, "/auth/") ||
		path == "/api/auth" || strings.HasPrefix(path, "/api/auth/")
}
