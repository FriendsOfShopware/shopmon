package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/suppression"
)

type createAdvisorySuppressionInput struct {
	AdvisoryID string `path:"advisoryId"`
	Body       api.CreateAdvisorySuppressionRequest
}

type createAdvisorySuppressionOutput struct {
	Body api.AdvisorySuppression
}

// CreateAdvisorySuppression records that an advisory is accepted or mitigated
// for a shop, removing it from the "affecting my shops" list and silencing its
// alerts.
func (h *Handler) CreateAdvisorySuppression(ctx context.Context, input *createAdvisorySuppressionInput) (*createAdvisorySuppressionOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	var environmentID *int32
	if input.Body.EnvironmentId != nil {
		id := int32(*input.Body.EnvironmentId)
		environmentID = &id
	}

	created, err := h.suppressions.Create(ctx, suppression.CreateCommand{
		UserID:        user.ID,
		ShopID:        int32(input.Body.ShopId),
		EnvironmentID: environmentID,
		AdvisoryID:    input.AdvisoryID,
		Reason:        input.Body.Reason,
		ExpiresAt:     input.Body.ExpiresAt,
	})
	if err != nil {
		return nil, writeSuppressionError(ctx, err)
	}

	return &createAdvisorySuppressionOutput{Body: toAPISuppression(created)}, nil
}

type revokeAdvisorySuppressionInput struct {
	SuppressionID int64 `path:"suppressionId"`
}

// RevokeAdvisorySuppression lifts a suppression so the advisory resurfaces. The
// row is soft-deleted, preserving who accepted the risk and why.
func (h *Handler) RevokeAdvisorySuppression(ctx context.Context, input *revokeAdvisorySuppressionInput) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load user organizations", "userId", user.ID, "error", err)
		return nil, huma.Error500InternalServerError("failed to revoke suppression")
	}

	if _, err := h.suppressions.Revoke(ctx, input.SuppressionID, user.ID, orgIDs); err != nil {
		return nil, writeSuppressionError(ctx, err)
	}

	return nil, nil
}

type listSuppressionsInput struct {
	IncludeInactive bool `query:"includeInactive"`
}

type listSuppressionsOutput struct {
	Body api.AdvisorySuppressionListResponse
}

// ListSuppressions returns the suppressions across the caller's organizations.
func (h *Handler) ListSuppressions(ctx context.Context, input *listSuppressionsInput) (*listSuppressionsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	orgIDs, err := h.advisories.OrganizationIDsForUser(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to load user organizations", "userId", user.ID, "error", err)
		return nil, huma.Error500InternalServerError("failed to list suppressions")
	}

	records, err := h.suppressions.List(ctx, orgIDs, input.IncludeInactive)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list suppressions", "error", err)
		return nil, huma.Error500InternalServerError("failed to list suppressions")
	}

	items := make([]api.AdvisorySuppression, 0, len(records))
	for _, record := range records {
		items = append(items, toAPISuppression(record))
	}
	return &listSuppressionsOutput{Body: api.AdvisorySuppressionListResponse{Suppressions: items}}, nil
}

func toAPISuppression(record suppression.Suppression) api.AdvisorySuppression {
	item := api.AdvisorySuppression{
		Id:             record.ID,
		AdvisoryId:     record.AdvisoryID,
		OrganizationId: record.OrganizationID,
		ShopId:         int(record.ShopID),
		ShopName:       record.ShopName,
		Reason:         record.Reason,
		ExpiresAt:      record.ExpiresAt,
		CreatedBy:      record.CreatedBy,
		CreatedAt:      record.CreatedAt,
		RevokedAt:      record.RevokedAt,
	}
	if record.EnvironmentID != nil {
		id := int(*record.EnvironmentID)
		item.EnvironmentId = &id
	}
	if record.EnvironmentName != nil && *record.EnvironmentName != "" {
		item.EnvironmentName = record.EnvironmentName
	}
	if record.CreatedByName != nil && *record.CreatedByName != "" {
		item.CreatedByName = record.CreatedByName
	}
	return item
}

// writeSuppressionError maps capability errors onto transport status codes.
// A duplicate is a 409 rather than a silent overwrite, which is the behaviour
// the legacy whole-array ignores PATCH could not offer.
func writeSuppressionError(ctx context.Context, err error) error {
	switch {
	case errors.Is(err, suppression.ErrReasonRequired):
		return huma.Error400BadRequest("a reason is required")
	case errors.Is(err, suppression.ErrExpiryInPast):
		return huma.Error400BadRequest("expiry must be in the future")
	case errors.Is(err, suppression.ErrEnvironmentScope):
		return huma.Error400BadRequest("environment does not belong to shop")
	case errors.Is(err, suppression.ErrNotAuthorized):
		return huma.Error403Forbidden("not authorized to suppress for this shop")
	case errors.Is(err, suppression.ErrShopNotFound):
		return huma.Error404NotFound("shop not found")
	case errors.Is(err, suppression.ErrAdvisoryNotFound):
		return huma.Error404NotFound("advisory not found")
	case errors.Is(err, suppression.ErrNotFound):
		return huma.Error404NotFound("suppression not found")
	case errors.Is(err, suppression.ErrAlreadySuppressed):
		return huma.Error409Conflict("an active suppression already covers this scope")
	default:
		slog.ErrorContext(ctx, "suppression request failed", "error", err)
		return huma.Error500InternalServerError("failed to process suppression")
	}
}
