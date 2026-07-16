package handler

import (
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// HashApiKeyToken is retained for compatibility with integrations that seeded
// API keys through handler tests. New code should use deployment.HashAPIKeyToken.
func HashApiKeyToken(token string) string {
	return deployment.HashAPIKeyToken(token)
}

// GetApiKeyScopes returns available API key scopes.
func (h *Handler) GetApiKeyScopes(w http.ResponseWriter, _ *http.Request) {
	scopes := deployment.AvailableScopes()
	result := make([]api.ApiKeyScope, 0, len(scopes))
	for _, scope := range scopes {
		result = append(result, api.ApiKeyScope{
			Value:       scope.Value,
			Label:       scope.Label,
			Description: scope.Description,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetApiKeys lists API keys for a shop.
func (h *Handler) GetApiKeys(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	keys, err := h.deployments.ListAPIKeys(r.Context(), deployment.ShopCommand{
		UserID:         user.ID,
		OrganizationID: orgID,
		ShopID:         int32(shopID),
	})
	if err != nil {
		h.writeDeploymentError(w, r, "list api keys", err)
		return
	}

	result := make([]api.ApiKey, 0, len(keys))
	for _, key := range keys {
		result = append(result, api.ApiKey{
			Id:         key.ID,
			Name:       key.Name,
			Scopes:     key.Scopes,
			CreatedAt:  key.CreatedAt,
			LastUsedAt: key.LastUsedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateApiKey creates a new API key for a shop.
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	var request api.CreateApiKeyRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	key, err := h.deployments.CreateAPIKey(r.Context(), deployment.CreateAPIKeyCommand{
		ShopCommand: deployment.ShopCommand{
			UserID:         user.ID,
			OrganizationID: orgID,
			ShopID:         int32(shopID),
		},
		Name:   request.Name,
		Scopes: request.Scopes,
	})
	if err != nil {
		h.writeDeploymentError(w, r, "create api key", err)
		return
	}
	httputil.WriteJSON(w, http.StatusCreated, api.CreateApiKeyResponse{
		Id:     key.ID,
		Name:   key.Name,
		Scopes: key.Scopes,
		Token:  key.Token,
	})
}

// DeleteApiKey deletes an API key.
func (h *Handler) DeleteApiKey(w http.ResponseWriter, r *http.Request, orgID api.OrgId, shopID api.ShopId, keyID api.KeyId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if err := h.deployments.DeleteAPIKey(r.Context(), deployment.ShopCommand{
		UserID:         user.ID,
		OrganizationID: orgID,
		ShopID:         int32(shopID),
	}, keyID); err != nil {
		h.writeDeploymentError(w, r, "delete api key", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
