package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
)

// GetOrganizationShops returns all shops in an organization.
func (h *Handler) GetOrganizationShops(w http.ResponseWriter, r *http.Request, orgID api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	result, err := h.environments.OrganizationShops(r.Context(), user.ID, orgID)
	if errors.Is(err, environmentread.ErrNotAuthorized) {
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	}
	if err != nil {
		slog.ErrorContext(r.Context(), "failed to list shops", "organizationID", orgID, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shops")
		return
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateShop creates a new shop with its first environment.
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.CreateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" || req.EnvironmentName == "" || req.EnvironmentUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, environmentName, environmentUrl, clientId, and clientSecret are required")
		return
	}

	result, err := h.monitoring.CreateShop(r.Context(), monitoring.CreateShopCommand{
		UserID:          user.ID,
		OrganizationID:  orgId,
		Name:            req.Name,
		Description:     req.Description,
		GitURL:          req.GitUrl,
		EnvironmentName: req.EnvironmentName,
		EnvironmentURL:  req.EnvironmentUrl,
		ClientID:        req.ClientId,
		ClientSecret:    req.ClientSecret,
	})
	if errors.Is(err, monitoring.ErrNotAuthorized) {
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	}
	if errors.Is(err, monitoring.ErrConnectionFailed) {
		slog.Error("failed to validate shop connection for new shop", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	}
	if err != nil {
		slog.Error("failed to create shop with environment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create shop")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, api.Shop{
		Id:                   int(result.ShopID),
		Name:                 req.Name,
		Description:          req.Description,
		GitUrl:               req.GitUrl,
		OrganizationId:       orgId,
		DefaultEnvironmentId: ptr.Int(&result.EnvironmentID),
	})
}

// UpdateShop updates an existing shop.
func (h *Handler) UpdateShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var req api.UpdateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	var defaultEnvironmentID *int32
	if req.DefaultEnvironmentId != nil {
		value := int32(*req.DefaultEnvironmentId)
		defaultEnvironmentID = &value
	}
	err := h.monitoring.UpdateShop(r.Context(), monitoring.UpdateShopCommand{
		UserID:               user.ID,
		OrganizationID:       orgId,
		ShopID:               int32(shopId),
		Name:                 req.Name,
		Description:          req.Description,
		GitURL:               req.GitUrl,
		DefaultEnvironmentID: defaultEnvironmentID,
	})
	switch {
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrShopNotFound):
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
		return
	case errors.Is(err, monitoring.ErrDefaultEnvironmentNotFound):
		httputil.WriteError(w, http.StatusBadRequest, "default environment not found")
		return
	case errors.Is(err, monitoring.ErrDefaultEnvironmentMismatch):
		httputil.WriteError(w, http.StatusForbidden, "default environment does not belong to this shop")
		return
	case err != nil:
		slog.Error("failed to update shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShop deletes a shop.
func (h *Handler) DeleteShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	err := h.monitoring.DeleteShop(r.Context(), user.ID, orgId, int32(shopId))
	switch {
	case errors.Is(err, monitoring.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
		return
	case errors.Is(err, monitoring.ErrShopNotFound):
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
		return
	case errors.Is(err, monitoring.ErrShopHasEnvironments):
		httputil.WriteError(w, http.StatusConflict, "cannot delete shop with existing environments")
		return
	case err != nil:
		slog.Error("failed to delete shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
