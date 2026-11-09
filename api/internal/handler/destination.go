package handler

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/delivery"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/friendsofshopware/shopmon/api/internal/organization/destination"
)

type destinationOutput struct {
	Body api.NotificationDestinationStatus
}

type listNotificationDestinationsInput struct {
	OrgID api.OrgId `path:"orgId"`
}

func (h *Handler) ListNotificationDestinations(ctx context.Context, input *listNotificationDestinationsInput) (*destinationOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	status, err := h.destinations.Status(ctx, user.ID, target)
	if err != nil {
		return nil, h.destinationError(ctx, err)
	}
	result := api.NotificationDestinationStatus{CanManage: status.CanManage, Destinations: []api.NotificationDestination{}}
	for _, d := range status.Destinations {
		result.Destinations = append(result.Destinations, api.NotificationDestination{ID: d.ID, Name: d.Name, Provider: d.Provider, Summary: d.Summary, ConnectionError: d.ConnectionError})
	}
	return &destinationOutput{Body: result}, nil
}

type createNotificationDestinationInput struct {
	OrgID api.OrgId `path:"orgId"`
	Body  api.SaveNotificationDestinationRequest
}

func (h *Handler) CreateNotificationDestination(ctx context.Context, input *createNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if err := h.destinations.Save(ctx, user.ID, target, "", input.Body.Name, input.Body.DSN); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type updateNotificationDestinationInput struct {
	OrgID         api.OrgId `path:"orgId"`
	DestinationID string    `path:"destinationId"`
	Body          api.SaveNotificationDestinationRequest
}

func (h *Handler) UpdateNotificationDestination(ctx context.Context, input *updateNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if err := h.destinations.Save(ctx, user.ID, target, input.DestinationID, input.Body.Name, input.Body.DSN); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type deleteNotificationDestinationInput struct {
	OrgID         api.OrgId `path:"orgId"`
	DestinationID string    `path:"destinationId"`
}

func (h *Handler) DeleteNotificationDestination(ctx context.Context, input *deleteNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if err := h.destinations.Delete(ctx, user.ID, target, input.DestinationID); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type testNotificationDestinationInput struct {
	OrgID         api.OrgId `path:"orgId"`
	DestinationID string    `path:"destinationId"`
}

func (h *Handler) TestNotificationDestination(ctx context.Context, input *testNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if err := h.destinations.SendTest(ctx, user.ID, target, input.DestinationID); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type listShopNotificationDestinationsInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
}

func (h *Handler) ListShopNotificationDestinations(ctx context.Context, input *listShopNotificationDestinationsInput) (*destinationOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if input.ShopID < 1 || input.ShopID > math.MaxInt32 {
		return nil, huma.Error404NotFound("shop not found")
	}
	target.ShopID = int32(input.ShopID)
	status, err := h.destinations.Status(ctx, user.ID, target)
	if err != nil {
		return nil, h.destinationError(ctx, err)
	}
	result := api.NotificationDestinationStatus{CanManage: status.CanManage, Destinations: []api.NotificationDestination{}}
	for _, d := range status.Destinations {
		result.Destinations = append(result.Destinations, api.NotificationDestination{ID: d.ID, Name: d.Name, Provider: d.Provider, Summary: d.Summary, ConnectionError: d.ConnectionError})
	}
	return &destinationOutput{Body: result}, nil
}

type createShopNotificationDestinationInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
	Body   api.SaveNotificationDestinationRequest
}

func (h *Handler) CreateShopNotificationDestination(ctx context.Context, input *createShopNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if input.ShopID < 1 || input.ShopID > math.MaxInt32 {
		return nil, huma.Error404NotFound("shop not found")
	}
	target.ShopID = int32(input.ShopID)
	if err := h.destinations.Save(ctx, user.ID, target, "", input.Body.Name, input.Body.DSN); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type updateShopNotificationDestinationInput struct {
	OrgID         api.OrgId  `path:"orgId"`
	ShopID        api.ShopId `path:"shopId"`
	DestinationID string     `path:"destinationId"`
	Body          api.SaveNotificationDestinationRequest
}

func (h *Handler) UpdateShopNotificationDestination(ctx context.Context, input *updateShopNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if input.ShopID < 1 || input.ShopID > math.MaxInt32 {
		return nil, huma.Error404NotFound("shop not found")
	}
	target.ShopID = int32(input.ShopID)
	if err := h.destinations.Save(ctx, user.ID, target, input.DestinationID, input.Body.Name, input.Body.DSN); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type deleteShopNotificationDestinationInput struct {
	OrgID         api.OrgId  `path:"orgId"`
	ShopID        api.ShopId `path:"shopId"`
	DestinationID string     `path:"destinationId"`
}

func (h *Handler) DeleteShopNotificationDestination(ctx context.Context, input *deleteShopNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if input.ShopID < 1 || input.ShopID > math.MaxInt32 {
		return nil, huma.Error404NotFound("shop not found")
	}
	target.ShopID = int32(input.ShopID)
	if err := h.destinations.Delete(ctx, user.ID, target, input.DestinationID); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

type testShopNotificationDestinationInput struct {
	OrgID         api.OrgId  `path:"orgId"`
	ShopID        api.ShopId `path:"shopId"`
	DestinationID string     `path:"destinationId"`
}

func (h *Handler) TestShopNotificationDestination(ctx context.Context, input *testShopNotificationDestinationInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	target := destination.Target{OrganizationID: input.OrgID}
	if input.ShopID < 1 || input.ShopID > math.MaxInt32 {
		return nil, huma.Error404NotFound("shop not found")
	}
	target.ShopID = int32(input.ShopID)
	if err := h.destinations.SendTest(ctx, user.ID, target, input.DestinationID); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}

func (h *Handler) destinationError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		return huma.Error403Forbidden(httputil.MsgForbidden)
	case errors.Is(err, delivery.ErrInvalidDSN), errors.Is(err, destination.ErrInvalidName), errors.Is(err, destination.ErrInvalidPreference):
		return huma.Error400BadRequest(err.Error())
	case errors.Is(err, destination.ErrShopNotFound), errors.Is(err, destination.ErrNotConfigured):
		return huma.Error404NotFound(err.Error())
	case errors.Is(err, destination.ErrDeliveryFailed):
		slog.WarnContext(ctx, "notification test failed", "error", err)
		return huma.Error502BadGateway("Test delivery failed; check the destination and credentials")
	default:
		slog.ErrorContext(ctx, "notification destination operation failed", "error", err)
		return huma.Error500InternalServerError("notification destination operation failed")
	}
}
