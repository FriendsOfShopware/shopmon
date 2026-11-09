package handler

import (
	"context"
	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
)

type destinationPreferencePath struct {
	OrgID         string `path:"orgId"`
	DestinationID string `path:"destinationId"`
}
type setDestinationPreferenceInput struct {
	OrgID         string `path:"orgId"`
	DestinationID string `path:"destinationId"`
	Body          api.NotificationPreferenceInput
}
type deleteDestinationPreferenceInput struct {
	OrgID         string `path:"orgId"`
	DestinationID string `path:"destinationId"`
	ScopeType     string `query:"scopeType"`
	ScopeID       string `query:"scopeId"`
	EventType     string `query:"eventType"`
	Channel       string `query:"channel"`
}

func (h *Handler) GetDestinationNotificationPreferences(ctx context.Context, input *destinationPreferencePath) (*getNotificationPreferencesOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	rows, err := h.destinations.ListPreferences(ctx, user.ID, input.OrgID, input.DestinationID)
	if err != nil {
		return nil, h.destinationError(ctx, err)
	}
	result := make([]api.NotificationPreference, 0, len(rows))
	for _, p := range rows {
		result = append(result, api.NotificationPreference{ScopeType: p.ScopeType, ScopeId: p.ScopeID, EventType: p.EventType, Channel: p.Channel, Enabled: p.Enabled})
	}
	return &getNotificationPreferencesOutput{Body: result}, nil
}
func (h *Handler) SetDestinationNotificationPreference(ctx context.Context, input *setDestinationPreferenceInput) (*struct{}, error) {
	pref := notification.Preference{ScopeType: input.Body.ScopeType, Channel: input.Body.Channel, Enabled: input.Body.Enabled}
	if input.Body.ScopeId != nil {
		pref.ScopeID = *input.Body.ScopeId
	}
	if input.Body.EventType != nil {
		pref.EventType = *input.Body.EventType
	}
	return h.changeDestinationPreference(ctx, input.OrgID, input.DestinationID, pref, false)
}
func (h *Handler) DeleteDestinationNotificationPreference(ctx context.Context, input *deleteDestinationPreferenceInput) (*struct{}, error) {
	return h.changeDestinationPreference(ctx, input.OrgID, input.DestinationID, notification.Preference{ScopeType: input.ScopeType, ScopeID: input.ScopeID, EventType: input.EventType, Channel: input.Channel}, true)
}
func (h *Handler) changeDestinationPreference(ctx context.Context, orgID, id string, pref notification.Preference, inherit bool) (*struct{}, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}
	known := pref.EventType == ""
	for _, event := range notify.ListEventTypes() {
		if event.Type == pref.EventType {
			known = true
			break
		}
	}
	if !known {
		return nil, huma.Error400BadRequest("invalid notification event type")
	}
	if err := h.destinations.SetPreference(ctx, user.ID, orgID, id, pref, inherit); err != nil {
		return nil, h.destinationError(ctx, err)
	}
	return nil, nil
}
