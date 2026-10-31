package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
)

type getEnvironmentAccountsOutput struct {
	Body api.EnvironmentAccounts
}

// GetEnvironmentAccounts returns the account-monitoring view: findings awaiting
// review, the change history and posture hints. It deliberately does not
// return a directory of the shop's accounts.
func (h *Handler) GetEnvironmentAccounts(ctx context.Context, input *environmentIDInput) (*getEnvironmentAccountsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	result, err := h.environments.Accounts(ctx, user.ID, int32(input.EnvironmentID))
	if err != nil {
		return nil, h.writeEnvironmentReadError(ctx, "get environment accounts", err)
	}
	return &getEnvironmentAccountsOutput{Body: result}, nil
}

type acknowledgeAccountEventInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	EventID       int64             `path:"eventId" doc:"Account event ID"`
	Body          api.AcknowledgeAccountEventRequest
}

// AcknowledgeEnvironmentAccountEvent records that a reviewer accepted an
// account change. The finding stops escalating the environment status once the
// refresh this triggers has run.
func (h *Handler) AcknowledgeEnvironmentAccountEvent(ctx context.Context, input *acknowledgeAccountEventInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)
	err = h.monitoring.AcknowledgeAccountEvent(ctx, user.ID, environmentID, input.EventID, input.Body.Note)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrAccountEventNotFound):
		// Also covers an event that someone else already reviewed: a repeated
		// submit must not read as a second, independent review.
		return nil, huma.Error404NotFound("account event not found or already acknowledged")
	case err != nil:
		slog.ErrorContext(ctx, "failed to acknowledge account event",
			"environmentID", environmentID, "eventID", input.EventID, "error", err)
		return nil, huma.Error500InternalServerError("failed to acknowledge account event")
	}

	return &noContentOutput{}, nil
}

type updateAccountMonitoringInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Body          api.UpdateAccountMonitoringRequest
}

// UpdateAccountMonitoringSettings is the per-environment opt-out. It lives
// outside the generic environment PATCH so a caller that omits the field
// cannot switch the security feature off by accident.
func (h *Handler) UpdateAccountMonitoringSettings(ctx context.Context, input *updateAccountMonitoringInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)
	err = h.monitoring.SetAccountMonitoring(ctx, user.ID, environmentID, input.Body.Enabled)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case err != nil:
		slog.ErrorContext(ctx, "failed to update account monitoring settings",
			"environmentID", environmentID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update account monitoring settings")
	}

	return &noContentOutput{}, nil
}
