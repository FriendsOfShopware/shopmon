package handler

import (
	"context"
	"errors"
	"log/slog"
	"math"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	"github.com/friendsofshopware/shopmon/api/internal/organization/slackwebhook"
	"github.com/friendsofshopware/shopmon/api/internal/slack"
)

type slackWebhookPathInput struct {
	OrgID api.OrgId `path:"orgId"`
}

type shopSlackWebhookPathInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
}

type updateSlackWebhookInput struct {
	OrgID api.OrgId `path:"orgId"`
	Body  api.UpdateSlackWebhookRequest
}

type updateShopSlackWebhookInput struct {
	OrgID  api.OrgId  `path:"orgId"`
	ShopID api.ShopId `path:"shopId"`
	Body   api.UpdateSlackWebhookRequest
}

type getSlackWebhookOutput struct {
	Body api.SlackWebhookStatus
}

func orgSlackTarget(orgID string) slackwebhook.Target {
	return slackwebhook.Target{OrganizationID: orgID}
}

// shopSlackTarget rejects IDs that do not fit a shop: zero would address the
// organization webhook, and values beyond int32 would wrap onto another shop.
func shopSlackTarget(orgID string, shopID api.ShopId) (slackwebhook.Target, error) {
	if shopID < 1 || shopID > math.MaxInt32 {
		return slackwebhook.Target{}, huma.Error404NotFound("shop not found")
	}
	return slackwebhook.Target{OrganizationID: orgID, ShopID: int32(shopID)}, nil
}

// GetSlackWebhook reports whether the organization has a Slack webhook.
func (h *Handler) GetSlackWebhook(ctx context.Context, input *slackWebhookPathInput) (*getSlackWebhookOutput, error) {
	return h.getSlackWebhook(ctx, orgSlackTarget(input.OrgID))
}

// GetShopSlackWebhook reports whether a shop has a Slack webhook.
func (h *Handler) GetShopSlackWebhook(ctx context.Context, input *shopSlackWebhookPathInput) (*getSlackWebhookOutput, error) {
	target, err := shopSlackTarget(input.OrgID, input.ShopID)
	if err != nil {
		return nil, err
	}
	return h.getSlackWebhook(ctx, target)
}

// UpdateSlackWebhook stores the organization's Slack webhook.
func (h *Handler) UpdateSlackWebhook(ctx context.Context, input *updateSlackWebhookInput) (*struct{}, error) {
	return h.updateSlackWebhook(ctx, orgSlackTarget(input.OrgID), input.Body.WebhookUrl)
}

// UpdateShopSlackWebhook stores a shop's Slack webhook.
func (h *Handler) UpdateShopSlackWebhook(ctx context.Context, input *updateShopSlackWebhookInput) (*struct{}, error) {
	target, err := shopSlackTarget(input.OrgID, input.ShopID)
	if err != nil {
		return nil, err
	}
	return h.updateSlackWebhook(ctx, target, input.Body.WebhookUrl)
}

// DeleteSlackWebhook removes the organization's Slack webhook.
func (h *Handler) DeleteSlackWebhook(ctx context.Context, input *slackWebhookPathInput) (*struct{}, error) {
	return h.deleteSlackWebhook(ctx, orgSlackTarget(input.OrgID))
}

// DeleteShopSlackWebhook removes a shop's Slack webhook.
func (h *Handler) DeleteShopSlackWebhook(ctx context.Context, input *shopSlackWebhookPathInput) (*struct{}, error) {
	target, err := shopSlackTarget(input.OrgID, input.ShopID)
	if err != nil {
		return nil, err
	}
	return h.deleteSlackWebhook(ctx, target)
}

// TestSlackWebhook posts a test message to the organization's Slack webhook.
func (h *Handler) TestSlackWebhook(ctx context.Context, input *slackWebhookPathInput) (*struct{}, error) {
	return h.testSlackWebhook(ctx, orgSlackTarget(input.OrgID))
}

// TestShopSlackWebhook posts a test message to a shop's Slack webhook.
func (h *Handler) TestShopSlackWebhook(ctx context.Context, input *shopSlackWebhookPathInput) (*struct{}, error) {
	target, err := shopSlackTarget(input.OrgID, input.ShopID)
	if err != nil {
		return nil, err
	}
	return h.testSlackWebhook(ctx, target)
}

func (h *Handler) getSlackWebhook(ctx context.Context, target slackwebhook.Target) (*getSlackWebhookOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	status, err := h.slack.Status(ctx, user.ID, target)
	if err != nil {
		return nil, h.writeSlackError(ctx, "get slack webhook", err)
	}
	return &getSlackWebhookOutput{Body: api.SlackWebhookStatus{
		Configured: status.Configured,
		CanManage:  status.CanManage,
	}}, nil
}

func (h *Handler) updateSlackWebhook(ctx context.Context, target slackwebhook.Target, webhookURL string) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.slack.Set(ctx, user.ID, target, webhookURL); err != nil {
		return nil, h.writeSlackError(ctx, "update slack webhook", err)
	}
	return nil, nil
}

func (h *Handler) deleteSlackWebhook(ctx context.Context, target slackwebhook.Target) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.slack.Delete(ctx, user.ID, target); err != nil {
		return nil, h.writeSlackError(ctx, "delete slack webhook", err)
	}
	return nil, nil
}

func (h *Handler) testSlackWebhook(ctx context.Context, target slackwebhook.Target) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.slack.SendTest(ctx, user.ID, target); err != nil {
		return nil, h.writeSlackError(ctx, "test slack webhook", err)
	}
	return nil, nil
}

func (h *Handler) writeSlackError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, organization.ErrMembershipNotFound), errors.Is(err, organization.ErrRoleNotAllowed):
		return huma.Error403Forbidden(httputil.MsgForbidden)
	case errors.Is(err, slack.ErrInvalidWebhookURL):
		return huma.Error400BadRequest(err.Error())
	case errors.Is(err, slackwebhook.ErrShopNotFound):
		return huma.Error404NotFound("shop not found")
	case errors.Is(err, slackwebhook.ErrNotConfigured):
		return huma.Error404NotFound("slack webhook not configured")
	case errors.Is(err, slackwebhook.ErrDeliveryFailed):
		slog.WarnContext(ctx, "slack test message failed", "error", err)
		return huma.Error502BadGateway("Slack rejected the test message; check the webhook URL")
	default:
		slog.ErrorContext(ctx, "slack webhook operation failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("slack webhook operation failed")
	}
}
