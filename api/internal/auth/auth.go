package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
)

type User struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Image         *string  `json:"image"`
	Role          string   `json:"role"`
	Notifications []string `json:"notifications"`
	Banned        *bool    `json:"banned"`
}

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

	if su.User.Banned != nil && *su.User.Banned {
		return nil, fmt.Errorf("user is banned")
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
