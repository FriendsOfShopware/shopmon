package identity

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/access"
)

// SessionLifetime is the validity window granted to a regular session token.
const SessionLifetime = 30 * 24 * time.Hour

// ErrUserBanned reports that a valid session belongs to a currently banned user.
var ErrUserBanned = errors.New("user is banned")

// Ban is the account-access restriction associated with an identity.
type Ban struct {
	Enabled   bool
	ExpiresAt *time.Time
}

// ActiveAt reports whether the ban blocks access at the supplied time.
func (b Ban) ActiveAt(now time.Time) bool {
	return b.Enabled && (b.ExpiresAt == nil || !b.ExpiresAt.Before(now))
}

// SessionRecord is the persistence-neutral result needed for authentication.
type SessionRecord struct {
	Principal access.Principal
	Ban       Ban
}

// SessionRepository is the persistence required to authenticate and rotate sessions.
type SessionRepository interface {
	FindByToken(ctx context.Context, token string) (SessionRecord, error)
	RefreshExpiry(ctx context.Context, sessionID string, expiresAt time.Time) error
}

// SessionAuthenticator resolves tokens into request principals.
type SessionAuthenticator struct {
	sessions SessionRepository
}

func NewSessionAuthenticator(sessions SessionRepository) *SessionAuthenticator {
	return &SessionAuthenticator{sessions: sessions}
}

// Authenticate validates a session and returns the principal to attach to the
// request. Regular sessions are rotated once they enter the second half of
// their lifetime; deliberately short-lived impersonation sessions are not.
func (a *SessionAuthenticator) Authenticate(ctx context.Context, token string) (*access.Principal, error) {
	record, err := a.sessions.FindByToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("find session: %w", err)
	}

	now := time.Now()
	if record.Ban.ActiveAt(now) {
		return nil, ErrUserBanned
	}

	principal := record.Principal

	if principal.Session.ImpersonatedBy == nil && principal.Session.ExpiresAt.Sub(now) < SessionLifetime/2 {
		newExpiry := now.Add(SessionLifetime)
		if err := a.sessions.RefreshExpiry(ctx, principal.Session.ID, newExpiry); err != nil {
			slog.ErrorContext(ctx, "failed to refresh session expiry", "sessionID", principal.Session.ID, "error", err)
		} else {
			principal.Session.ExpiresAt = newExpiry
		}
	}

	return &principal, nil
}
