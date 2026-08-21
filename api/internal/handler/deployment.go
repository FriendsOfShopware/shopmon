package handler

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"strings"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
)

type getDeploymentsInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId"`
	Limit         int               `query:"limit"`
	Offset        int               `query:"offset"`
}

type getDeploymentsOutput struct {
	Body []api.Deployment
}

type getDeploymentInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId"`
	DeploymentID  api.DeploymentId  `path:"deploymentId"`
}

type getDeploymentOutput struct {
	Body api.DeploymentDetail
}

type deleteDeploymentInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId"`
	DeploymentID  api.DeploymentId  `path:"deploymentId"`
}

type deleteDeploymentOutput struct {
	Status int
}

type createCliDeploymentInput struct {
	Authorization string `header:"Authorization"`
	Body          api.CreateCliDeploymentRequest
}

type createCliDeploymentOutput struct {
	Status int
	Body   api.CreateCliDeploymentResponse
}

// GetDeployments lists deployments for an environment.
func (h *Handler) GetDeployments(ctx context.Context, input *getDeploymentsInput) (*getDeploymentsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	limit := int32(25)
	offset := int32(input.Offset)
	if input.Limit != 0 {
		limit = int32(input.Limit)
	}
	rows, err := h.deployments.List(ctx, user.ID, int32(input.EnvironmentID), limit, offset)
	if err != nil {
		return nil, h.writeDeploymentError(ctx, "list deployments", err)
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
	return &getDeploymentsOutput{Body: result}, nil
}

// GetDeployment returns deployment details with output.
func (h *Handler) GetDeployment(ctx context.Context, input *getDeploymentInput) (*getDeploymentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	row, err := h.deployments.Get(ctx, user.ID, int32(input.EnvironmentID), int32(input.DeploymentID))
	if err != nil {
		return nil, h.writeDeploymentError(ctx, "get deployment", err)
	}
	name := row.Name
	return &getDeploymentOutput{Body: api.DeploymentDetail{
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
	}}, nil
}

// DeleteDeployment deletes a deployment.
func (h *Handler) DeleteDeployment(ctx context.Context, input *deleteDeploymentInput) (*deleteDeploymentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.deployments.Delete(ctx, user.ID, int32(input.EnvironmentID), int32(input.DeploymentID)); err != nil {
		return nil, h.writeDeploymentError(ctx, "delete deployment", err)
	}
	return &deleteDeploymentOutput{Status: http.StatusNoContent}, nil
}

// CreateCliDeployment creates a deployment via CLI using API key auth.
func (h *Handler) CreateCliDeployment(ctx context.Context, input *createCliDeploymentInput) (*createCliDeploymentOutput, error) {
	if !strings.HasPrefix(input.Authorization, "Bearer ") {
		return nil, huma.Error401Unauthorized("missing or invalid authorization header")
	}

	result, err := h.deployments.CreateCLI(ctx, deployment.CreateCLICommand{
		Token:         strings.TrimPrefix(input.Authorization, "Bearer "),
		EnvironmentID: int32(input.Body.EnvironmentId),
		Name:          input.Body.Name,
		Command:       input.Body.Command,
		ReturnCode:    int32(input.Body.ReturnCode),
		StartDate:     input.Body.StartDate,
		EndDate:       input.Body.EndDate,
		ExecutionTime: input.Body.ExecutionTime,
		Composer:      input.Body.Composer,
		Reference:     input.Body.Reference,
	})
	if err != nil {
		return nil, h.writeDeploymentError(ctx, "create CLI deployment", err)
	}

	return &createCliDeploymentOutput{
		Status: http.StatusCreated,
		Body: api.CreateCliDeploymentResponse{
			DeploymentId: int(result.DeploymentID),
			Name:         result.Name,
			Success:      true,
			UploadUrl:    result.UploadURL,
			Url:          result.URL,
		},
	}, nil
}

func (h *Handler) writeDeploymentError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, deployment.ErrAPIKeyInvalid):
		return huma.Error401Unauthorized("invalid api key")
	case errors.Is(err, deployment.ErrDeploymentScopeRequired):
		return huma.Error403Forbidden("api key does not have deployments scope")
	case errors.Is(err, deployment.ErrNotAuthorized):
		return huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, deployment.ErrShopOrganizationMismatch):
		return huma.Error403Forbidden("shop does not belong to this organization")
	case errors.Is(err, deployment.ErrEnvironmentShopMismatch):
		return huma.Error403Forbidden("environment does not belong to this shop")
	case errors.Is(err, deployment.ErrShopNotFound):
		return huma.Error404NotFound("shop not found")
	case errors.Is(err, deployment.ErrEnvironmentNotFound):
		return huma.Error404NotFound("environment not found")
	case errors.Is(err, deployment.ErrDeploymentNotFound):
		return huma.Error404NotFound("deployment not found")
	case errors.Is(err, deployment.ErrAPIKeyNameRequired):
		return huma.Error400BadRequest("name is required")
	case errors.Is(err, deployment.ErrEnvironmentIDRequired):
		return huma.Error400BadRequest("environmentId is required")
	case errors.Is(err, deployment.ErrCommandRequired):
		return huma.Error400BadRequest("command is required")
	case errors.Is(err, deployment.ErrOutputUnavailable):
		slog.ErrorContext(ctx, "deployment output operation failed", "operation", operation, "error", err)
		return huma.Error502BadGateway("deployment output unavailable")
	default:
		slog.ErrorContext(ctx, "deployment operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("deployment operation failed")
	}
}
