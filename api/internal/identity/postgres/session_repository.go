package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	"github.com/jackc/pgx/v5/pgtype"
)

// SessionRepository adapts sqlc session queries to the identity application boundary.
type SessionRepository struct {
	queries *queries.Queries
}

func NewSessionRepository(q *queries.Queries) *SessionRepository {
	return &SessionRepository{queries: q}
}

func (r *SessionRepository) FindByToken(ctx context.Context, token string) (identity.SessionRecord, error) {
	row, err := r.queries.GetSessionByToken(ctx, token)
	if err != nil {
		return identity.SessionRecord{}, fmt.Errorf("get session by token: %w", err)
	}

	notifications := []string{}
	if row.UserNotifications != nil {
		if err := json.Unmarshal(row.UserNotifications, &notifications); err != nil {
			slog.ErrorContext(ctx, "failed to unmarshal user notifications", "userID", row.UserID_2, "error", err)
			notifications = []string{}
		}
	}
	if notifications == nil {
		notifications = []string{}
	}

	var banExpiresAt *time.Time
	if row.UserBanExpires.Valid {
		banExpiresAt = &row.UserBanExpires.Time
	}

	return identity.SessionRecord{
		Principal: access.Principal{
			Session: access.Session{
				ID:                   row.ID,
				UserID:               row.UserID,
				ExpiresAt:            row.ExpiresAt.Time,
				ActiveOrganizationID: row.ActiveOrganizationID,
				ImpersonatedBy:       row.ImpersonatedBy,
			},
			User: access.User{
				ID:            row.UserID_2,
				Name:          row.UserName,
				Email:         row.UserEmail,
				Image:         row.UserImage,
				Role:          row.UserRole,
				Notifications: notifications,
			},
		},
		Ban: identity.Ban{
			Enabled:   row.UserBanned != nil && *row.UserBanned,
			ExpiresAt: banExpiresAt,
		},
	}, nil
}

func (r *SessionRepository) RefreshExpiry(ctx context.Context, sessionID string, expiresAt time.Time) error {
	if err := r.queries.RefreshSessionExpiry(ctx, queries.RefreshSessionExpiryParams{
		ID:        sessionID,
		ExpiresAt: pgtype.Timestamp{Time: expiresAt, Valid: true},
	}); err != nil {
		return fmt.Errorf("refresh session expiry: %w", err)
	}
	return nil
}
