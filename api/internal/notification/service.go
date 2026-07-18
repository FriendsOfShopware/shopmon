package notification

import (
	"context"
	"fmt"
	"time"
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

type Repository interface {
	List(ctx context.Context, userID string) ([]Notification, error)
	DeleteAll(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string, notificationID int32) error
	MarkAllRead(ctx context.Context, userID string) error
	ListPreferences(ctx context.Context, userID string) ([]Preference, error)
	SetPreference(ctx context.Context, userID string, pref Preference) error
	DeletePreference(ctx context.Context, userID, scopeType, scopeID, eventType, channel string) error
}

type Service struct {
	repository Repository
}

func NewService(repository Repository) *Service {
	return &Service{repository: repository}
}

func (s *Service) List(ctx context.Context, userID string) ([]Notification, error) {
	notifications, err := s.repository.List(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notifications: %w", err)
	}
	return notifications, nil
}

func (s *Service) DeleteAll(ctx context.Context, userID string) error {
	if err := s.repository.DeleteAll(ctx, userID); err != nil {
		return fmt.Errorf("delete all notifications: %w", err)
	}
	return nil
}

func (s *Service) Delete(ctx context.Context, userID string, notificationID int32) error {
	if err := s.repository.Delete(ctx, userID, notificationID); err != nil {
		return fmt.Errorf("delete notification: %w", err)
	}
	return nil
}

func (s *Service) MarkAllRead(ctx context.Context, userID string) error {
	if err := s.repository.MarkAllRead(ctx, userID); err != nil {
		return fmt.Errorf("mark all notifications read: %w", err)
	}
	return nil
}

func (s *Service) ListPreferences(ctx context.Context, userID string) ([]Preference, error) {
	prefs, err := s.repository.ListPreferences(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list notification preferences: %w", err)
	}
	return prefs, nil
}

func (s *Service) SetPreference(ctx context.Context, userID string, pref Preference) error {
	if err := s.repository.SetPreference(ctx, userID, pref); err != nil {
		return fmt.Errorf("set notification preference: %w", err)
	}
	return nil
}

func (s *Service) DeletePreference(ctx context.Context, userID, scopeType, scopeID, eventType, channel string) error {
	if err := s.repository.DeletePreference(ctx, userID, scopeType, scopeID, eventType, channel); err != nil {
		return fmt.Errorf("delete notification preference: %w", err)
	}
	return nil
}
