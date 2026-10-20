package identity

import (
	"context"
	"fmt"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionMetadata struct {
	IPAddress string
	UserAgent string
}

type SessionService struct {
	queries *queries.Queries
	now     func() time.Time
}

func NewSessionService(q *queries.Queries) *SessionService {
	return &SessionService{queries: q, now: time.Now}
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
	if _, err := s.queries.CreateSession(ctx, queries.CreateSessionParams{
		ID:        uuid.NewString(),
		ExpiresAt: pgtype.Timestamp{Time: s.now().Add(SessionLifetime), Valid: true},
		Token:     token,
		IpAddress: ipAddress,
		UserAgent: userAgent,
		UserID:    userID,
	}); err != nil {
		return "", fmt.Errorf("create session: %w", err)
	}
	return token, nil
}

func (s *SessionService) Delete(ctx context.Context, token string) error {
	if token == "" {
		return nil
	}
	if err := s.queries.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}
