package apirouter

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsAuthPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		path string
		want bool
	}{
		{path: "/auth", want: true},
		{path: "/auth/sign-in/email", want: true},
		{path: "/api/auth", want: true},
		{path: "/api/auth/session", want: true},
		{path: "/api/auth/sign-in/email", want: true},
		{path: "/api/account/me", want: false},
		{path: "/api/health", want: false},
		{path: "/api/authorize", want: false},
		{path: "/health", want: false},
		{path: "/", want: false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, isAuthPath(tt.path))
		})
	}
}

func TestAuthPathMiddlewareSkipsNonAuthRoutes(t *testing.T) {
	t.Parallel()

	limited := false
	limit := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limited = true
			next.ServeHTTP(w, r)
		})
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := authPathMiddleware(limit)(inner)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.False(t, limited)
}

func TestAuthPathMiddlewareAppliesToAuthRoutes(t *testing.T) {
	t.Parallel()

	limited := false
	limit := func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			limited = true
			next.ServeHTTP(w, r)
		})
	}

	inner := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := authPathMiddleware(limit)(inner)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/auth/session", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusNoContent, rec.Code)
	assert.True(t, limited)
}
