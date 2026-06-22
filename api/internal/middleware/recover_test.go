package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func panicHandler() http.Handler {
	return http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		panic("boom")
	})
}

func TestRecovererAPIPathReturnsJSON(t *testing.T) {
	h := middleware.Recoverer(panicHandler())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/environments/1", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
	assert.Equal(t, "internal server error", body["message"])
	// The panic value must not leak into the response.
	assert.NotContains(t, rec.Body.String(), "boom")
}

func TestRecovererNonAPIPathBareStatus(t *testing.T) {
	h := middleware.Recoverer(panicHandler())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/dashboard", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusInternalServerError, rec.Code)
	assert.NotContains(t, rec.Body.String(), "boom")
}

func TestRecovererNoPanicPassesThrough(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("fine"))
	})
	h := middleware.Recoverer(ok)

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/api/health", nil)
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	assert.Equal(t, "fine", rec.Body.String())
}
