// Package apirouter wires the generated API and auth handlers onto an /api
// sub-router with consistent JSON error handling. It is the single source of
// truth shared by the production server (server.go) and the test harness
// (internal/testutil) so the two cannot drift apart.
package apirouter

import (
	"log/slog"
	"net/http"

	apiserver "github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/go-chi/chi/v5"
)

// Options configures Mount.
type Options struct {
	// AuthMiddlewares are applied to the generated auth routes (e.g. rate
	// limiting). May be empty.
	AuthMiddlewares []authapi.MiddlewareFunc
}

// Mount registers the auth and API handlers on apiRouter (which must already be
// scoped to /api) and installs the shared JSON error responders: a JSON 400 for
// request-parameter binding failures, and JSON 404 / 405 for unmatched routes
// and methods. Every error leaving the /api surface therefore uses the standard
// {"message": "..."} shape.
func Mount(apiRouter chi.Router, api apiserver.ServerInterface, authHandler authapi.ServerInterface, opts Options) {
	// paramErrorHandler converts oapi-codegen's request-binding failures
	// (missing/invalid path or query params) into the standard JSON error shape.
	// The default generated handler writes plain text via http.Error and echoes
	// the raw error; we keep the 400 status but return JSON and a generic message
	// so the contract stays consistent and details aren't leaked.
	paramErrorHandler := func(w http.ResponseWriter, r *http.Request, err error) {
		slog.DebugContext(r.Context(), "request parameter binding failed", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "invalid request parameters")
	}

	authapi.HandlerWithOptions(authHandler, authapi.ChiServerOptions{
		BaseRouter:       apiRouter,
		Middlewares:      opts.AuthMiddlewares,
		ErrorHandlerFunc: paramErrorHandler,
	})

	apiserver.HandlerWithOptions(api, apiserver.ChiServerOptions{
		BaseRouter:       apiRouter,
		ErrorHandlerFunc: paramErrorHandler,
	})

	// Unmatched /api/* routes return a JSON 404 instead of falling through to
	// the SPA fallback (which would serve HTML for a missing endpoint).
	apiRouter.NotFound(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteError(w, http.StatusNotFound, "not found")
	})
	apiRouter.MethodNotAllowed(func(w http.ResponseWriter, r *http.Request) {
		httputil.WriteError(w, http.StatusMethodNotAllowed, "method not allowed")
	})
}
