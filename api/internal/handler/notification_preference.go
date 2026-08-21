package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
)

type getNotificationEventTypesOutput struct {
	Body []api.NotificationEventType
}

type getNotificationPreferencesOutput struct {
	Body []api.NotificationPreference
}

type setNotificationPreferenceInput struct {
	Body api.NotificationPreferenceInput
}

type deleteNotificationPreferenceInput struct {
	ScopeType string `query:"scopeType"`
	ScopeId   string `query:"scopeId"`
	EventType string `query:"eventType"`
	Channel   string `query:"channel"`
}

func (h *Handler) GetNotificationEventTypes(ctx context.Context, _ *struct{}) (*getNotificationEventTypesOutput, error) {
	if _, err := h.requireUser(ctx); err != nil {
		return nil, err
	}

	types := notify.ListEventTypes()
	result := make([]api.NotificationEventType, 0, len(types))
	for _, t := range types {
		result = append(result, api.NotificationEventType{
			Type:            t.Type,
			DefaultChannels: t.DefaultChannels,
		})
	}

	return &getNotificationEventTypesOutput{Body: result}, nil
}

// validPreferenceScopes and validPreferenceChannels bound the values a client
// may store. The empty channel is the subscription marker and is managed via
// the subscribe endpoints, so it is not settable here.
var (
	validPreferenceScopes   = map[string]bool{"global": true, "organization": true, "environment": true}
	validPreferenceChannels = map[string]bool{"in_app": true, "email": true}
)

func (h *Handler) GetNotificationPreferences(ctx context.Context, _ *struct{}) (*getNotificationPreferencesOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := h.notifications.ListPreferences(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list notification preferences", "error", err)
		return nil, huma.Error500InternalServerError("failed to get preferences")
	}

	result := make([]api.NotificationPreference, 0, len(rows))
	for _, row := range rows {
		pref := api.NotificationPreference{
			ScopeType: row.ScopeType,
			ScopeId:   row.ScopeID,
			EventType: row.EventType,
			Channel:   row.Channel,
			Enabled:   row.Enabled,
		}
		result = append(result, pref)
	}

	return &getNotificationPreferencesOutput{Body: result}, nil
}

func (h *Handler) SetNotificationPreference(ctx context.Context, input *setNotificationPreferenceInput) (*notificationNoContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	body := input.Body
	if !validPreferenceScopes[body.ScopeType] {
		return nil, huma.Error400BadRequest("invalid scope type")
	}
	if !validPreferenceChannels[body.Channel] {
		return nil, huma.Error400BadRequest("invalid channel")
	}

	scopeID := ""
	if body.ScopeId != nil {
		scopeID = *body.ScopeId
	}
	eventType := ""
	if body.EventType != nil {
		eventType = *body.EventType
	}

	if err := h.notifications.SetPreference(ctx, user.ID, notification.Preference{
		ScopeType: body.ScopeType,
		ScopeID:   scopeID,
		EventType: eventType,
		Channel:   body.Channel,
		Enabled:   body.Enabled,
	}); err != nil {
		slog.ErrorContext(ctx, "failed to set notification preference", "error", err)
		return nil, huma.Error500InternalServerError("failed to set preference")
	}

	return &notificationNoContentOutput{Status: http.StatusNoContent}, nil
}

func (h *Handler) DeleteNotificationPreference(ctx context.Context, input *deleteNotificationPreferenceInput) (*notificationNoContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if !validPreferenceScopes[input.ScopeType] {
		return nil, huma.Error400BadRequest("invalid scope type")
	}
	if !validPreferenceChannels[input.Channel] {
		return nil, huma.Error400BadRequest("invalid channel")
	}

	if err := h.notifications.DeletePreference(ctx, user.ID, input.ScopeType, input.ScopeId, input.EventType, input.Channel); err != nil {
		slog.ErrorContext(ctx, "failed to delete notification preference", "error", err)
		return nil, huma.Error500InternalServerError("failed to delete preference")
	}

	return &notificationNoContentOutput{Status: http.StatusNoContent}, nil
}
