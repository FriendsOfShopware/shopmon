package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	smithy "github.com/aws/smithy-go"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/jackc/pgx/v5/pgtype"
	goqueue "github.com/shyim/go-queue"
)

// isS3NotFound reports whether err indicates the requested S3 object does not
// exist (as opposed to a transport, permission, or other genuine failure).
func isS3NotFound(err error) bool {
	var noSuchKey *types.NoSuchKey
	if errors.As(err, &noSuchKey) {
		return true
	}
	var notFound *types.NotFound
	if errors.As(err, &notFound) {
		return true
	}
	var apiErr smithy.APIError
	if errors.As(err, &apiErr) {
		switch apiErr.ErrorCode() {
		case "NoSuchKey", "NotFound", "404":
			return true
		}
	}
	return false
}

// GetDeployments lists deployments for an environment.
func (h *Handler) GetDeployments(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, params api.GetDeploymentsParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
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
		EnvironmentID: int32(environmentId),
		Limit:         limit,
		Offset:        offset,
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
func (h *Handler) GetDeployment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, deploymentId api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	row, err := h.queries.GetDeploymentByID(r.Context(), queries.GetDeploymentByIDParams{
		ID:            int32(deploymentId),
		EnvironmentID: int32(environmentId),
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
		switch {
		case err == nil:
			output = &o
		case isS3NotFound(err):
			// No output stored yet: legitimately absent, return null.
		default:
			slog.Error("failed to get deployment output from S3", "deploymentId", row.ID, "error", err)
			httputil.WriteError(w, http.StatusBadGateway, "failed to get deployment output")
			return
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
func (h *Handler) DeleteDeployment(w http.ResponseWriter, r *http.Request, environmentId api.EnvironmentId, deploymentId api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if _, ok := h.loadAuthorizedEnvironment(w, r, user, int32(environmentId)); !ok {
		return
	}

	if err := h.queries.DeleteDeployment(r.Context(), queries.DeleteDeploymentParams{
		ID:            int32(deploymentId),
		EnvironmentID: int32(environmentId),
	}); err != nil {
		slog.Error("failed to delete deployment", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete deployment")
		return
	}

	// Clean up S3 output
	if h.storage != nil {
		cleanupCtx, cancel := deploymentOutputCleanupContext(r.Context())
		defer cancel()

		if err := h.storage.DeleteDeploymentOutput(cleanupCtx, deploymentId); err != nil {
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

	// Look up the API key by its hashed-at-rest token.
	apiKey, err := h.queries.GetApiKeyByToken(r.Context(), HashApiKeyToken(token))
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

	if req.EnvironmentId <= 0 {
		httputil.WriteError(w, http.StatusBadRequest, "environmentId is required")
		return
	}
	if strings.TrimSpace(req.Command) == "" {
		httputil.WriteError(w, http.StatusBadRequest, "command is required")
		return
	}

	// Verify environment belongs to the API key's shop
	environment, err := h.queries.GetEnvironmentByID(r.Context(), int32(req.EnvironmentId))
	if err != nil {
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
		return
	}

	if environment.ShopID != apiKey.ShopID {
		httputil.WriteError(w, http.StatusForbidden, "environment does not belong to this shop")
		return
	}

	name := fmt.Sprintf("CLI Deployment #%d", req.EnvironmentId)
	if req.Name != nil {
		name = *req.Name
	}

	var composerBytes []byte
	if req.Composer != nil {
		composerBytes = []byte(*req.Composer)
	}

	deploymentID, err := h.queries.CreateDeployment(r.Context(), queries.CreateDeploymentParams{
		EnvironmentID: int32(req.EnvironmentId),
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

	// Update the active deployment on the environment
	deployID := deploymentID
	_ = h.queries.UpdateEnvironmentActiveDeployment(r.Context(), queries.UpdateEnvironmentActiveDeploymentParams{
		ActiveDeploymentID: &deployID,
		ID:                 int32(req.EnvironmentId),
	})

	// Refresh the shop data after the deployment so Shopmon reflects the new
	// state. The scrape is delayed to give post-deploy tasks (theme compile,
	// indexing, cache warming) time to settle; the delay is configurable via
	// DEPLOYMENT_SCRAPE_DELAY.
	if err := goqueue.Dispatch(r.Context(), h.bus, jobs.EnvironmentScrape{EnvironmentID: int32(req.EnvironmentId)}, goqueue.WithDelay(h.cfg.DeploymentScrapeDelay)); err != nil {
		slog.Error("failed to enqueue post-deployment scrape", "environmentId", req.EnvironmentId, "error", err)
	}

	// Get presigned upload URL from S3
	var uploadURL string
	if h.storage != nil {
		url, err := h.storage.PresignUpload(r.Context(), int(deploymentID))
		if err != nil {
			slog.Error("failed to get presigned upload URL", "deploymentId", deploymentID, "error", err)
			httputil.WriteError(w, http.StatusBadGateway, "failed to create upload url")
			return
		}
		uploadURL = url
	}

	httputil.WriteJSON(w, http.StatusCreated, api.CreateCliDeploymentResponse{
		DeploymentId: int(deploymentID),
		Name:         name,
		Success:      true,
		UploadUrl:    uploadURL,
		Url:          fmt.Sprintf("%s/environments/%d/deployments/%d", h.cfg.FrontendURL, req.EnvironmentId, deploymentID),
	})
}
