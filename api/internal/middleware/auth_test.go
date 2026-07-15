package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/access"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type stubSessionAuthenticator struct {
	principal *access.Principal
	err       error
	calls     int
	token     string
}

func (s *stubSessionAuthenticator) Authenticate(_ context.Context, token string) (*access.Principal, error) {
	s.calls++
	s.token = token
	return s.principal, s.err
}

func TestOptionalAuthMiddlewareAuthenticatesOnceAndSetsPrincipal(t *testing.T) {
	principal := &access.Principal{
		User:    access.User{ID: "user-1"},
		Session: access.Session{ID: "session-1"},
	}
	authenticator := &stubSessionAuthenticator{principal: principal}

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Same(t, principal, access.PrincipalFromContext(r.Context()))
		assert.Same(t, &principal.User, access.UserFromContext(r.Context()))
		assert.Same(t, &principal.Session, access.SessionFromContext(r.Context()))
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer session-token")
	rec := httptest.NewRecorder()

	OptionalAuthMiddleware(authenticator)(next).ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, 1, authenticator.calls)
	assert.Equal(t, "session-token", authenticator.token)
}

func TestRequireAuth_WithUserCallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
	ctx := access.WithPrincipal(req.Context(), &access.Principal{User: access.User{ID: "user-1"}})
	req = req.WithContext(ctx)

	rec := httptest.NewRecorder()
	RequireAuth(next).ServeHTTP(rec, req)

	assert.True(t, called, "next handler should be called when a user is present")
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRequireAuth_WithoutUserReturns401(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
	rec := httptest.NewRecorder()
	RequireAuth(next).ServeHTTP(rec, req)

	assert.False(t, called, "next handler must not be called without a user")
	assert.Equal(t, http.StatusUnauthorized, rec.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "unauthorized", body["message"])
}
