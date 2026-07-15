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
	ID        int32
	UserID    string
	Key       string
	Level     string
	Title     string
	Message   string
	Link      *Link
	Read      bool
	CreatedAt time.Time
}

type Repository interface {
	List(ctx context.Context, userID string) ([]Notification, error)
	DeleteAll(ctx context.Context, userID string) error
	Delete(ctx context.Context, userID string, notificationID int32) error
	MarkAllRead(ctx context.Context, userID string) error
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
