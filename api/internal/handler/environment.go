package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
)

type getOrganizationEnvironmentsInput struct {
	OrgID string `path:"orgId" doc:"Organization ID"`
}

type getOrganizationEnvironmentsOutput struct {
	Body []api.AccountEnvironment
}

func (h *Handler) GetOrganizationEnvironments(ctx context.Context, input *getOrganizationEnvironmentsInput) (*getOrganizationEnvironmentsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	result, err := h.environments.OrganizationEnvironments(ctx, user.ID, input.OrgID)
	if err != nil {
		return nil, h.writeEnvironmentReadError(ctx, "list organization environments", err)
	}
	return &getOrganizationEnvironmentsOutput{Body: result}, nil
}

type getEnvironmentInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Language      string            `query:"language" enum:"en,de" doc:"Language for localized store text (label, description, manual, changelog). Falls back to English."`
}

type getEnvironmentOutput struct {
	Body api.EnvironmentDetail
}

func (h *Handler) GetEnvironment(ctx context.Context, input *getEnvironmentInput) (*getEnvironmentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	var language *string
	if input.Language != "" {
		language = &input.Language
	}

	subscribed, err := h.monitoring.IsEnvironmentSubscribed(ctx, user.ID, int32(environmentID))
	if err != nil {
		slog.ErrorContext(ctx, "failed to check environment subscription", "environmentId", environmentID, "error", err)
		subscribed = false
	}
	detail, err := h.environments.Detail(ctx, user.ID, int32(environmentID), language, subscribed)
	if err != nil {
		return nil, h.writeEnvironmentReadError(ctx, "get environment detail", err)
	}
	return &getEnvironmentOutput{Body: detail}, nil
}

func (h *Handler) writeEnvironmentReadError(ctx context.Context, operation string, err error) error {
	switch {
	case errors.Is(err, environmentread.ErrNotFound):
		return huma.Error404NotFound("environment not found")
	case errors.Is(err, environmentread.ErrNotAuthorized):
		return huma.Error403Forbidden("not a member of this organization")
	default:
		slog.ErrorContext(ctx, "failed to load environment", "error", err, "operation", operation)
		return huma.Error500InternalServerError("failed to load environment")
	}
}

type createEnvironmentInput struct {
	Body api.CreateEnvironmentRequest
}

type createEnvironmentOutput struct {
	Body struct {
		ID int32 `json:"id"`
	}
}

func (h *Handler) CreateEnvironment(ctx context.Context, input *createEnvironmentInput) (*createEnvironmentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	req := input.Body
	if req.Name == "" || req.ShopUrl == "" || req.ClientId == "" || req.ClientSecret == "" {
		return nil, huma.Error400BadRequest("name, shopUrl, clientId, and clientSecret are required")
	}

	environmentToken := ""
	if req.EnvironmentToken != nil {
		environmentToken = *req.EnvironmentToken
	}

	environmentID, err := h.monitoring.CreateEnvironment(ctx, monitoring.CreateEnvironmentCommand{
		UserID:           user.ID,
		Name:             req.Name,
		URL:              req.ShopUrl,
		ClientID:         req.ClientId,
		ClientSecret:     req.ClientSecret,
		ShopID:           int32(req.ShopId),
		EnvironmentToken: environmentToken,
	})
	switch {
	case errors.Is(err, monitoring.ErrShopNotFound):
		return nil, huma.Error400BadRequest("shop not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrConnectionFailed):
		slog.ErrorContext(ctx, "failed to create environment", "error", err)
		return nil, huma.Error400BadRequest("Cannot reach shop. Check your credentials and shop URL.")
	case err != nil:
		slog.ErrorContext(ctx, "failed to create environment", "error", err)
		return nil, huma.Error500InternalServerError("failed to create environment")
	}

	out := &createEnvironmentOutput{}
	out.Body.ID = environmentID
	return out, nil
}

type updateEnvironmentInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Body          api.UpdateEnvironmentRequest
}

