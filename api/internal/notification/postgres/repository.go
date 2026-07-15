package postgres

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
)

type Repository struct {
	queries *queries.Queries
}

var _ notification.Repository = (*Repository)(nil)

func NewRepository(q *queries.Queries) *Repository {
	return &Repository{queries: q}
}

func (r *Repository) List(ctx context.Context, userID string) ([]notification.Notification, error) {
	rows, err := r.queries.ListNotifications(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	result := make([]notification.Notification, 0, len(rows))
	for _, row := range rows {
		item := notification.Notification{
			ID:        row.ID,
			UserID:    row.UserID,
			Key:       row.Key,
			Level:     row.Level,
			Title:     row.Title,
			Message:   row.Message,
			Read:      row.Read,
			CreatedAt: row.CreatedAt.Time,
		}
		if len(row.Link) > 0 {
			var link struct {
				URL   string `json:"url"`
				Label string `json:"label"`
			}
			if err := json.Unmarshal(row.Link, &link); err != nil {
				return nil, fmt.Errorf("decode notification %d link: %w", row.ID, err)
			}
			if link.URL != "" {
				item.Link = &notification.Link{URL: link.URL, Label: link.Label}
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (r *Repository) DeleteAll(ctx context.Context, userID string) error {
	if err := r.queries.DeleteAllNotifications(ctx, userID); err != nil {
		return fmt.Errorf("delete all notifications: %w", err)
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, userID string, notificationID int32) error {
	if err := r.queries.DeleteNotification(ctx, queries.DeleteNotificationParams{ID: notificationID, UserID: userID}); err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

func (r *Repository) MarkAllRead(ctx context.Context, userID string) error {
	if err := r.queries.MarkAllNotificationsRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}
