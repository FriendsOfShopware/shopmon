package handler

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/google/uuid"
)

// HashApiKeyToken returns the sha256 hex digest of an API key token. Tokens are
// random 64-char hex (high entropy), so a plain sha256 is sufficient and allows
// constant-time comparison by comparing the stored and computed hashes.
func HashApiKeyToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

// GetApiKeyScopes returns available API key scopes.
func (h *Handler) GetApiKeyScopes(w http.ResponseWriter, r *http.Request) {
	scopes := []api.ApiKeyScope{
		{
			Value:       "deployments",
			Label:       "Deployments",
			Description: "Create and manage deployments",
		},
	}

	httputil.WriteJSON(w, http.StatusOK, scopes)
}

// GetApiKeys lists API keys for a shop.
func (h *Handler) GetApiKeys(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	rows, err := h.queries.ListApiKeysByShop(r.Context(), int32(shopId))
	if err != nil {
		slog.Error("failed to list api keys", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get api keys")
		return
	}

	result := make([]api.ApiKey, 0, len(rows))
	for _, row := range rows {
		var scopes []string
		if row.Scopes != nil {
			if err := json.Unmarshal(row.Scopes, &scopes); err != nil {
				slog.Error("failed to unmarshal api key scopes", "error", err)
			}
		}
		if scopes == nil {
			scopes = []string{}
		}

		result = append(result, api.ApiKey{
			Id:         row.ID,
			Name:       row.Name,
			Scopes:     scopes,
			CreatedAt:  pgtimeToTime(row.CreatedAt),
			LastUsedAt: pgtimeToTimePtr(row.LastUsedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateApiKey creates a new API key for a shop.
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	var req api.CreateApiKeyRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if req.Name == "" {
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
		return
	}

	// Generate a random token. This plaintext is returned to the user exactly
	// once; only its hash is persisted.
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		slog.Error("failed to generate api key token", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	scopesJSON, err := json.Marshal(req.Scopes)
	if err != nil {
		slog.Error("failed to marshal api key scopes", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to serialize scopes")
		return
	}

	id := uuid.New().String()
	_, err = h.queries.CreateApiKey(r.Context(), queries.CreateApiKeyParams{
		ID:     id,
		ShopID: int32(shopId),
		Name:   req.Name,
		Token:  HashApiKeyToken(token),
		Scopes: scopesJSON,
	})
	if err != nil {
		slog.Error("failed to create api key", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, api.CreateApiKeyResponse{
		Id:     id,
		Name:   req.Name,
		Scopes: req.Scopes,
		Token:  token,
	})
}

// DeleteApiKey deletes an API key.
func (h *Handler) DeleteApiKey(w http.ResponseWriter, r *http.Request, orgId api.OrgId, shopId api.ShopId, keyId api.KeyId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireShopInOrganization(w, r, int32(shopId), orgId) {
		return
	}

	if err := h.queries.DeleteApiKey(r.Context(), queries.DeleteApiKeyParams{
		ID:     keyId,
		ShopID: int32(shopId),
	}); err != nil {
		slog.Error("failed to delete api key", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete api key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
