package shopwareaccount

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPluginsByNameRetries429WithRetryAfter(t *testing.T) {
	var hits atomic.Int64
	var slept []time.Duration

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n < 3 {
			w.Header().Set("Retry-After", "2")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode([]map[string]any{
			{"id": 1, "name": "FroshTools", "label": "Tools", "version": "1.0.0"},
		}))
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL, srv.Client())
	client.maxAttempts = 5
	client.sleepFn = func(_ context.Context, d time.Duration) error {
		slept = append(slept, d)
		return nil
	}

	plugins, err := client.PluginsByName(context.Background(), "en_GB", "6.6.0.0", []string{"FroshTools"})
	require.NoError(t, err)
	require.Len(t, plugins, 1)
	assert.Equal(t, "FroshTools", plugins[0].Name)
	assert.Equal(t, int64(3), hits.Load(), "should retry until success")
	require.Len(t, slept, 2)
	assert.Equal(t, 2*time.Second, slept[0], "honor Retry-After")
	assert.Equal(t, 2*time.Second, slept[1], "honor Retry-After on second wait")
}

func TestPluginsByNameRetries429WithExponentialBackoff(t *testing.T) {
	var hits atomic.Int64
	var slept []time.Duration

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n < 3 {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte("[]"))
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL, srv.Client())
	client.maxAttempts = 5
	client.baseBackoff = 100 * time.Millisecond
	client.maxBackoff = time.Second
	client.sleepFn = func(_ context.Context, d time.Duration) error {
		slept = append(slept, d)
		return nil
	}

	_, err := client.PluginsByName(context.Background(), "en_GB", "6.6.0.0", []string{"FroshTools"})
	require.NoError(t, err)
	assert.Equal(t, int64(3), hits.Load())
	require.Len(t, slept, 2)
	// Without Retry-After: base*2^(attempt-1) + jitter in [0, base].
	assert.GreaterOrEqual(t, slept[0], 100*time.Millisecond)
	assert.LessOrEqual(t, slept[0], 200*time.Millisecond)
	assert.GreaterOrEqual(t, slept[1], 200*time.Millisecond)
	assert.LessOrEqual(t, slept[1], 300*time.Millisecond)
}

func TestPluginsByNameExhausts429Retries(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Retry-After", "1")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL, srv.Client())
	client.maxAttempts = 3
	client.sleepFn = func(context.Context, time.Duration) error { return nil }

	_, err := client.PluginsByName(context.Background(), "en_GB", "6.6.0.0", []string{"FroshTools"})
	require.Error(t, err)
	assert.True(t, IsRateLimited(err), "error should be rate-limited: %v", err)
	assert.Equal(t, int64(3), hits.Load())
}

func TestPluginsByNameDoesNotRetryNon429(t *testing.T) {
	var hits atomic.Int64
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.WriteHeader(http.StatusBadRequest)
	}))
	t.Cleanup(srv.Close)

	client := NewClient(srv.URL, srv.Client())
	client.maxAttempts = 5
	client.sleepFn = func(context.Context, time.Duration) error {
		t.Fatal("should not sleep for non-429")
		return nil
	}

	_, err := client.PluginsByName(context.Background(), "en_GB", "6.6.0.0", []string{"FroshTools"})
	require.Error(t, err)
	var apiErr *APIError
	require.ErrorAs(t, err, &apiErr)
	assert.Equal(t, http.StatusBadRequest, apiErr.StatusCode)
	assert.False(t, IsRateLimited(err))
	assert.Equal(t, int64(1), hits.Load())
}

func TestParseRetryAfter(t *testing.T) {
	assert.Equal(t, time.Duration(0), parseRetryAfter(""))
	assert.Equal(t, 5*time.Second, parseRetryAfter("5"))
	assert.Equal(t, time.Duration(0), parseRetryAfter("-1"))
	assert.Equal(t, time.Duration(0), parseRetryAfter("not-a-date"))

	future := time.Now().UTC().Add(3 * time.Second).Format(http.TimeFormat)
	d := parseRetryAfter(future)
	assert.Greater(t, d, time.Duration(0))
	assert.LessOrEqual(t, d, 3*time.Second+time.Second)
}
