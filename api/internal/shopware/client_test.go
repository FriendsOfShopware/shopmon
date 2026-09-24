package shopware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Token caching, singleflight and the 401 retry are covered by the library's
// own tests; these tests pin the Shopmon-specific wiring.

type recordedRequest struct {
	path      string
	userAgent string
	shopToken string
	auth      string
}

func newRecordingServer(t *testing.T, apiStatus int, apiBody string) (*httptest.Server, *[]recordedRequest, *atomic.Int32) {
	t.Helper()
	var requests []recordedRequest
	var tokenHits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests = append(requests, recordedRequest{
			path:      r.URL.Path,
			userAgent: r.Header.Get("User-Agent"),
			shopToken: r.Header.Get(ShopTokenHeader),
			auth:      r.Header.Get("Authorization"),
		})
		if r.URL.Path == "/api/oauth/token" {
			tokenHits.Add(1)
			var body map[string]string
			_ = json.NewDecoder(r.Body).Decode(&body)
			if body["client_id"] != "client-id" || body["client_secret"] != "client-secret" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"access_token":"tok","expires_in":3600}`))
			return
		}
		w.WriteHeader(apiStatus)
		_, _ = w.Write([]byte(apiBody))
	}))
	t.Cleanup(srv.Close)
	return srv, &requests, &tokenHits
}

func TestClientSendsShopTokenAndUserAgent(t *testing.T) {
	srv, requests, tokenHits := newRecordingServer(t, http.StatusOK, `{"ok":true}`)
	c := NewClient(srv.URL+"/", "client-id", "client-secret", "shop-token")

	for range 2 {
		resp, err := c.Get(context.Background(), "/_info/config")
		require.NoError(t, err)
		assert.JSONEq(t, `{"ok":true}`, string(resp.Body))
	}
	assert.Equal(t, int32(1), tokenHits.Load(), "token is cached across requests")

	require.Len(t, *requests, 3)
	for _, r := range *requests {
		assert.Equal(t, "shop-token", r.shopToken, "shop token header on %s", r.path)
		assert.True(t, strings.HasPrefix(r.userAgent, "Shopmon"), "User-Agent on %s = %q", r.path, r.userAgent)
	}
	assert.Equal(t, "/api/_info/config", (*requests)[1].path)
	assert.Equal(t, "Bearer tok", (*requests)[1].auth)
}

func TestClientReturnsAPIError(t *testing.T) {
	srv, _, _ := newRecordingServer(t, http.StatusNotFound, `{"errors":[{"detail":"no route"}]}`)
	c := NewClient(srv.URL, "client-id", "client-secret", "shop-token")

	_, err := c.Get(context.Background(), "/x")
	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok, "expected *APIError, got %v", err)
	assert.Equal(t, http.StatusNotFound, apiErr.StatusCode)
	assert.Equal(t, "no route", apiErr.Detail)
}

func TestAuthenticateRejectsBadCredentials(t *testing.T) {
	srv, _, _ := newRecordingServer(t, http.StatusOK, `{}`)
	c := NewClient(srv.URL, "client-id", "wrong", "shop-token")

	err := c.Authenticate(context.Background())
	apiErr, ok := errors.AsType[*APIError](err)
	require.True(t, ok, "expected *APIError, got %v", err)
	assert.Equal(t, http.StatusUnauthorized, apiErr.StatusCode)
}
