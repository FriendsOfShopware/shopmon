package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
)

type noContentOutput struct{}

type getOrganizationShopsInput struct {
	OrgID string `path:"orgId" doc:"Organization ID"`
}

type getOrganizationShopsOutput struct {
	Body []api.Shop
}

func (h *Handler) GetOrganizationShops(ctx context.Context, input *getOrganizationShopsInput) (*getOrganizationShopsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.environments.OrganizationShops(ctx, user.ID, input.OrgID)
	if errors.Is(err, environmentread.ErrNotAuthorized) {
		return nil, huma.Error403Forbidden("not a member of this organization")
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to list shops", "organizationID", input.OrgID, "error", err)
		return nil, huma.Error500InternalServerError("failed to get shops")
	}
	return &getOrganizationShopsOutput{Body: result}, nil
}

type createShopInput struct {
	OrgID string `path:"orgId" doc:"Organization ID"`
	Body  api.CreateShopRequest
}

type createShopOutput struct {
	Body api.Shop
}

func (h *Handler) CreateShop(ctx context.Context, input *createShopInput) (*createShopOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	req := input.Body
	if req.Name == "" || req.EnvironmentName == "" || req.EnvironmentUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		return nil, huma.Error400BadRequest("name, environmentName, environmentUrl, clientId, and clientSecret are required")
	}

	result, err := h.monitoring.CreateShop(ctx, monitoring.CreateShopCommand{
		UserID:          user.ID,
		OrganizationID:  input.OrgID,
		Name:            req.Name,
		Description:     req.Description,
		GitURL:          req.GitUrl,
		EnvironmentName: req.EnvironmentName,
		EnvironmentURL:  req.EnvironmentUrl,
		ClientID:        req.ClientId,
		ClientSecret:    req.ClientSecret,
	})
	if errors.Is(err, monitoring.ErrNotAuthorized) {
		return nil, huma.Error403Forbidden("not a member of this organization")
	}
	if errors.Is(err, monitoring.ErrConnectionFailed) {
		slog.ErrorContext(ctx, "failed to validate shop connection for new shop", "error", err)
		return nil, huma.Error400BadRequest("Cannot reach shop. Check your credentials and shop URL.")
	}
	if err != nil {
		slog.ErrorContext(ctx, "failed to create shop with environment", "error", err)
		return nil, huma.Error500InternalServerError("failed to create shop")
	}

	return &createShopOutput{Body: api.Shop{
		Id:                   int(result.ShopID),
		Name:                 req.Name,
		Description:          req.Description,
		GitUrl:               req.GitUrl,
		OrganizationId:       input.OrgID,
		DefaultEnvironmentId: ptr.Int(&result.EnvironmentID),
	}}, nil
}

type updateShopInput struct {
	OrgID  string     `path:"orgId" doc:"Organization ID"`
	ShopID api.ShopId `path:"shopId" doc:"Shop ID"`
	Body   api.UpdateShopRequest
}

func (h *Handler) UpdateShop(ctx context.Context, input *updateShopInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	req := input.Body
	var defaultEnvironmentID *int32
	if req.DefaultEnvironmentId != nil {
		value := int32(*req.DefaultEnvironmentId)
		defaultEnvironmentID = &value
	}
	err = h.monitoring.UpdateShop(ctx, monitoring.UpdateShopCommand{
		UserID:               user.ID,
		OrganizationID:       input.OrgID,
		ShopID:               int32(input.ShopID),
		Name:                 req.Name,
		Description:          req.Description,
		GitURL:               req.GitUrl,
		DefaultEnvironmentID: defaultEnvironmentID,
	})
	switch {
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrShopNotFound):
		return nil, huma.Error404NotFound("shop not found")
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		return nil, huma.Error403Forbidden("shop does not belong to this organization")
	case errors.Is(err, monitoring.ErrDefaultEnvironmentNotFound):
		return nil, huma.Error400BadRequest("default environment not found")
	case errors.Is(err, monitoring.ErrDefaultEnvironmentMismatch):
		return nil, huma.Error403Forbidden("default environment does not belong to this shop")
	case err != nil:
		slog.ErrorContext(ctx, "failed to update shop", "error", err)
		return nil, huma.Error500InternalServerError("failed to update shop")
	}

	return &noContentOutput{}, nil
}

type deleteShopInput struct {
	OrgID  string     `path:"orgId" doc:"Organization ID"`
	ShopID api.ShopId `path:"shopId" doc:"Shop ID"`
}

func (h *Handler) DeleteShop(ctx context.Context, input *deleteShopInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	err = h.monitoring.DeleteShop(ctx, user.ID, input.OrgID, int32(input.ShopID))
	switch {
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrShopNotFound):
		return nil, huma.Error404NotFound("shop not found")
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		return nil, huma.Error403Forbidden("shop does not belong to this organization")
	case errors.Is(err, monitoring.ErrShopHasEnvironments):
		return nil, huma.Error409Conflict("cannot delete shop with existing environments")
	case err != nil:
		slog.ErrorContext(ctx, "failed to delete shop", "error", err)
		return nil, huma.Error500InternalServerError("failed to delete shop")
	}

	return &noContentOutput{}, nil
}
