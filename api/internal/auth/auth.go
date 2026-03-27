package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

func ValidateSession(ctx context.Context, pool *pgxpool.Pool, token string) (*SessionUser, error) {
	row := pool.QueryRow(ctx, `
		SELECT s.id, s.expires_at, s.user_id, s.active_organization_id, s.impersonated_by,
		       u.id, u.name, u.email, u.image, u.role, u.notifications, u.banned
		FROM session s
		JOIN "user" u ON u.id = s.user_id
		WHERE s.token = $1 AND s.expires_at > NOW()
	`, token)

	var su SessionUser
	var notificationsJSON []byte
	err := row.Scan(
		&su.Session.ID, &su.Session.ExpiresAt, &su.Session.UserID,
		&su.Session.ActiveOrganizationID, &su.Session.ImpersonatedBy,
		&su.User.ID, &su.User.Name, &su.User.Email, &su.User.Image,
		&su.User.Role, &notificationsJSON, &su.User.Banned,
	)
	if err != nil {
		return nil, err
	}

	if su.User.Banned != nil && *su.User.Banned {
		return nil, fmt.Errorf("user is banned")
	}

	if notificationsJSON != nil {
		if err := json.Unmarshal(notificationsJSON, &su.User.Notifications); err != nil {
			slog.Error("failed to unmarshal user notifications", "error", err, "userID", su.User.ID)
		}
	}
	if su.User.Notifications == nil {
		su.User.Notifications = []string{}
	}

	return &su, nil
}
