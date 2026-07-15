package handler

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetDeployments lists deployments for an environment.
func (h *Handler) GetDeployments(w http.ResponseWriter, r *http.Request, environmentID api.EnvironmentId, params api.GetDeploymentsParams) {
	user := h.requireUser(w, r)
	if user == nil {
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
	rows, err := h.deployments.List(r.Context(), user.ID, int32(environmentID), limit, offset)
	if err != nil {
		h.writeDeploymentError(w, r, "list deployments", err)
		return
	}

	result := make([]api.Deployment, 0, len(rows))
	for _, row := range rows {
		name := row.Name
		result = append(result, api.Deployment{
			Id:            int(row.ID),
			Command:       row.Command,
			ReturnCode:    int(row.ReturnCode),
			StartDate:     row.StartDate,
			EndDate:       row.EndDate,
			ExecutionTime: row.ExecutionTime,
			Name:          &name,
			Reference:     row.Reference,
			CreatedAt:     row.CreatedAt,
		})
	}
	httputil.WriteJSON(w, http.StatusOK, result)
}

// GetDeployment returns deployment details with output.
func (h *Handler) GetDeployment(w http.ResponseWriter, r *http.Request, environmentID api.EnvironmentId, deploymentID api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	row, err := h.deployments.Get(r.Context(), user.ID, int32(environmentID), int32(deploymentID))
	if err != nil {
		h.writeDeploymentError(w, r, "get deployment", err)
		return
	}
	name := row.Name
	httputil.WriteJSON(w, http.StatusOK, api.DeploymentDetail{
		Id:            int(row.ID),
		Command:       row.Command,
		ReturnCode:    int(row.ReturnCode),
		StartDate:     row.StartDate,
		EndDate:       row.EndDate,
		ExecutionTime: row.ExecutionTime,
		Name:          &name,
		Reference:     row.Reference,
		CreatedAt:     row.CreatedAt,
		Output:        row.Output,
	})
}

// DeleteDeployment deletes a deployment.
func (h *Handler) DeleteDeployment(w http.ResponseWriter, r *http.Request, environmentID api.EnvironmentId, deploymentID api.DeploymentId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.deployments.Delete(r.Context(), user.ID, int32(environmentID), int32(deploymentID)); err != nil {
		h.writeDeploymentError(w, r, "delete deployment", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// CreateCliDeployment creates a deployment via CLI using API key auth.
func (h *Handler) CreateCliDeployment(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		httputil.WriteError(w, http.StatusUnauthorized, "missing or invalid authorization header")
		return
	}

	var request api.CreateCliDeploymentRequest
	if err := httputil.DecodeBody(r, &request); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	result, err := h.deployments.CreateCLI(r.Context(), deployment.CreateCLICommand{
		Token:         strings.TrimPrefix(authHeader, "Bearer "),
		EnvironmentID: int32(request.EnvironmentId),
		Name:          request.Name,
		Command:       request.Command,
		ReturnCode:    int32(request.ReturnCode),
		StartDate:     request.StartDate,
		EndDate:       request.EndDate,
		ExecutionTime: request.ExecutionTime,
		Composer:      request.Composer,
		Reference:     request.Reference,
	})
	if err != nil {
		h.writeDeploymentError(w, r, "create CLI deployment", err)
		return
	}

	httputil.WriteJSON(w, http.StatusCreated, api.CreateCliDeploymentResponse{
		DeploymentId: int(result.DeploymentID),
		Name:         result.Name,
		Success:      true,
		UploadUrl:    result.UploadURL,
		Url:          result.URL,
	})
}

func (h *Handler) writeDeploymentError(w http.ResponseWriter, r *http.Request, operation string, err error) {
	switch {
	case errors.Is(err, deployment.ErrAPIKeyInvalid):
		httputil.WriteError(w, http.StatusUnauthorized, "invalid api key")
	case errors.Is(err, deployment.ErrDeploymentScopeRequired):
		httputil.WriteError(w, http.StatusForbidden, "api key does not have deployments scope")
	case errors.Is(err, deployment.ErrNotAuthorized):
		httputil.WriteError(w, http.StatusForbidden, "not a member of this organization")
	case errors.Is(err, deployment.ErrShopOrganizationMismatch):
		httputil.WriteError(w, http.StatusForbidden, "shop does not belong to this organization")
	case errors.Is(err, deployment.ErrEnvironmentShopMismatch):
		httputil.WriteError(w, http.StatusForbidden, "environment does not belong to this shop")
	case errors.Is(err, deployment.ErrShopNotFound):
		httputil.WriteError(w, http.StatusNotFound, "shop not found")
	case errors.Is(err, deployment.ErrEnvironmentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "environment not found")
	case errors.Is(err, deployment.ErrDeploymentNotFound):
		httputil.WriteError(w, http.StatusNotFound, "deployment not found")
	case errors.Is(err, deployment.ErrAPIKeyNameRequired):
		httputil.WriteError(w, http.StatusBadRequest, "name is required")
	case errors.Is(err, deployment.ErrEnvironmentIDRequired):
		httputil.WriteError(w, http.StatusBadRequest, "environmentId is required")
	case errors.Is(err, deployment.ErrCommandRequired):
		httputil.WriteError(w, http.StatusBadRequest, "command is required")
	case errors.Is(err, deployment.ErrOutputUnavailable):
		slog.ErrorContext(r.Context(), "deployment output operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusBadGateway, "deployment output unavailable")
	default:
		slog.ErrorContext(r.Context(), "deployment operation failed", "operation", operation, "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "deployment operation failed")
	}
}
