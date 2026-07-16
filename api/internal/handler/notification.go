package handler

import (
	"log/slog"
	"net/http"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GetNotifications returns all notifications for the current user.
func (h *Handler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	user := h.requireUser(w, r)
	if user == nil {
		return
	}

	rows, err := h.notifications.List(r.Context(), user.ID)
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
			CreatedAt: row.CreatedAt,
		}

		if row.Link != nil {
			n.Link = &api.NotificationLink{Url: row.Link.URL, Label: row.Link.Label}
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

	if err := h.notifications.DeleteAll(r.Context(), user.ID); err != nil {
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

	if err := h.notifications.Delete(r.Context(), user.ID, int32(id)); err != nil {
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

	if err := h.notifications.MarkAllRead(r.Context(), user.ID); err != nil {
		slog.Error("failed to mark notifications read", "error", err)
		httputil.WriteError(w, http.StatusInternalServerError, "failed to mark notifications read")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
