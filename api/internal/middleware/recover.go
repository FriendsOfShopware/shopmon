package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// Recoverer recovers from panics in downstream handlers. For API requests
// (/api/*) it responds with the standard JSON error shape so clients always
// receive {"message": "..."} instead of chi's plain-text 500. Non-API requests
// (the embedded frontend) fall back to a bare 500 status.
//
// The panic value and stack are logged via slog; they are never written to the
// response body.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			rec := recover()
			if rec == nil {
				return
			}
			// http.ErrAbortHandler is a sentinel used to abort a handler without
			// logging a stack; preserve that behaviour by re-panicking.
			if rec == http.ErrAbortHandler {
				panic(rec)
			}

			slog.ErrorContext(r.Context(), "panic recovered", "panic", rec, "stack", string(debug.Stack()))

			if isAPIPath(r.URL.Path) {
				httputil.WriteError(w, http.StatusInternalServerError, "internal server error")
				return
			}
			w.WriteHeader(http.StatusInternalServerError)
		}()

		next.ServeHTTP(w, r)
	})
}

// isAPIPath reports whether the request targets the JSON API surface.
func isAPIPath(path string) bool {
	return path == "/api" || strings.HasPrefix(path, "/api/")
}
