package shopwareaccount

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// DefaultBaseURL is the public Shopware account/store API base URL.
const DefaultBaseURL = "https://api.shopware.com"

// Client talks to the Shopware account/store API.
type Client struct {
	baseURL    string
	httpClient *http.Client

	// Optional retry tuning for 429 responses. Zero values use defaults; tests
	// set these to keep backoff fast and deterministic.
	maxAttempts   int
	baseBackoff   time.Duration
	maxBackoff    time.Duration
	maxRetryAfter time.Duration
	sleepFn       func(context.Context, time.Duration) error
}

// NewClient creates a Client for the given base URL. If httpClient is nil a
// default instrumented client with a 30s timeout is used.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if httpClient == nil {
		httpClient = httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second))
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}

// ConfigureRetry overrides 429 retry behaviour. Intended for tests that need
// deterministic, fast backoff; production callers should use NewClient defaults.
func (c *Client) ConfigureRetry(maxAttempts int, sleep func(context.Context, time.Duration) error) {
	c.maxAttempts = maxAttempts
	c.sleepFn = sleep
}
