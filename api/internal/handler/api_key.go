package handler

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/google/uuid"
)

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

// GetApiKeys lists API keys for a project.
func (h *Handler) GetApiKeys(w http.ResponseWriter, r *http.Request, orgId api.OrgId, projectId api.ProjectId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireProjectInOrganization(w, r, int32(projectId), orgId) {
		return
	}

	rows, err := h.queries.ListApiKeysByProject(r.Context(), int32(projectId))
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

		// Parse ID string to int for openapi type
		id, _ := strconv.Atoi(row.ID)

		result = append(result, api.ApiKey{
			Id:         id,
			Name:       row.Name,
			Scopes:     scopes,
			CreatedAt:  pgtimeToTime(row.CreatedAt),
			LastUsedAt: pgtimeToTimePtr(row.LastUsedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// CreateApiKey creates a new API key for a project.
func (h *Handler) CreateApiKey(w http.ResponseWriter, r *http.Request, orgId api.OrgId, projectId api.ProjectId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireProjectInOrganization(w, r, int32(projectId), orgId) {
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

	// Generate a random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to generate token")
		return
	}
	token := hex.EncodeToString(tokenBytes)

	scopesJSON, err := json.Marshal(req.Scopes)
	if err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to serialize scopes")
		return
	}

	id := uuid.New().String()
	_, err = h.queries.CreateApiKey(r.Context(), queries.CreateApiKeyParams{
		ID:        id,
		ProjectID: int32(projectId),
		Name:      req.Name,
		Token:     token,
		Scopes:    scopesJSON,
	})
	if err != nil {
		slog.Error("failed to create api key", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create api key")
		return
	}

	idInt, _ := strconv.Atoi(id)
	httputil.WriteJSON(w, http.StatusCreated, api.CreateApiKeyResponse{
		Id:     idInt,
		Name:   req.Name,
		Scopes: req.Scopes,
		Token:  token,
	})
}

// DeleteApiKey deletes an API key.
func (h *Handler) DeleteApiKey(w http.ResponseWriter, r *http.Request, orgId api.OrgId, projectId api.ProjectId, keyId api.KeyId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}
	if !h.requireOrgMembership(w, r, user, orgId) {
		return
	}

	if !h.requireProjectInOrganization(w, r, int32(projectId), orgId) {
		return
	}

	if err := h.queries.DeleteApiKey(r.Context(), queries.DeleteApiKeyParams{
		ID:        strconv.Itoa(keyId),
		ProjectID: int32(projectId),
	}); err != nil {
		slog.Error("failed to delete api key", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete api key")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
