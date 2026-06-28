package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
)

// GetNotificationEventTypes lists the notifiable event types (the source of
// truth for the preferences UI), so new event types appear without a frontend
// change.
func (h *Handler) GetNotificationEventTypes(w http.ResponseWriter, r *http.Request) {
	if user := h.requireUser(w, r); user == nil {
		return
	}

	types := notify.ListEventTypes()
	result := make([]api.NotificationEventType, 0, len(types))
	for _, t := range types {
		result = append(result, api.NotificationEventType{
			Type:            t.Type,
			DefaultChannels: t.DefaultChannels,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// validPreferenceScopes and validPreferenceChannels bound the values a client
// may store. The empty channel is the subscription marker and is managed via
// the subscribe endpoints, so it is not settable here.
var (
	validPreferenceScopes   = map[string]bool{"global": true, "organization": true, "environment": true}
	validPreferenceChannels = map[string]bool{"in_app": true, "email": true}
)

// GetNotificationPreferences returns the current user's notification preferences.
func (h *Handler) GetNotificationPreferences(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.ListNotificationPreferences(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list notification preferences", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get preferences")
		return
	}

	result := make([]api.NotificationPreference, 0, len(rows))
	for _, row := range rows {
		result = append(result, api.NotificationPreference{
			ScopeType: row.ScopeType,
			ScopeId:   row.ScopeID,
			EventType: row.EventType,
			Channel:   row.Channel,
			Enabled:   row.Enabled,
		})
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// SetNotificationPreference creates or updates a single notification preference.
func (h *Handler) SetNotificationPreference(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	var body api.SetNotificationPreferenceJSONRequestBody
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		httputil.WriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	if !validPreferenceScopes[body.ScopeType] {
		httputil.WriteError(w, http.StatusBadRequest, "invalid scope type")
		return
	}
	if !validPreferenceChannels[body.Channel] {
		httputil.WriteError(w, http.StatusBadRequest, "invalid channel")
		return
	}

	scopeID := ""
	if body.ScopeId != nil {
		scopeID = *body.ScopeId
	}
	eventType := ""
	if body.EventType != nil {
		eventType = *body.EventType
	}
	// A global preference has no scope id; a scoped one requires it.
	if body.ScopeType == "global" {
		scopeID = ""
	} else if scopeID == "" {
		httputil.WriteError(w, http.StatusBadRequest, "scope id required for non-global scope")
		return
	}

	if err := h.queries.UpsertNotificationPreference(r.Context(), queries.UpsertNotificationPreferenceParams{
		UserID:    user.ID,
		ScopeType: body.ScopeType,
		ScopeID:   scopeID,
		EventType: eventType,
		Channel:   body.Channel,
		Enabled:   body.Enabled,
		Config:    json.RawMessage("{}"),
	}); err != nil {
		slog.Error("failed to upsert notification preference", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to save preference")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteNotificationPreference removes a single preference, reverting that scope
// to its inherited/default behaviour.
func (h *Handler) DeleteNotificationPreference(w http.ResponseWriter, r *http.Request, params api.DeleteNotificationPreferenceParams) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if !validPreferenceScopes[params.ScopeType] {
		httputil.WriteError(w, http.StatusBadRequest, "invalid scope type")
		return
	}

	scopeID := ""
	if params.ScopeId != nil {
		scopeID = *params.ScopeId
	}
	eventType := ""
	if params.EventType != nil {
		eventType = *params.EventType
	}

	if err := h.queries.DeleteNotificationPreference(r.Context(), queries.DeleteNotificationPreferenceParams{
		UserID:    user.ID,
		ScopeType: params.ScopeType,
		ScopeID:   scopeID,
		EventType: eventType,
		Channel:   params.Channel,
	}); err != nil {
		slog.Error("failed to delete notification preference", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete preference")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
