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
			ID:         row.ID,
			UserID:     row.UserID,
			Key:        row.Key,
			Level:      row.Level,
			Title:      row.Title,
			Message:    row.Message,
			TitleKey:   row.TitleKey,
			MessageKey: row.MessageKey,
			Read:       row.Read,
			CreatedAt:  row.CreatedAt.Time,
		}
		if len(row.Params) > 0 {
			var params map[string]any
			if err := json.Unmarshal(row.Params, &params); err != nil {
				return nil, fmt.Errorf("decode notification %d params: %w", row.ID, err)
			}
			if len(params) > 0 {
				item.Params = params
			}
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

func (r *Repository) ListPreferences(ctx context.Context, userID string) ([]notification.Preference, error) {
	rows, err := r.queries.ListNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notification preferences: %w", err)
	}
	result := make([]notification.Preference, 0, len(rows))
	for _, row := range rows {
		result = append(result, notification.Preference{
			ScopeType: row.ScopeType,
			ScopeID:   row.ScopeID,
			EventType: row.EventType,
			Channel:   row.Channel,
			Enabled:   row.Enabled,
		})
	}
	return result, nil
}

func (r *Repository) SetPreference(ctx context.Context, userID string, pref notification.Preference) error {
	if err := r.queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID:    userID,
		ScopeType: pref.ScopeType,
		ScopeID:   pref.ScopeID,
		EventType: pref.EventType,
		Channel:   pref.Channel,
		Enabled:   pref.Enabled,
		Config:    json.RawMessage(`{}`),
	}); err != nil {
		return fmt.Errorf("upsert notification preference: %w", err)
	}
	return nil
}

func (r *Repository) DeletePreference(ctx context.Context, userID, scopeType, scopeID, eventType, channel string) error {
	if err := r.queries.DeleteNotificationPreference(ctx, queries.DeleteNotificationPreferenceParams{
		UserID:    userID,
		ScopeType: scopeType,
		ScopeID:   scopeID,
		EventType: eventType,
		Channel:   channel,
	}); err != nil {
		return fmt.Errorf("delete notification preference: %w", err)
	}
	return nil
}
