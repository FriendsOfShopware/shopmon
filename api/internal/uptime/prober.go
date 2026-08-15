package uptime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

// maxContentRead bounds how much of a response body is read for content
// matching, so a huge storefront page cannot occupy the probe worker.
const maxContentRead = 2 << 20 // 2 MiB

// maxErrorLength bounds persisted/transported error strings.
const maxErrorLength = 200

// probeUserAgent identifies probes so shop operators can tell Shopmon traffic
// from real customers in their logs.
const probeUserAgent = "Shopmon-Uptime/1.0"

// ProbeConfig describes a single probe target.
type ProbeConfig struct {
	URL string
	// ExpectedStatus 0 accepts any 2xx/3xx response; otherwise the exact
	// status code is required.
	ExpectedStatus int
	// ContentMatch is an optional substring that must appear in the body.
	ContentMatch string
	// Timeout bounds the whole probe (DNS + connect + TLS + body).
	Timeout time.Duration
}

// ProbeResult is one probe outcome. Err is empty on success.
type ProbeResult struct {
	OK         bool
	StatusCode int // 0 when no HTTP response was received
	Latency    time.Duration
	Err        string
}

// Prober executes a single uptime probe.
type Prober interface {
	Probe(ctx context.Context, cfg ProbeConfig) ProbeResult
}

type httpProber struct {
	client *http.Client
}

// NewHTTPProber returns the default storefront prober. The client is shared
// across probes; per-probe deadlines come from the request context.
func NewHTTPProber() Prober {
	return &httpProber{
		client: &http.Client{
			// Redirects are followed (Go default: up to 10 hops) since
			// storefronts frequently redirect (www, locale, trailing slash).
			Transport: &http.Transport{
				Proxy:                 nil, // probes must be direct, never via a proxy
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

func (p *httpProber) Probe(ctx context.Context, cfg ProbeConfig) ProbeResult {
	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, cfg.URL, nil)
	if err != nil {
		return ProbeResult{Err: truncateError(fmt.Sprintf("invalid url: %v", err))}
	}
	req.Header.Set("User-Agent", probeUserAgent)
	req.Header.Set("Accept", "text/html,application/xhtml+xml")

	start := time.Now()
	resp, err := p.client.Do(req)
	latency := time.Since(start)
	if err != nil {
		return ProbeResult{Latency: latency, Err: classifyError(err)}
	}
	defer func() { _ = resp.Body.Close() }()

	ok := statusMatches(resp.StatusCode, cfg.ExpectedStatus)
	failReason := ""
	if !ok {
		failReason = fmt.Sprintf("status %d", resp.StatusCode)
	}

	if ok && cfg.ContentMatch != "" {
		body, err := io.ReadAll(io.LimitReader(resp.Body, maxContentRead))
		if err != nil {
			ok = false
			failReason = "body read error"
		} else if !strings.Contains(string(body), cfg.ContentMatch) {
			ok = false
			failReason = "content mismatch"
		}
	}

	// Drain a small body so the connection can be reused; skip large ones.
	if ok && cfg.ContentMatch == "" {
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))
	}

	return ProbeResult{
		OK:         ok,
		StatusCode: resp.StatusCode,
		Latency:    latency,
		Err:        failReason,
	}
}

func statusMatches(code, expected int) bool {
	if expected > 0 {
		return code == expected
	}
	return code >= 200 && code < 400
}

// classifyError turns transport errors into short, user-presentable reasons.
func classifyError(err error) string {
	switch {
	case errors.Is(err, context.DeadlineExceeded):
		return "timeout"
	case errors.Is(err, context.Canceled):
		return "canceled"
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns lookup failed"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) {
		if opErr.Op == "dial" {
			return "connection refused"
		}
	}

	msg := err.Error()
	if strings.Contains(msg, "tls:") || strings.Contains(msg, "certificate") {
		return "tls error"
	}
	return truncateError(msg)
}

func truncateError(msg string) string {
	if len(msg) > maxErrorLength {
		return msg[:maxErrorLength]
	}
	return msg
}
