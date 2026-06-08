package handler

import (
	"fmt"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetOrganizationShops returns all shops in an organization.
func (h *Handler) GetOrganizationShops(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
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
			Id:                   int(row.ID),
			Name:                 row.Name,
			Description:          row.Description,
			GitUrl:               row.GitUrl,
			OrganizationId:       row.OrganizationID,
			DefaultEnvironmentId: derefInt32(row.DefaultEnvironmentID),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateShop creates a new shop with its first environment.
func (h *Handler) CreateShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId) {
	user := h.requireUser(w, r)
	if user == nil {
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

	if req.Name == "" || req.EnvironmentName == "" || req.EnvironmentUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name, environmentName, environmentUrl, clientId, and clientSecret are required")
		return
	}

	// Validate the shop connection before opening a transaction so we never do
	// network I/O while holding a DB transaction open. The discovered version is
	// threaded into insertEnvironment so it does not re-validate inside the tx.
	shopInfo, err := h.validateShopConnection(r.Context(), req.EnvironmentUrl, req.ClientId, req.ClientSecret, "")
	if err != nil {
		slog.Error("failed to validate shop connection for new shop", "error", err)
		httputil.WriteError(w, http.StatusBadRequest, "Cannot reach shop. Check your credentials and shop URL.")
		return
	}

	// Shop insert, first environment insert and default-environment assignment must
	// be atomic so a crash cannot leave an orphaned shop or an environment without
	// a default. The initial scrape is enqueued only after the transaction commits.
	var shopID, environmentID int32
	if err := h.withTx(r.Context(), func(txq *queries.Queries) error {
		var err error
		shopID, err = txq.CreateShop(r.Context(), queries.CreateShopParams{
			OrganizationID: orgId,
			Name:           req.Name,
			Description:    req.Description,
			GitUrl:         req.GitUrl,
		})
		if err != nil {
			return fmt.Errorf("create shop: %w", err)
		}

		environmentID, err = h.insertEnvironment(r.Context(), txq, orgId, createEnvironmentCommand{
			Name:            req.EnvironmentName,
			ShopURL:         req.EnvironmentUrl,
			ClientID:        req.ClientId,
			ClientSecret:    req.ClientSecret,
			ShopID:          shopID,
			ShopwareVersion: shopInfo.Version,
		})
		if err != nil {
			return err
		}

		if err := txq.SetShopDefaultEnvironment(r.Context(), queries.SetShopDefaultEnvironmentParams{
			DefaultEnvironmentID: &environmentID,
			ID:                   shopID,
			OrganizationID:       orgId,
		}); err != nil {
			return fmt.Errorf("set default environment: %w", err)
		}

		return nil
	}); err != nil {
		slog.Error("failed to create shop with environment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create shop")
		return
	}

	h.enqueueInitialScrape(r.Context(), environmentID)

	httputil.WriteJSON(w, http.StatusCreated, api.Shop{
		Id:                   int(shopID),
		Name:                 req.Name,
		Description:          req.Description,
		GitUrl:               req.GitUrl,
		OrganizationId:       orgId,
		DefaultEnvironmentId: int(environmentID),
	})
}

// UpdateShop updates an existing shop.
func (h *Handler) UpdateShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
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

	if req.DefaultEnvironmentId != nil {
		v := int32(*req.DefaultEnvironmentId)

		env, err := h.queries.GetEnvironmentByID(r.Context(), v)
		if err != nil {
			httputil.WriteError(w, http.StatusBadRequest, "default environment not found")
			return
		}
		if env.ShopID != int32(shopId) || env.OrganizationID != orgId {
			httputil.WriteError(w, http.StatusForbidden, "default environment does not belong to this shop")
			return
		}

		if err := h.queries.SetShopDefaultEnvironment(r.Context(), queries.SetShopDefaultEnvironmentParams{
			DefaultEnvironmentID: &v,
			ID:                   int32(shopId),
			OrganizationID:       orgId,
		}); err != nil {
			slog.Error("failed to set default environment", "error", err)
			httputil.WriteError(w, http.StatusInternalServerError, "failed to set default environment")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteShop deletes a shop.
func (h *Handler) DeleteShop(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
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
