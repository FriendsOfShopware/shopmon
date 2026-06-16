package shopware

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// newTestServer returns an httptest server that issues OAuth tokens and serves
// API requests, recording how many times each is hit. The token handler returns
// a freshly numbered access token on every call so tests can tell a cached token
// from a re-fetched one.
type testServer struct {
	*httptest.Server
	tokenHits atomic.Int32
	apiHits   atomic.Int32

	mu          sync.Mutex
	expiresIn   int    // token lifetime advertised to the client
	apiStatus   int    // status code the /api/* handler returns
	apiBody     string // body the /api/* handler returns
	force401For int    // number of leading API calls that should 401 (for re-auth tests)
}

func newTestServer() *testServer {
	ts := &testServer{expiresIn: 3600, apiStatus: http.StatusOK, apiBody: `{"ok":true}`}
	mux := http.NewServeMux()
	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		n := ts.tokenHits.Add(1)
		ts.mu.Lock()
		expiresIn := ts.expiresIn
		ts.mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(tokenResponse{
			AccessToken: "token-" + itoa(int(n)),
			ExpiresIn:   expiresIn,
		})
	})
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		hit := ts.apiHits.Add(1)
		ts.mu.Lock()
		status, body, force401 := ts.apiStatus, ts.apiBody, ts.force401For
		ts.mu.Unlock()
		if int(hit) <= force401 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	})
	ts.Server = httptest.NewServer(mux)
	return ts
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}

func newTestClient(baseURL string) *Client {
	return NewClient(baseURL, "client-id", "client-secret", "shop-token")
}

func TestGetTokenCaches(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newTestClient(ts.URL)

	for i := 0; i < 3; i++ {
		if _, err := c.Get(context.Background(), "/_info/config"); err != nil {
			t.Fatalf("Get: %v", err)
		}
	}
	if got := ts.tokenHits.Load(); got != 1 {
		t.Fatalf("expected token to be fetched once, got %d fetches", got)
	}
	if got := ts.apiHits.Load(); got != 3 {
		t.Fatalf("expected 3 API hits, got %d", got)
	}
}

func TestGetTokenExpiryMargin(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newTestClient(ts.URL)

	// A normal long-lived token has the safety margin applied: it stays cached
	// but expires ~margin earlier than its raw lifetime.
	ts.mu.Lock()
	ts.expiresIn = 3600
	ts.mu.Unlock()
	if _, err := c.Get(context.Background(), "/a"); err != nil {
		t.Fatalf("Get: %v", err)
	}

	c.mu.Lock()
	expiresAt := c.cachedToken.expiresAt
	c.mu.Unlock()

	rawExpiry := time.Now().Add(3600 * time.Second)
	// Expiry must be earlier than the raw lifetime by roughly the margin.
	if !expiresAt.Before(rawExpiry.Add(-tokenExpiryMargin + 5*time.Second)) {
		t.Fatalf("expected expiry to be reduced by ~%s margin, got expiresAt=%v rawExpiry=%v",
			tokenExpiryMargin, expiresAt, rawExpiry)
	}
}

func TestGetTokenShortLifetimeNotPushedIntoPast(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	// A token whose advertised lifetime is below the margin must still be cached
	// with a future (not negative) expiry — the margin must not make it stale on
	// arrival, which would defeat caching entirely.
	ts.mu.Lock()
	ts.expiresIn = 10 // < tokenExpiryMargin (30s)
	ts.mu.Unlock()
	c := newTestClient(ts.URL)

	if _, err := c.Get(context.Background(), "/a"); err != nil {
		t.Fatalf("first Get: %v", err)
	}
	if tok := c.cachedAccessToken(); tok == "" {
		t.Fatal("expected sub-margin token to remain cached for its real lifetime")
	}
	if _, err := c.Get(context.Background(), "/b"); err != nil {
		t.Fatalf("second Get: %v", err)
	}
	if got := ts.tokenHits.Load(); got != 1 {
		t.Fatalf("expected sub-margin token to be reused, got %d fetches", got)
	}
}

