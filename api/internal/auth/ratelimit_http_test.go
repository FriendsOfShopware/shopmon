package auth_test

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAuthRoutesReturnJSON429WhenRateLimited(t *testing.T) {
	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.AuthRateLimitMax = 2
	})

	sessionURL := env.Server.URL + "/api/auth/session"
	for i := range 2 {
		resp, err := testutil.Get(t, sessionURL)
		require.NoError(t, err)
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "request %d should reach the handler", i+1)
	}

	resp, err := testutil.Get(t, sessionURL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Equal(t, "60", resp.Header.Get("Retry-After"))
	assert.Contains(t, resp.Header.Get("Content-Type"), "application/json")

	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Equal(t, map[string]string{"message": "too many requests, please try again later"}, body)
}

func TestAuthRateLimitDoesNotApplyOutsideAuth(t *testing.T) {
	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.AuthRateLimitMax = 1
	})

	resp, err := testutil.Get(t, env.Server.URL+"/api/auth/session")
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	resp, err = testutil.Get(t, env.Server.URL+"/api/auth/session")
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)

	resp, err = testutil.Get(t, env.Server.URL+"/api/health")
	require.NoError(t, err)
	_, _ = io.Copy(io.Discard, resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	resp, err = testutil.Get(t, env.Server.URL+"/api/account/me")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}

func TestOptionalAuthStillWorksWithRateLimiterMounted(t *testing.T) {
	env := testutil.Setup(t, func(cfg *config.Config) {
		cfg.AuthRateLimitMax = 1
	})
	token := env.SeedUser(t, "user-1", "Test User", "test@example.com", "user")

	req := env.AuthRequest(t, http.MethodGet, "/api/account/me", token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
