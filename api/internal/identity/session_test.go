package identity

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSessionRepository struct {
	record          SessionRecord
	findErr         error
	refreshErr      error
	findCalls       int
	refreshCalls    int
	token           string
	refreshedID     string
	refreshedExpiry time.Time
}

func (r *stubSessionRepository) FindByToken(_ context.Context, token string) (SessionRecord, error) {
	r.findCalls++
	r.token = token
	return r.record, r.findErr
}

func (r *stubSessionRepository) RefreshExpiry(_ context.Context, sessionID string, expiresAt time.Time) error {
	r.refreshCalls++
	r.refreshedID = sessionID
	r.refreshedExpiry = expiresAt
	return r.refreshErr
}

func TestSessionAuthenticatorBanPolicy(t *testing.T) {
	now := time.Now()
	future := now.Add(time.Hour)
	past := now.Add(-time.Hour)

	tests := []struct {
		name      string
		ban       Ban
		wantError bool
	}{
		{name: "permanent ban", ban: Ban{Enabled: true}, wantError: true},
		{name: "unexpired ban", ban: Ban{Enabled: true, ExpiresAt: &future}, wantError: true},
		{name: "expired ban", ban: Ban{Enabled: true, ExpiresAt: &past}},
		{name: "not banned", ban: Ban{}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repository := &stubSessionRepository{record: SessionRecord{
				Principal: access.Principal{Session: access.Session{ExpiresAt: now.Add(29 * 24 * time.Hour)}},
				Ban:       tt.ban,
			}}

			principal, err := NewSessionAuthenticator(repository).Authenticate(t.Context(), "token")

			if tt.wantError {
				require.ErrorIs(t, err, ErrUserBanned)
				assert.Nil(t, principal)
				return
			}
			require.NoError(t, err)
			assert.NotNil(t, principal)
		})
	}
}

func TestSessionAuthenticatorRotatesRegularSession(t *testing.T) {
	repository := &stubSessionRepository{record: SessionRecord{
		Principal: access.Principal{Session: access.Session{
			ID:        "session-1",
			ExpiresAt: time.Now().Add(time.Hour),
		}},
	}}

	principal, err := NewSessionAuthenticator(repository).Authenticate(t.Context(), "token-1")

	require.NoError(t, err)
	assert.Equal(t, 1, repository.findCalls)
	assert.Equal(t, "token-1", repository.token)
	assert.Equal(t, 1, repository.refreshCalls)
	assert.Equal(t, "session-1", repository.refreshedID)
	assert.WithinDuration(t, time.Now().Add(SessionLifetime), repository.refreshedExpiry, time.Second)
	assert.Equal(t, repository.refreshedExpiry, principal.Session.ExpiresAt)
}

func TestSessionAuthenticatorDoesNotRotateImpersonationSession(t *testing.T) {
	adminID := "admin-1"
	originalExpiry := time.Now().Add(time.Hour)
	repository := &stubSessionRepository{record: SessionRecord{
		Principal: access.Principal{Session: access.Session{
			ID:             "session-1",
			ExpiresAt:      originalExpiry,
			ImpersonatedBy: &adminID,
		}},
	}}

	principal, err := NewSessionAuthenticator(repository).Authenticate(t.Context(), "token")

	require.NoError(t, err)
	assert.Equal(t, 0, repository.refreshCalls)
	assert.Equal(t, originalExpiry, principal.Session.ExpiresAt)
}

func TestSessionAuthenticatorWrapsRepositoryError(t *testing.T) {
	repositoryErr := errors.New("database unavailable")
	repository := &stubSessionRepository{findErr: repositoryErr}

	principal, err := NewSessionAuthenticator(repository).Authenticate(t.Context(), "token")

	require.ErrorIs(t, err, repositoryErr)
	assert.Nil(t, principal)
}
