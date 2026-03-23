package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/api"
)

// GetNotifications returns all notifications for the current user.
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.queries.ListNotifications(r.Context(), user.ID)
	if err != nil {
		slog.Error("failed to list notifications", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to get notifications")
		return
	}

	result := make([]api.Notification, 0, len(rows))
	for _, row := range rows {
		n := api.Notification{
			Id:        int(row.ID),
			UserId:    row.UserID,
			Key:       row.Key,
			Level:     row.Level,
			Title:     row.Title,
			Message:   row.Message,
			Read:      row.Read,
			CreatedAt: pgtimeToTime(row.CreatedAt),
		}

		if row.Link != nil && len(row.Link) > 0 {
			var link api.NotificationLink
			if err := json.Unmarshal(row.Link, &link); err == nil && link.Url != "" {
				n.Link = &link
			}
		}

		result = append(result, n)
	}

	httputil.WriteJSON(w, http.StatusOK, result)
}

// DeleteAllNotifications deletes all notifications for the current user.
func (h *Handler) DeleteAllNotifications(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.queries.DeleteAllNotifications(r.Context(), user.ID); err != nil {
		slog.Error("failed to delete all notifications", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete notifications")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// DeleteNotification deletes a single notification.
func (h *Handler) DeleteNotification(w http.ResponseWriter, r *http.Request, id api.NotificationId) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.queries.DeleteNotification(r.Context(), queries.DeleteNotificationParams{
		ID:     int32(id),
		UserID: user.ID,
	}); err != nil {
		slog.Error("failed to delete notification", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to delete notification")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// MarkNotificationsRead marks all notifications as read for the current user.
func (h *Handler) MarkNotificationsRead(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	if err := h.queries.MarkAllNotificationsRead(r.Context(), user.ID); err != nil {
		slog.Error("failed to mark notifications read", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to mark notifications read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
