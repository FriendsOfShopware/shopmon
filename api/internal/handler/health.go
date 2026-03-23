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

	httputil.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
