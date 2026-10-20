package notification

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type Link struct {
	URL   string
	Label string
}

type Notification struct {
	ID         int32
	UserID     string
	Key        string
	Level      string
	Title      string
	Message    string
	TitleKey   *string
	MessageKey *string
	Params     map[string]any
	Link       *Link
	Read       bool
	CreatedAt  time.Time
}

type Preference struct {
	ScopeType string
	ScopeID   string
	EventType string
	Channel   string
	Enabled   bool
}

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

func (s *Service) List(ctx context.Context, userID string) ([]Notification, error) {
	rows, err := s.queries.ListNotifications(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	result := make([]Notification, 0, len(rows))
	for _, row := range rows {
		item := Notification{
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
				item.Link = &Link{URL: link.URL, Label: link.Label}
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *Service) DeleteAll(ctx context.Context, userID string) error {
	if err := s.queries.DeleteAllNotifications(ctx, userID); err != nil {
		return fmt.Errorf("delete all notifications: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID string, notificationID int32) error {
	if err := s.queries.DeleteNotification(ctx, queries.DeleteNotificationParams{ID: notificationID, UserID: userID}); err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	if err := s.queries.MarkAllNotificationsRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *Service) ListPreferences(ctx context.Context, userID string) ([]Preference, error) {
	rows, err := s.queries.ListNotificationPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notification preferences: %w", err)
	}
	result := make([]Preference, 0, len(rows))
	for _, row := range rows {
		result = append(result, Preference{
			ScopeType: row.ScopeType,
			ScopeID:   row.ScopeID,
			EventType: row.EventType,
			Channel:   row.Channel,
			Enabled:   row.Enabled,
		})
	}
	return result, nil
}

func (s *Service) SetPreference(ctx context.Context, userID string, pref Preference) error {
	if err := s.queries.UpsertNotificationPreference(ctx, queries.UpsertNotificationPreferenceParams{
		UserID:    userID,
		ScopeType: pref.ScopeType,
		ScopeID:   pref.ScopeID,
		EventType: pref.EventType,
		Channel:   pref.Channel,
		Enabled:   pref.Enabled,
		Config:    json.RawMessage(`{}`),
	}); err != nil {
		return fmt.Errorf("set notification preference: %w", err)
	}
	return nil
}

func (s *Service) DeletePreference(ctx context.Context, userID, scopeType, scopeID, eventType, channel string) error {
	if err := s.queries.DeleteNotificationPreference(ctx, queries.DeleteNotificationPreferenceParams{
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
