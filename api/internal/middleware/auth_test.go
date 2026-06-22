package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequireAuth_WithUserCallsNext(t *testing.T) {
	called := false
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/protected", nil)
	ctx := context.WithValue(req.Context(), UserContextKey, &auth.User{ID: "user-1"})
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