func (h *Handler) UpdateEnvironment(ctx context.Context, input *updateEnvironmentInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	req := input.Body
	err = h.monitoring.UpdateEnvironment(ctx, monitoring.UpdateEnvironmentCommand{
		UserID:        user.ID,
		EnvironmentID: int32(environmentID),
		ShopID:        int32(req.ShopId),
		Name:          req.Name,
		URL:           req.ShopUrl,
		ClientID:      req.ClientId,
		ClientSecret:  req.ClientSecret,
		Ignores:       req.Ignores,

		TaskGraceMinutes: taskGraceMinutes(req.TaskGraceMinutes),
	})
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrShopNotFound):
		return nil, huma.Error400BadRequest("target shop not found")
	case errors.Is(err, monitoring.ErrShopOrganizationMismatch):
		return nil, huma.Error403Forbidden("target shop belongs to a different organization")
	case errors.Is(err, monitoring.ErrInvalidTaskGrace):
		return nil, huma.Error400BadRequest("taskGraceMinutes must not be negative")
	case errors.Is(err, monitoring.ErrConnectionFailed):
		slog.ErrorContext(ctx, "failed to build environment update", "error", err)
		return nil, huma.Error400BadRequest("Cannot reach shop with new credentials. Check your credentials and shop URL.")
	case err != nil:
		slog.ErrorContext(ctx, "failed to update environment", "environmentID", environmentID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update environment")
	}

	return &noContentOutput{}, nil
}

type environmentIDInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
}

func (h *Handler) DeleteEnvironment(ctx context.Context, input *environmentIDInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	err = h.monitoring.DeleteEnvironment(ctx, user.ID, int32(environmentID))
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case err != nil:
		slog.ErrorContext(ctx, "failed to delete environment", "environmentID", environmentID, "error", err)
		return nil, huma.Error500InternalServerError("failed to delete environment")
	}

	return &noContentOutput{}, nil
}

type refreshEnvironmentInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Body          *api.RefreshEnvironmentJSONBody
}

func (h *Handler) RefreshEnvironment(ctx context.Context, input *refreshEnvironmentInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	sitespeed := false
	if input.Body != nil && input.Body.Sitespeed != nil {
		sitespeed = *input.Body.Sitespeed
	}
	err = h.monitoring.RefreshEnvironment(ctx, user.ID, int32(environmentID), sitespeed)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case err != nil:
		slog.ErrorContext(ctx, "failed to enqueue scrape task", "error", err)
		return nil, huma.Error500InternalServerError("failed to enqueue refresh task")
	}

	return &noContentOutput{}, nil
}

func (h *Handler) ClearEnvironmentCache(ctx context.Context, input *environmentIDInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	err = h.monitoring.ClearEnvironmentCache(ctx, user.ID, int32(environmentID))
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrCredentialDecryption):
		slog.ErrorContext(ctx, "failed to decrypt client secret", "error", err)
		return nil, huma.Error500InternalServerError("failed to clear cache")
	case errors.Is(err, monitoring.ErrRemoteOperation):
		slog.ErrorContext(ctx, "failed to clear environment cache", "error", err)
		return nil, huma.Error502BadGateway("failed to clear cache on environment")
	case err != nil:
		slog.ErrorContext(ctx, "failed to clear environment cache", "error", err)
		return nil, huma.Error500InternalServerError("failed to clear cache")
	}

	return &noContentOutput{}, nil
}

type rescheduleTaskInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	TaskID        string            `path:"taskId" doc:"Scheduled task ID"`
}

func (h *Handler) RescheduleTask(ctx context.Context, input *rescheduleTaskInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	err = h.monitoring.RescheduleTask(ctx, user.ID, int32(environmentID), input.TaskID)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case errors.Is(err, monitoring.ErrCredentialDecryption):
		slog.ErrorContext(ctx, "failed to decrypt client secret", "error", err)
		return nil, huma.Error500InternalServerError("failed to reschedule task")
	case errors.Is(err, monitoring.ErrRemoteOperation):
		slog.ErrorContext(ctx, "failed to reschedule task", "error", err)
		return nil, huma.Error502BadGateway("failed to reschedule task on environment")
	case err != nil:
		slog.ErrorContext(ctx, "failed to reschedule task", "error", err)
		return nil, huma.Error500InternalServerError("failed to reschedule task")
	}

	return &noContentOutput{}, nil
}

