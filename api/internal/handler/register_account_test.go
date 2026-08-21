package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestRegisterAccountAndNotifications(t *testing.T) {
	r := chi.NewRouter()
	config := huma.DefaultConfig("test", "1.0.0")
	config.CreateHooks = nil
	config.Transformers = nil
	api := humachi.New(r, config)
	h := New(Dependencies{})
	registerAccount(api, h)
	registerNotifications(api, h)

	for _, tc := range []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/account/me"},
		{http.MethodPatch, "/account/me"},
		{http.MethodGet, "/account/extensions"},
		{http.MethodGet, "/account/extensions/SwagPayPal"},
		{http.MethodGet, "/account/organizations"},
		{http.MethodGet, "/account/environments"},
		{http.MethodGet, "/account/shops"},
		{http.MethodGet, "/account/changelogs"},
		{http.MethodGet, "/account/subscribed-environments"},
		{http.MethodGet, "/notifications"},
		{http.MethodDelete, "/notifications"},
		{http.MethodGet, "/notifications/event-types"},
		{http.MethodPost, "/notifications/mark-read"},
		{http.MethodDelete, "/notifications/1"},
		{http.MethodGet, "/account/notification-preferences"},
		{http.MethodPut, "/account/notification-preferences"},
		{http.MethodDelete, "/account/notification-preferences?scopeType=global&channel=email"},
	} {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequestWithContext(t.Context(), tc.method, tc.path, nil)
			if tc.method == http.MethodPatch || tc.method == http.MethodPut {
				req.Header.Set("Content-Type", "application/json")
				req.Body = http.NoBody
			}
			rec := httptest.NewRecorder()
			r.ServeHTTP(rec, req)
			assert.NotEqual(t, http.StatusNotFound, rec.Code)
			assert.NotEqual(t, http.StatusMethodNotAllowed, rec.Code)
		})
	}
}
