package handler

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
)

func registerNotifications(api huma.API, h *Handler) {
	huma.Register(api, huma.Operation{
		OperationID: "getNotifications",
		Method:      http.MethodGet,
		Path:        "/notifications",
		Summary:     "Get all notifications for the current user",
		Tags:        []string{"Notifications"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetNotifications)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteAllNotifications",
		Method:        http.MethodDelete,
		Path:          "/notifications",
		Summary:       "Delete all notifications",
		Tags:          []string{"Notifications"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized},
	}, h.DeleteAllNotifications)

	huma.Register(api, huma.Operation{
		OperationID: "getNotificationEventTypes",
		Method:      http.MethodGet,
		Path:        "/notifications/event-types",
		Summary:     "List the notifiable event types and their default channels",
		Tags:        []string{"Notifications"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetNotificationEventTypes)

	huma.Register(api, huma.Operation{
		OperationID:   "markNotificationsRead",
		Method:        http.MethodPost,
		Path:          "/notifications/mark-read",
		Summary:       "Mark all notifications as read",
		Tags:          []string{"Notifications"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized},
	}, h.MarkNotificationsRead)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteNotification",
		Method:        http.MethodDelete,
		Path:          "/notifications/{id}",
		Summary:       "Delete a single notification",
		Tags:          []string{"Notifications"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusUnauthorized, http.StatusNotFound},
	}, h.DeleteNotification)

	huma.Register(api, huma.Operation{
		OperationID: "getNotificationPreferences",
		Method:      http.MethodGet,
		Path:        "/account/notification-preferences",
		Summary:     "Get the current user's notification preferences",
		Tags:        []string{"Account"},
		Errors:      []int{http.StatusUnauthorized},
	}, h.GetNotificationPreferences)

	huma.Register(api, huma.Operation{
		OperationID:   "setNotificationPreference",
		Method:        http.MethodPut,
		Path:          "/account/notification-preferences",
		Summary:       "Create or update a single notification preference",
		Tags:          []string{"Account"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized},
	}, h.SetNotificationPreference)

	huma.Register(api, huma.Operation{
		OperationID:   "deleteNotificationPreference",
		Method:        http.MethodDelete,
		Path:          "/account/notification-preferences",
		Summary:       "Delete a single notification preference (revert to inherited/default)",
		Tags:          []string{"Account"},
		DefaultStatus: http.StatusNoContent,
		Errors:        []int{http.StatusBadRequest, http.StatusUnauthorized},
	}, h.DeleteNotificationPreference)
}
