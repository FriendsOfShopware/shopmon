package postgres

import (
	"context"
	"fmt"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
)

type SessionLifecycleRepository struct {
	queries *queries.Queries
}

func NewSessionLifecycleRepository(q *queries.Queries) *SessionLifecycleRepository {
	return &SessionLifecycleRepository{queries: q}
}

func (r *SessionLifecycleRepository) Create(ctx context.Context, session identity.NewSession) error {
	if _, err := r.queries.CreateSession(ctx, queries.CreateSessionParams{
		ID: session.ID, ExpiresAt: pgtype.Timestamp{Time: session.ExpiresAt, Valid: true},
		Token: session.Token, IpAddress: session.IPAddress, UserAgent: session.UserAgent, UserID: session.UserID,
	}); err != nil {
		return fmt.Errorf("insert session: %w", err)
	}
	return nil
}

func (r *SessionLifecycleRepository) Delete(ctx context.Context, token string) error {
	if err := r.queries.DeleteSession(ctx, token); err != nil {
		return fmt.Errorf("execute session delete: %w", err)
	}
	return nil
}
