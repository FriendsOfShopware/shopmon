package handler

import (
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetHealth handles the health check endpoint.
func (h *Handler) GetHealth(w http.ResponseWriter, r *http.Request) {
	// Check database connection
	if err := h.pool.Ping(r.Context()); err != nil {
		httputil.WriteError(w, http.StatusServiceUnavailable, "database connection failed")
		return
	}

	// Respond with text/plain "ok" to match the OpenAPI spec.
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}
