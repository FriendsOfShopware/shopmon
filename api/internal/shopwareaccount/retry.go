package shopwareaccount

import (
	"context"
	"errors"
	"fmt"
	"io"
	"math/rand/v2"
	"net/http"
	"strconv"
	"time"
)

const (
	defaultMaxAttempts   = 5
	defaultBaseBackoff   = 500 * time.Millisecond
	defaultMaxBackoff    = 30 * time.Second
	defaultMaxRetryAfter = 60 * time.Second
)

// APIError is returned for non-2xx store API responses after any retries are
// exhausted (or immediately for non-retryable statuses).
type APIError struct {
	StatusCode int
}

func (e *APIError) Error() string {
	return fmt.Sprintf("store api returned status %d", e.StatusCode)
}

// IsRateLimited reports whether err is (or wraps) a store API 429 response.
func IsRateLimited(err error) bool {
	var apiErr *APIError
	return errors.As(err, &apiErr) && apiErr.StatusCode == http.StatusTooManyRequests
}

func (c *Client) retryMaxAttempts() int {
	if c.maxAttempts > 0 {
		return c.maxAttempts
	}
	return defaultMaxAttempts
}

func (c *Client) retryBaseBackoff() time.Duration {
	if c.baseBackoff > 0 {
		return c.baseBackoff
	}
	return defaultBaseBackoff
}

func (c *Client) retryMaxBackoff() time.Duration {
	if c.maxBackoff > 0 {
		return c.maxBackoff
	}
	return defaultMaxBackoff
}

func (c *Client) retryMaxRetryAfter() time.Duration {
	if c.maxRetryAfter > 0 {
		return c.maxRetryAfter
	}
	return defaultMaxRetryAfter
}

func (c *Client) sleep(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return nil
	}
	if c.sleepFn != nil {
		return c.sleepFn(ctx, d)
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

// do performs req, retrying on HTTP 429 with Retry-After when present and
// otherwise exponential backoff with jitter. Non-429 responses are returned
// immediately (caller interprets the status). Context cancellation aborts waits.
func (c *Client) do(req *http.Request) (*http.Response, error) {
	attempts := c.retryMaxAttempts()
	var lastStatus int

	for attempt := 1; attempt <= attempts; attempt++ {
		resp, err := c.httpClient.Do(req)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		lastStatus = resp.StatusCode
		retryAfter := parseRetryAfter(resp.Header.Get("Retry-After"))
		// Drain and close so the connection can be reused on the next attempt.
		_, _ = io.Copy(io.Discard, resp.Body)
		_ = resp.Body.Close()

		if attempt == attempts {
			break
		}

		delay := c.backoffDelay(attempt, retryAfter)
		if err := c.sleep(req.Context(), delay); err != nil {
			return nil, fmt.Errorf("wait before retry: %w", err)
		}
	}

	return nil, &APIError{StatusCode: lastStatus}
}

func (c *Client) backoffDelay(attempt int, retryAfter time.Duration) time.Duration {
	if retryAfter > 0 {
		if capAt := c.retryMaxRetryAfter(); retryAfter > capAt {
			return capAt
		}
		return retryAfter
	}

	// attempt is 1-based for the failed request. Cap exponential growth and add
	// full-jitter in [0, base].
	base := c.retryBaseBackoff()
	mult := 1 << min(attempt-1, 20)
	exp := min(base*time.Duration(mult), c.retryMaxBackoff())
	jitter := time.Duration(rand.Int64N(int64(base) + 1))
	return exp + jitter
}

// parseRetryAfter parses a Retry-After header as delta-seconds or HTTP-date.
// Returns 0 when the header is missing or unparseable.
func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 0
	}
	if seconds, err := strconv.Atoi(h); err == nil {
		if seconds < 0 {
			return 0
		}
		return time.Duration(seconds) * time.Second
	}
	if t, err := http.ParseTime(h); err == nil {
		d := time.Until(t)
		if d < 0 {
			return 0
		}
		return d
	}
	return 0
}