func (h *Handler) SubscribeToEnvironment(ctx context.Context, input *environmentIDInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	err = h.monitoring.SubscribeToEnvironment(ctx, user.ID, int32(environmentID), user.Notifications)
	if err := h.writeEnvironmentSubscriptionError(ctx, "subscribe", err); err != nil {
		return nil, err
	}
	return &noContentOutput{}, nil
}

func (h *Handler) UnsubscribeFromEnvironment(ctx context.Context, input *environmentIDInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	err = h.monitoring.UnsubscribeFromEnvironment(ctx, user.ID, int32(environmentID), user.Notifications)
	if err := h.writeEnvironmentSubscriptionError(ctx, "unsubscribe", err); err != nil {
		return nil, err
	}
	return &noContentOutput{}, nil
}

func (h *Handler) writeEnvironmentSubscriptionError(ctx context.Context, operation string, err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return huma.Error403Forbidden("not a member of this organization")
	default:
		slog.ErrorContext(ctx, "environment subscription failed", "operation", operation, "error", err)
		return huma.Error500InternalServerError("failed to update environment subscription")
	}
}

// taskGraceMinutes narrows the request's int to the int32 the domain
// carries. Huma rejects negative values against the schema minimum before the
// handler runs, and the service checks again for non-HTTP callers.
func taskGraceMinutes(value *int) *int32 {
	if value == nil {
		return nil
	}
	converted := int32(*value)
	return &converted
}

type updateSitespeedSettingsInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Body          api.SitespeedSettingsRequest
}

func (h *Handler) UpdateSitespeedSettings(ctx context.Context, input *updateSitespeedSettingsInput) (*noContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	req := input.Body
	var urls []string
	if req.Urls != nil {
		urls = append(urls, (*req.Urls)...)
	}
	err = h.monitoring.UpdateSitespeedSettings(ctx, user.ID, int32(environmentID), req.Enabled, urls)
	switch {
	case errors.Is(err, monitoring.ErrEnvironmentNotFound):
		return nil, huma.Error404NotFound("environment not found")
	case errors.Is(err, monitoring.ErrNotAuthorized):
		return nil, huma.Error403Forbidden("not a member of this organization")
	case err != nil:
		slog.ErrorContext(ctx, "failed to update sitespeed settings", "environmentID", environmentID, "error", err)
		return nil, huma.Error500InternalServerError("failed to update sitespeed settings")
	}

	return &noContentOutput{}, nil
}

type getEnvironmentChangelogsInput struct {
	EnvironmentID api.EnvironmentId `path:"environmentId" doc:"Environment ID"`
	Limit         int               `query:"limit" default:"10"`
	Offset        int               `query:"offset" default:"0"`
}

type getEnvironmentChangelogsOutput struct {
	Body api.EnvironmentChangelogsResponse
}

func (h *Handler) GetEnvironmentChangelogs(ctx context.Context, input *getEnvironmentChangelogsInput) (*getEnvironmentChangelogsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	// Query binding only parses integers; it does not enforce the spec bounds.
	// Reject out-of-range values here so they cannot reach PostgreSQL or wrap
	// during the int32 conversion. Omitted limit defaults to 10 via the input tag.
	if input.Limit < 1 || input.Limit > 100 {
		return nil, huma.Error400BadRequest("limit must be between 1 and 100")
	}
	if input.Offset < 0 {
		return nil, huma.Error400BadRequest("offset must be greater than or equal to 0")
	}

	limit := int32(input.Limit)
	offset := int32(input.Offset)

	result, err := h.environments.Changelogs(ctx, user.ID, int32(environmentID), limit, offset)
	if err != nil {
		return nil, h.writeEnvironmentReadError(ctx, "list environment changelogs", err)
	}
	return &getEnvironmentChangelogsOutput{Body: result}, nil
}

type getEnvironmentStatusEventsOutput struct {
	Body []api.StatusEvent
}

func (h *Handler) GetEnvironmentStatusEvents(ctx context.Context, input *environmentIDInput) (*getEnvironmentStatusEventsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	environmentID := int32(input.EnvironmentID)

	events, err := h.environments.StatusEvents(ctx, user.ID, int32(environmentID))
	if err != nil {
		return nil, h.writeEnvironmentReadError(ctx, "list environment status events", err)
	}
	return &getEnvironmentStatusEventsOutput{Body: events}, nil
}
