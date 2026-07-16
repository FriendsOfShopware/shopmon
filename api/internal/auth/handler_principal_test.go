package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuthUsesValidatedRequestPrincipal(t *testing.T) {
	principal := &access.Principal{
		User:    access.User{ID: "user-1"},
		Session: access.Session{ID: "session-1"},
	}
	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req = req.WithContext(access.WithPrincipal(req.Context(), principal))
	rec := httptest.NewRecorder()

	got := (&AuthHandler{}).requireAuth(rec, req)

	require.NotNil(t, got)
	assert.Same(t, principal, got)
	assert.Equal(t, http.StatusOK, rec.Code)
}
