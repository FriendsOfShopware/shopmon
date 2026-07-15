package access

import (
	"context"
	"time"
)

// User is the authenticated user information needed by request authorization.
// It deliberately excludes authentication-provider and persistence details.
type User struct {
	ID            string   `json:"id"`
	Name          string   `json:"name"`
	Email         string   `json:"email"`
	Image         *string  `json:"image"`
	Role          string   `json:"role"`
	Notifications []string `json:"notifications"`
}

// Session is the authenticated session information needed during a request.
type Session struct {
	ID                   string    `json:"id"`
	UserID               string    `json:"userId"`
	ExpiresAt            time.Time `json:"expiresAt"`
	ActiveOrganizationID *string   `json:"activeOrganizationId"`
	ImpersonatedBy       *string   `json:"impersonatedBy"`
}

// Principal is the single authenticated identity attached to a request.
type Principal struct {
	User    User
	Session Session
}

type principalContextKey struct{}

// WithPrincipal attaches an authenticated principal to a context.
func WithPrincipal(ctx context.Context, principal *Principal) context.Context {
	return context.WithValue(ctx, principalContextKey{}, principal)
}

// PrincipalFromContext returns the authenticated request principal, if any.
func PrincipalFromContext(ctx context.Context) *Principal {
	principal, _ := ctx.Value(principalContextKey{}).(*Principal)
	return principal
}

// UserFromContext returns the authenticated request user, if any.
func UserFromContext(ctx context.Context) *User {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return nil
	}
	return &principal.User
}

// SessionFromContext returns the authenticated request session, if any.
func SessionFromContext(ctx context.Context) *Session {
	principal := PrincipalFromContext(ctx)
	if principal == nil {
		return nil
	}
	return &principal.Session
}
