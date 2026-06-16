package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5/pgtype"
)

// isBanActive reports whether a ban should still block access. A ban with a
// ban_expires timestamp in the past is treated as expired and no longer blocks.
func isBanActive(banned *bool, banExpires pgtype.Timestamp) bool {
	if banned == nil || !*banned {
		return false
	}
	if banExpires.Valid && banExpires.Time.Before(time.Now()) {
		return false
	}
	return true
}

type User struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Image         *string  `json:"image"`
	Role          string   `json:"role"`
	Notifications []string `json:"notifications"`
	Banned        *bool    `json:"banned"`
}

// sessionLifetime is the validity window granted to a session token. A session
// is rotated (its expiry slid forward) once it enters the second half of this
// window, so an actively used session never expires out from under the user
// while an abandoned one still lapses within sessionLifetime.
const sessionLifetime = 30 * 24 * time.Hour

type Session struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"userId"`
	ExpiresAt            time.Time `json:"expiresAt"`
	ActiveOrganizationID *string   `json:"activeOrganizationId"`
	ImpersonatedBy       *string   `json:"impersonatedBy"`
}

type SessionUser struct {
	User    User
	Session Session
}

func ValidateSession(ctx context.Context, q *queries.Queries, token string) (*SessionUser, error) {
	row, err := q.GetSessionByToken(ctx, token)
	if err != nil {
		return nil, err
	}

	su := SessionUser{
		Session: Session{
			ID:                   row.ID,
			UserID:               row.UserID,
			ExpiresAt:            row.ExpiresAt.Time,
			ActiveOrganizationID: row.ActiveOrganizationID,
			ImpersonatedBy:       row.ImpersonatedBy,
		},
		User: User{
			ID:     row.UserID_2,
			Name:   row.UserName,
			Email:  row.UserEmail,
			Image:  row.UserImage,
			Role:   row.UserRole,
			Banned: row.UserBanned,
		},
	}

	if isBanActive(su.User.Banned, row.UserBanExpires) {
		return nil, fmt.Errorf("user is banned")
	}

	// Slide the expiry forward once the session is past the halfway point of its
	// lifetime, keeping actively used sessions alive. A best-effort failure here
	// must not block an otherwise valid session. Impersonation sessions are
	// deliberately short-lived (see AdminImpersonate) and must never be rotated,
	// or their one-hour security bound would be extended to a full lifetime.
	if su.Session.ImpersonatedBy == nil && time.Until(su.Session.ExpiresAt) < sessionLifetime/2 {
		newExpiry := time.Now().Add(sessionLifetime)
		if err := q.RefreshSessionExpiry(ctx, queries.RefreshSessionExpiryParams{
			ID:        su.Session.ID,
			ExpiresAt: pgtype.Timestamp{Time: newExpiry, Valid: true},
		}); err != nil {
			slog.Error("failed to refresh session expiry", "error", err, "sessionID", su.Session.ID)
		} else {
			su.Session.ExpiresAt = newExpiry
		}
	}

	if row.UserNotifications != nil {
		if err := json.Unmarshal(row.UserNotifications, &su.User.Notifications); err != nil {
			slog.Error("failed to unmarshal user notifications", "error", err, "userID", su.User.ID)
		}
	}
	if su.User.Notifications == nil {
		su.User.Notifications = []string{}
	}

	return &su, nil
}
