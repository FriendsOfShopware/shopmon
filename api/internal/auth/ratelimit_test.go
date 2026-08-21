package auth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRateLimitMiddlewareAllowsThenRejects(t *testing.T) {
	t.Parallel()

	limiter := RateLimitMiddleware(NewRateLimiter(t.Context(), time.Minute, 2))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := limiter(next)

	for i := range 2 {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/auth/session", nil)
		req.RemoteAddr = "203.0.113.10:1"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
		assert.Equal(t, http.StatusNoContent, rec.Code, "request %d should be allowed", i+1)
	}

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/auth/session", nil)
	req.RemoteAddr = "203.0.113.10:1"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "60", rec.Header().Get("Retry-After"))
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, map[string]string{"message": "too many requests, please try again later"}, body)
}

func TestRateLimitMiddlewareIsPerIP(t *testing.T) {
	t.Parallel()

	limiter := RateLimitMiddleware(NewRateLimiter(t.Context(), time.Minute, 1))
	next := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	handler := limiter(next)

	first := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	first.RemoteAddr = "203.0.113.10:1"
	firstRec := httptest.NewRecorder()
	handler.ServeHTTP(firstRec, first)
	assert.Equal(t, http.StatusNoContent, firstRec.Code)

	blocked := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	blocked.RemoteAddr = "203.0.113.10:1"
	blockedRec := httptest.NewRecorder()
	handler.ServeHTTP(blockedRec, blocked)
	assert.Equal(t, http.StatusTooManyRequests, blockedRec.Code)

	other := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/", nil)
	other.RemoteAddr = "203.0.113.11:1"
	otherRec := httptest.NewRecorder()
	handler.ServeHTTP(otherRec, other)
	assert.Equal(t, http.StatusNoContent, otherRec.Code)
}