func TestGetTokenSingleflight(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	c := newTestClient(ts.URL)

	const n = 20
	var wg sync.WaitGroup
	errs := make([]error, n)
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(i int) {
			defer wg.Done()
			_, errs[i] = c.Get(context.Background(), "/_info/config")
		}(i)
	}
	wg.Wait()

	for i, err := range errs {
		if err != nil {
			t.Fatalf("concurrent Get %d: %v", i, err)
		}
	}
	if got := ts.tokenHits.Load(); got != 1 {
		t.Fatalf("expected concurrent token fetches to collapse to 1, got %d", got)
	}
	if got := ts.apiHits.Load(); got != n {
		t.Fatalf("expected %d API hits, got %d", n, got)
	}
}

func TestRequestReauthorizesOn401(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	ts.mu.Lock()
	ts.force401For = 1 // first API call 401s, forcing one re-auth + retry
	ts.mu.Unlock()
	c := newTestClient(ts.URL)

	body, err := c.Get(context.Background(), "/_info/config")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if string(body) != `{"ok":true}` {
		t.Fatalf("unexpected body: %s", body)
	}
	// One initial token, one re-auth after the 401.
	if got := ts.tokenHits.Load(); got != 2 {
		t.Fatalf("expected 2 token fetches (initial + re-auth), got %d", got)
	}
	// First API call 401s, retry succeeds.
	if got := ts.apiHits.Load(); got != 2 {
		t.Fatalf("expected 2 API hits, got %d", got)
	}
}

func TestRequestDoesNotRetryTwiceOn401(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	ts.mu.Lock()
	ts.force401For = 100 // every API call 401s
	ts.mu.Unlock()
	c := newTestClient(ts.URL)

	_, err := c.Get(context.Background(), "/_info/config")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected ApiError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", apiErr.StatusCode)
	}
	// Exactly one retry: 2 API hits, not an infinite loop.
	if got := ts.apiHits.Load(); got != 2 {
		t.Fatalf("expected exactly 2 API hits (one retry), got %d", got)
	}
}

func TestRequestReturnsApiErrorOn4xx(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	ts.mu.Lock()
	ts.apiStatus = http.StatusUnprocessableEntity
	ts.apiBody = `{"errors":["bad"]}`
	ts.mu.Unlock()
	c := newTestClient(ts.URL)

	_, err := c.Get(context.Background(), "/x")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected ApiError, got %v", err)
	}
	if apiErr.StatusCode != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422, got %d", apiErr.StatusCode)
	}
	if apiErr.Body != `{"errors":["bad"]}` {
		t.Fatalf("unexpected error body: %s", apiErr.Body)
	}
}

func TestRequestDetectsRedirect(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()
	ts.mu.Lock()
	ts.apiStatus = http.StatusFound // 302
	ts.apiBody = ""
	ts.mu.Unlock()
	c := newTestClient(ts.URL)

	_, err := c.Get(context.Background(), "/x")
	var apiErr *ApiError
	if !errors.As(err, &apiErr) {
		t.Fatalf("expected ApiError for redirect, got %v", err)
	}
	if apiErr.StatusCode != http.StatusFound {
		t.Fatalf("expected 302, got %d", apiErr.StatusCode)
	}
}

func TestAuthenticatePropagatesError(t *testing.T) {
	ts := newTestServer()
	c := newTestClient(ts.URL)
	ts.Close() // server down: token request must fail

	if err := c.Authenticate(context.Background()); err == nil {
		t.Fatal("expected error when token endpoint is unreachable")
	}
}

func TestRequestHonorsContextCancellation(t *testing.T) {
	block := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		<-block // hang until the test unblocks, forcing the client ctx to cancel
	}))
	defer srv.Close()
	defer close(block)

	c := newTestClient(srv.URL)
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := c.Get(ctx, "/_info/config")
	if err == nil {
		t.Fatal("expected error from cancelled context")
	}
}
