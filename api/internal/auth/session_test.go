package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestValidateSessionRotatesExpiry verifies that validating a session whose
// remaining lifetime has dropped below half the window slides its expiry
// forward, keeping actively used sessions alive.
func TestValidateSessionRotatesExpiry(t *testing.T) {
	env := testutil.Setup(t)

	// SeedUser creates a session expiring in 24h, well inside the rotation
	// window, so validating it must extend the expiry to ~30 days out.
	token := env.SeedUser(t, "rot-user", "Rot User", "rot@example.com", "user")

	su, err := auth.ValidateSession(context.Background(), env.Queries, token)
	require.NoError(t, err)

	assert.WithinDuration(t, time.Now().Add(30*24*time.Hour), su.Session.ExpiresAt, time.Minute,
		"expiry should slide forward by the full session lifetime")
}

// TestValidateSessionKeepsFreshExpiry verifies that a session still in the
// first half of its lifetime is left untouched.
func TestValidateSessionKeepsFreshExpiry(t *testing.T) {
	env := testutil.Setup(t)

	token := env.SeedUser(t, "fresh-user", "Fresh User", "fresh@example.com", "user")

	// Push the expiry far out so the session is well inside its first half.
	// The session column is a bare timestamp (no zone), so compare in UTC.
	freshExpiry := time.Now().UTC().Add(29 * 24 * time.Hour)
	_, err := env.Pool.Exec(context.Background(),
		`UPDATE session SET expires_at = $1 WHERE user_id = $2`, freshExpiry, "fresh-user")
	require.NoError(t, err)

	su, err := auth.ValidateSession(context.Background(), env.Queries, token)
	require.NoError(t, err)

	assert.WithinDuration(t, freshExpiry, su.Session.ExpiresAt.UTC(), time.Minute,
		"a fresh session should not be rotated")
}

// TestValidateSessionDoesNotRotateImpersonation verifies that a short-lived
// admin impersonation session is never slid forward: rotating it would extend
// its deliberate one-hour security bound to a full session lifetime.
func TestValidateSessionDoesNotRotateImpersonation(t *testing.T) {
	env := testutil.Setup(t)

	env.SeedUser(t, "imp-target", "Imp Target", "imp-target@example.com", "user")

	// Insert an impersonation session expiring in one hour (well inside the
	// rotation window) with impersonated_by set.
	impExpiry := time.Now().UTC().Add(time.Hour)
	now := time.Now().UTC()
	_, err := env.Pool.Exec(context.Background(), `
		INSERT INTO session (id, expires_at, token, created_at, updated_at, user_id, impersonated_by)
		VALUES ($1, $2, $3, $4, $4, $5, $6)
	`, "imp-session", impExpiry, "imp-token", now, "imp-target", "some-admin")
	require.NoError(t, err)

	su, err := auth.ValidateSession(context.Background(), env.Queries, "imp-token")
	require.NoError(t, err)

	assert.WithinDuration(t, impExpiry, su.Session.ExpiresAt.UTC(), time.Minute,
		"an impersonation session must not be rotated")
}
