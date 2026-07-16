package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SessionMetadata struct {
	IPAddress string
	UserAgent string
}

type NewSession struct {
	ID        string
	Token     string
	UserID    string
	ExpiresAt time.Time
	IPAddress *string
	UserAgent *string
}

type SessionLifecycleRepository interface {
	Create(ctx context.Context, session NewSession) error
	Delete(ctx context.Context, token string) error
}

type SessionService struct {
	repository SessionLifecycleRepository
	now        func() time.Time
}

func NewSessionService(repository SessionLifecycleRepository) *SessionService {
	return &SessionService{repository: repository, now: time.Now}
}

func (s *SessionService) Issue(ctx context.Context, userID string, metadata SessionMetadata) (string, error) {
	token, err := randomToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	var ipAddress, userAgent *string
	if metadata.IPAddress != "" {
		ipAddress = &metadata.IPAddress
	}
	if metadata.UserAgent != "" {
		userAgent = &metadata.UserAgent
	}
	if err := s.repository.Create(ctx, NewSession{
		ID: uuid.NewString(), Token: token, UserID: userID,
		ExpiresAt: s.now().Add(SessionLifetime), IPAddress: ipAddress, UserAgent: userAgent,
	}); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.repository.Delete(ctx, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
