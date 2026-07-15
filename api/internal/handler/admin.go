package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	adminread "github.com/friendsofshopware/shopmon/api/internal/readmodel/admin"
)

func (h *Handler) AdminGetStats(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.Stats(r.Context())
	if err != nil {
		h.writeAdminReadError(w, r, "get stats", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetOrganizations(w http.ResponseWriter, r *http.Request, params api.AdminGetOrganizationsParams) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.Organizations(r.Context(), params)
	if err != nil {
		h.writeAdminReadError(w, r, "list organizations", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetEnvironments(w http.ResponseWriter, r *http.Request, params api.AdminGetEnvironmentsParams) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.Environments(r.Context(), params)
	if err != nil {
		h.writeAdminReadError(w, r, "list environments", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetGrowth(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.Growth(r.Context())
	if err != nil {
		h.writeAdminReadError(w, r, "get growth", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetRecentActivity(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.RecentActivity(r.Context())
	if err != nil {
		h.writeAdminReadError(w, r, "get recent activity", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetShopwareVersions(w http.ResponseWriter, r *http.Request) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.ShopwareVersions(r.Context())
	if err != nil {
		h.writeAdminReadError(w, r, "get Shopware versions", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetOrganizationDetail(w http.ResponseWriter, r *http.Request, organizationID string) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.OrganizationDetail(r.Context(), organizationID)
	if err != nil {
		h.writeAdminReadError(w, r, "get organization detail", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetEnvironmentDetail(w http.ResponseWriter, r *http.Request, environmentID int) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.EnvironmentDetail(r.Context(), environmentID)
	if err != nil {
		h.writeAdminReadError(w, r, "get environment detail", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) AdminGetAuditLog(w http.ResponseWriter, r *http.Request, params api.AdminGetAuditLogParams) {
	if !h.requireAdmin(w, r) {
		return
	}
	result, err := h.admin.AuditLog(r.Context(), params)
	if err != nil {
		h.writeAdminReadError(w, r, "get audit log", err)
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

func (h *Handler) requireAdmin(w http.ResponseWriter, r *http.Request) bool {
	user := h.requireUser(w, r)
	if user == nil {
		return false
	}
	if user.Role != "admin" {
		httputil.WriteError(w, http.StatusForbidden, httputil.MsgAdminRequired)
		return false
	}
	return true
}

func (h *Handler) writeAdminReadError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, adminread.ErrOrganizationNotFound):
		httputil.WriteError(w, http.StatusNotFound, "organization not found")
	case errors.Is(err, adminread.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
	default:
		slog.ErrorContext(r.Context(), "admin read failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to load admin data")
	}
}
