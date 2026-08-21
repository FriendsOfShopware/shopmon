package handler

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/friendsofshopware/shopmon/api/internal/api"
)

type getNotificationsOutput struct {
	Body []api.Notification
}

type deleteNotificationInput struct {
	ID int `path:"id"`
}

type notificationNoContentOutput struct {
	Status int
}

func (h *Handler) GetNotifications(ctx context.Context, _ *struct{}) (*getNotificationsOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	rows, err := h.notifications.List(ctx, user.ID)
	if err != nil {
		slog.ErrorContext(ctx, "failed to list notifications", "error", err)
		return nil, huma.Error500InternalServerError("failed to get notifications")
	}

	result := make([]api.Notification, 0, len(rows))
	for _, row := range rows {
		n := api.Notification{
			Id:         int(row.ID),
			UserId:     row.UserID,
			Key:        row.Key,
			Level:      row.Level,
			Title:      row.Title,
			Message:    row.Message,
			TitleKey:   row.TitleKey,
			MessageKey: row.MessageKey,
			Read:       row.Read,
			CreatedAt:  row.CreatedAt,
		}

		if len(row.Params) > 0 {
			params := map[string]interface{}(row.Params)
			if len(params) > 0 {
				n.Params = &params
			}
		}

		if row.Link != nil {
			n.Link = &api.NotificationLink{Url: row.Link.URL, Label: row.Link.Label}
		}

		result = append(result, n)
	}

	return &getNotificationsOutput{Body: result}, nil
}

func (h *Handler) DeleteAllNotifications(ctx context.Context, _ *struct{}) (*notificationNoContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.notifications.DeleteAll(ctx, user.ID); err != nil {
		slog.ErrorContext(ctx, "failed to delete all notifications", "error", err)
		return nil, huma.Error500InternalServerError("failed to delete notifications")
	}

	return &notificationNoContentOutput{Status: http.StatusNoContent}, nil
}

func (h *Handler) DeleteNotification(ctx context.Context, input *deleteNotificationInput) (*notificationNoContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.notifications.Delete(ctx, user.ID, int32(input.ID)); err != nil {
		slog.ErrorContext(ctx, "failed to delete notification", "error", err)
		return nil, huma.Error500InternalServerError("failed to delete notification")
	}

	return &notificationNoContentOutput{Status: http.StatusNoContent}, nil
}

func (h *Handler) MarkNotificationsRead(ctx context.Context, _ *struct{}) (*notificationNoContentOutput, error) {
	user, err := h.requireUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := h.notifications.MarkAllRead(ctx, user.ID); err != nil {
		slog.ErrorContext(ctx, "failed to mark notifications read", "error", err)
		return nil, huma.Error500InternalServerError("failed to mark notifications read")
	}

	return &notificationNoContentOutput{Status: http.StatusNoContent}, nil
}
