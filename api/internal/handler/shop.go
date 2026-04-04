package handler

import (
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetOrganizationShops returns all shops in the active organization.
func (h *Handler) GetOrganizationShops(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	rows, err := h.queries.ListShopsByOrganization(r.Context(), orgId)
	if err != nil {
		slog.Error("failed to list shops", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get shops")
		return
	}

	result := make([]api.Shop, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.Shop{
			Id:             int(row.ID),
			Name:           row.Name,
			Description:    row.Description,
			GitUrl:         row.GitUrl,
			OrganizationId: row.OrganizationID,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateShop creates a new shop in the active organization.
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	var req api.CreateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	shopID, err := h.queries.CreateShop(r.Context(), queries.CreateShopParams{
		OrganizationID: orgId,
		Name:           req.Name,
		Description:    req.Description,
		GitUrl:         req.GitUrl,
	})
	if err != nil {
		slog.Error("failed to create shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create shop")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, api.Shop{
		Id:             int(shopID),
		Name:           req.Name,
		Description:    req.Description,
		GitUrl:         req.GitUrl,
		OrganizationId: orgId,
	})
}

// UpdateShop updates an existing shop.
func (h *Handler) UpdateShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	shop, err := h.queries.GetShopByID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if shop.OrganizationID != orgId {
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
		return
	}

	var req api.UpdateShopRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	name := shop.Name
	if req.Name != nil {
		name = *req.Name
	}
	description := shop.Description
	if req.Description != nil {
		description = req.Description
	}
	gitUrl := shop.GitUrl
	if req.GitUrl != nil {
		gitUrl = req.GitUrl
	}

	if err := h.queries.UpdateShop(r.Context(), queries.UpdateShopParams{
		Name:           name,
		Description:    description,
		GitUrl:         gitUrl,
		ID:             int32(shopId),
		OrganizationID: orgId,
	}); err != nil {
		slog.Error("failed to update shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to update shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShop deletes a shop.
func (h *Handler) DeleteShop(w http.ResponseWriter, r *http.Request, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	orgId := h.requireActiveOrganization(w, r)
	if orgId == "" {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	// Check that the shop belongs to this organization
	shop, err := h.queries.GetShopByID(r.Context(), int32(shopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if shop.OrganizationID != orgId {
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
		return
	}

	environmentCount, err := h.queries.CountEnvironmentsInShop(r.Context(), int32(shopId))
	if err == nil && environmentCount > 0 {
		httputil.WriteError(w, http.StatusConflict, "cannot delete shop with existing environments")
		return
	}

	if err := h.queries.DeleteShop(r.Context(), queries.DeleteShopParams{
		ID:             int32(shopId),
		OrganizationID: orgId,
	}); err != nil {
		slog.Error("failed to delete shop", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete shop")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
