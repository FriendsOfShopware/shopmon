package uptime

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"syscall"
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

// errAddressBlocked marks dials to non-public targets refused by the prober.
var errAddressBlocked = errors.New("target address is not allowed")

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

// NewHTTPProber returns the default storefront prober with the private-address
// guard enabled. The client is shared across probes; per-probe deadlines come
// from the request context.
func NewHTTPProber() Prober {
	return newHTTPProber(false)
}

// newHTTPProber builds the prober; tests may allow private targets so they can
// probe loopback httptest servers.
func newHTTPProber(allowPrivateTargets bool) Prober {
	dialer := &net.Dialer{
		Timeout:   10 * time.Second,
		KeepAlive: 30 * time.Second,
	}
	if !allowPrivateTargets {
		dialer.Control = blockPrivateAddresses
	}
	return &httpProber{
		client: &http.Client{
			// Redirects are followed (Go default: up to 10 hops) since
			// storefronts frequently redirect (www, locale, trailing slash).
			// The dial-time address check also guards redirect targets.
			Transport: &http.Transport{
				Proxy:                 nil, // probes must be direct, never via a proxy
				DialContext:           dialer.DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		},
	}
}

// blockPrivateAddresses is a net.Dialer Control hook that refuses connections
// to anything but public addresses. Uptime targets are public storefronts, so
// loopback, private, link-local (including the cloud metadata service at
// 169.254.169.254), multicast and unspecified targets are blocked. Checking at
// dial time rather than at settings-validation time means redirects and DNS
// answers that resolve differently than the URL looked like are covered too.
func blockPrivateAddresses(network, address string, _ syscall.RawConn) error {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		return fmt.Errorf("uptime probe: invalid dial address %q: %w", address, err)
	}
	ip := net.ParseIP(host)
	if ip == nil {
		return fmt.Errorf("uptime probe: non-IP dial address %q: %w", host, errAddressBlocked)
	}
	if !ip.IsGlobalUnicast() || ip.IsPrivate() || ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return fmt.Errorf("uptime probe: dial to %s %s: %w", network, host, errAddressBlocked)
	}
	return nil
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

	// Drain a small body remainder so the connection can be reused for the
	// next probe of this host; skip large ones.
	_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 64<<10))

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
	case errors.Is(err, errAddressBlocked):
		return "target address is not allowed"
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return "dns lookup failed"
	}

	var opErr *net.OpError
	if errors.As(err, &opErr) && opErr.Op == "dial" {
		if opErr.Timeout() {
			return "connect timeout"
		}
		return "connection failed"
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
