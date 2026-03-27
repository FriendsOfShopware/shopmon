package handler

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/jackc/pgx/v5/pgtype"
)

// GetDeployments lists deployments for a shop.
func (h *Handler) GetDeployments(w http.ResponseWriter, r *http.Request, shopId api.ShopId, params api.GetDeploymentsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	limit := int32(25)
	offset := int32(0)
	if params.Limit != nil {
		limit = int32(*params.Limit)
	}
	if params.Offset != nil {
		offset = int32(*params.Offset)
	}

	rows, err := h.queries.ListDeployments(r.Context(), queries.ListDeploymentsParams{
		ShopID: int32(shopId),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		slog.Error("failed to list deployments", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get deployments")
		return
	}

	result := make([]api.Deployment, 0, len(rows))
	for _, row := range rows {
		name := row.Name
		result = append(result, api.Deployment{
			Id:            int(row.ID),
			Command:       row.Command,
			ReturnCode:    int(row.ReturnCode),
			StartDate:     pgtimeToTime(row.StartDate),
			EndDate:       pgtimeToTime(row.EndDate),
			ExecutionTime: row.ExecutionTime,
			Name:          &name,
			Reference:     row.Reference,
			CreatedAt:     pgtimeToTime(row.CreatedAt),
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetDeployment returns deployment details with output.
func (h *Handler) GetDeployment(w http.ResponseWriter, r *http.Request, shopId api.ShopId, deploymentId api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	row, err := h.queries.GetDeploymentByID(r.Context(), queries.GetDeploymentByIDParams{
		ID:     int32(deploymentId),
		ShopID: int32(shopId),
	})
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "deployment not found")
		return
	}

	name := row.Name

	// Fetch output from S3
	var output *string
	if h.storage != nil {
		o, err := h.storage.GetDeploymentOutput(r.Context(), int(row.ID))
		if err == nil {
			output = &o
		} else {
			slog.Warn("failed to get deployment output from S3", "deploymentId", row.ID, "error", err)
		}
	}

	detail := api.DeploymentDetail{
		Id:            int(row.ID),
		Command:       row.Command,
		ReturnCode:    int(row.ReturnCode),
		StartDate:     pgtimeToTime(row.StartDate),
		EndDate:       pgtimeToTime(row.EndDate),
		ExecutionTime: row.ExecutionTime,
		Name:          &name,
		Reference:     row.Reference,
		CreatedAt:     pgtimeToTime(row.CreatedAt),
		Output:        output,
	}

	httputil.WriteJSON(w, http.StatusOK, detail)
}

// DeleteDeployment deletes a deployment.
func (h *Handler) DeleteDeployment(w http.ResponseWriter, r *http.Request, shopId api.ShopId, deploymentId api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedShop(w, r, user, int32(shopId)); !ok {
		return
	}

	if err := h.queries.DeleteDeployment(r.Context(), queries.DeleteDeploymentParams{
		ID:     int32(deploymentId),
		ShopID: int32(shopId),
	}); err != nil {
		slog.Error("failed to delete deployment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete deployment")
		return
	}

	// Clean up S3 output
	if h.storage != nil {
		if err := h.storage.DeleteDeploymentOutput(r.Context(), deploymentId); err != nil {
			slog.Warn("failed to delete deployment output from S3", "deploymentId", deploymentId, "error", err)
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

// CreateCliDeployment creates a deployment via CLI using API key auth.
func (h *Handler) CreateCliDeployment(w http.ResponseWriter, r *http.Request) {
	// Extract Bearer token from Authorization header
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		httputil.WriteError(w, http.StatusUnauthorized, "missing or invalid authorization header")
		return
	}
	token := strings.TrimPrefix(authHeader, "Bearer ")

	// Look up the API key
	apiKey, err := h.queries.GetApiKeyByToken(r.Context(), token)
	if err != nil {
		httputil.WriteError(w, http.StatusUnauthorized, "invalid api key")
		return
	}

	// Update last used
	_ = h.queries.UpdateApiKeyLastUsed(r.Context(), apiKey.ID)

	// Verify scope includes "deployments"
	var scopes []string
	if err := json.Unmarshal(apiKey.Scopes, &scopes); err != nil {
		httputil.WriteError(w, http.StatusInternalServerError, "failed to parse api key scopes")
		return
	}

	hasDeploymentScope := false
	for _, scope := range scopes {
		if scope == "deployments" {
			hasDeploymentScope = true
			break
		}
	}
	if !hasDeploymentScope {
		httputil.WriteError(w, http.StatusForbidden, "api key does not have deployments scope")
		return
	}

	var req api.CreateCliDeploymentRequest
	if err := httputil.DecodeBody(r, &req); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	// Verify shop belongs to the API key's project
	shop, err := h.queries.GetShopByID(r.Context(), int32(req.ShopId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
		return
	}

	if shop.ProjectID != apiKey.ProjectID {
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this project")
		return
	}

	name := fmt.Sprintf("CLI Deployment #%d", req.ShopId)
	if req.Name != nil {
		name = *req.Name
	}

	var composerBytes []byte
	if req.Composer != nil {
		composerBytes = []byte(*req.Composer)
	}

	deploymentID, err := h.queries.CreateDeployment(r.Context(), queries.CreateDeploymentParams{
		ShopID:        int32(req.ShopId),
		Name:          name,
		Command:       req.Command,
		ReturnCode:    int32(req.ReturnCode),
		StartDate:     pgtype.Timestamp{Time: req.StartDate, Valid: true},
		EndDate:       pgtype.Timestamp{Time: req.EndDate, Valid: true},
		ExecutionTime: req.ExecutionTime,
		Composer:      composerBytes,
		Reference:     req.Reference,
	})
	if err != nil {
		slog.Error("failed to create deployment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to create deployment")
		return
	}

	// Update the active deployment on the shop
	deployID := deploymentID
	_ = h.queries.UpdateShopActiveDeployment(r.Context(), queries.UpdateShopActiveDeploymentParams{
		ActiveDeploymentID: &deployID,
		ID:                 int32(req.ShopId),
	})

	// Get presigned upload URL from S3
	var uploadURL string
	if h.storage != nil {
		url, err := h.storage.PresignUpload(r.Context(), int(deploymentID))
		if err != nil {
			slog.Error("failed to get presigned upload URL", "error", err)
		} else {
			uploadURL = url
		}
	}

	httputil.WriteJSON(w, http.StatusCreated, api.CreateCliDeploymentResponse{
		DeploymentId: int(deploymentID),
		Name:         name,
		Success:      true,
		UploadUrl:    uploadURL,
		Url:          fmt.Sprintf("%s/shops/%d/deployments/%d", h.cfg.FrontendURL, req.ShopId, deploymentID),
	})
}
