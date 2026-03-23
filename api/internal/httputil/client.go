package httputil

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHTTPClient returns an HTTP client with OpenTelemetry transport instrumentation.
func NewHTTPClient(opts ...func(*http.Client)) *http.Client {
	c := &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// WithTimeout sets the client timeout.
func WithTimeout(d time.Duration) func(*http.Client) {
	return func(c *http.Client) {
		c.Timeout = d
	}
}
