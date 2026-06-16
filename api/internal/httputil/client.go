package httputil

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strings"
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

// NewSSRFSafeHTTPClient returns an HTTP client that refuses to connect to
// private, loopback, link-local, or other non-public IP addresses. Use it for
// requests whose target host is (partially) attacker-controlled, such as the
// OIDC token, JWKS, discovery and userinfo endpoints of a tenant-configured SSO
// provider.
//
// Proxies are intentionally disabled: with an HTTP(S)_PROXY set, the dialer
// would only see (and validate) the proxy address, letting a private target
// slip past the IP block. Every connection — including redirect hops — is
// validated by resolving the host and rejecting non-public IPs, and the dialed
// address is pinned to the vetted IP to defeat DNS rebinding.
func NewSSRFSafeHTTPClient(timeout time.Duration) *http.Client {
	dialer := &net.Dialer{Timeout: 10 * time.Second}

	transport := &http.Transport{
		// No proxy: see the function doc for why.
		Proxy: nil,
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, port, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}

			// If the host is already an IP literal, validate it directly.
			if literal := net.ParseIP(host); literal != nil {
				if !isPublicIP(literal) {
					return nil, fmt.Errorf("blocked connection to non-public address %s", literal)
				}
				return dialer.DialContext(ctx, network, addr)
			}

			ips, err := net.DefaultResolver.LookupIPAddr(ctx, host)
			if err != nil {
				return nil, err
			}
			if len(ips) == 0 {
				return nil, fmt.Errorf("no addresses for host %q", host)
			}

			for _, ip := range ips {
				if !isPublicIP(ip.IP) {
					return nil, fmt.Errorf("blocked connection to non-public address %s", ip.IP)
				}
			}

			// Pin to a vetted IP to avoid DNS-rebinding between check and dial.
			return dialer.DialContext(ctx, network, net.JoinHostPort(ips[0].IP.String(), port))
		},
		ForceAttemptHTTP2:     true,
		MaxIdleConns:          10,
		IdleConnTimeout:       30 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	return &http.Client{
		Timeout:   timeout,
		Transport: otelhttp.NewTransport(transport),
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 5 {
				return fmt.Errorf("too many redirects")
			}
			// This client is only used for non-loopback (production) targets, so
			// never downgrade to cleartext: require HTTPS on every redirect hop.
			// The dialer still validates the resolved IP.
			if req.URL.Scheme != "https" {
				return fmt.Errorf("refusing to follow redirect to non-HTTPS scheme %q", req.URL.Scheme)
			}
			return nil
		},
	}
}

// IsLoopbackURL reports whether rawURL targets the local machine — host exactly
// "localhost" or a loopback IP literal. It deliberately does NOT match suffixes
// like "localhost.evil.com". Used to permit local-dev OIDC providers without an
// explicit config flag while keeping every non-loopback target SSRF-protected.
func IsLoopbackURL(rawURL string) bool {
	u, err := url.Parse(rawURL)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// ClientForURL returns the HTTP client appropriate for fetching rawURL. A
// loopback target (localhost / 127.0.0.1 / ::1) gets a plain client so local-dev
// IdPs work; every other target gets the SSRF-safe client. Selection is made
// per-URL so a single loopback endpoint can never downgrade requests to other
// hosts (cf. mixed-endpoint loopback-downgrade SSRF).
func ClientForURL(rawURL string, timeout time.Duration) *http.Client {
	if IsLoopbackURL(rawURL) {
		return NewHTTPClient(WithTimeout(timeout))
	}
	return NewSSRFSafeHTTPClient(timeout)
}

// ValidateHTTPSEndpoint checks that rawURL is a syntactically valid absolute
// HTTPS URL with a host. The host must not be a loopback name or a private/
// loopback IP literal. This is used to validate tenant-supplied OIDC endpoints
// at provider create/update time so a production SSO config can never persist a
// target that the login flow would reach over plain HTTP or against an internal
// address. (DNS hostnames are re-validated at dial time by
// NewSSRFSafeHTTPClient, since DNS can change between configuration and use.)
func ValidateHTTPSEndpoint(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if u.Scheme != "https" {
		return fmt.Errorf("URL must use HTTPS")
	}
	host := u.Hostname()
	if host == "" {
		return fmt.Errorf("URL must include a host")
	}
	if strings.EqualFold(host, "localhost") {
		return fmt.Errorf("URL must not target localhost")
	}
	if ip := net.ParseIP(host); ip != nil && !isPublicIP(ip) {
		return fmt.Errorf("URL must not target a private or loopback address")
	}
	return nil
}

// ValidateSSOEndpoint validates a stored OIDC endpoint URL just before it is
// used at login time. It accepts exactly what the runtime client selection
// honors: a loopback dev URL (any scheme), or otherwise a public HTTPS URL.
//
// This guards against SSO provider rows that predate ValidateHTTPSEndpoint or
// were inserted out-of-band (SQL, fixtures) with an http:// non-loopback or a
// private/loopback hostname — none of which ValidateHTTPSEndpoint would have
// allowed and which the SSRF-safe dialer alone wouldn't catch for cleartext or
// for private *hostnames* resolving at dial time.
func ValidateSSOEndpoint(rawURL string) error {
	if IsLoopbackURL(rawURL) {
		return nil
	}
	return ValidateHTTPSEndpoint(rawURL)
}

// isPublicIP reports whether the IP is a globally routable public address.
func isPublicIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() ||
		ip.IsMulticast() || ip.IsUnspecified() || ip.IsPrivate() {
		return false
	}
	if ip4 := ip.To4(); ip4 != nil {
		// 100.64.0.0/10 (carrier-grade NAT) not covered by IsPrivate on all Go versions.
		if ip4[0] == 100 && ip4[1] >= 64 && ip4[1] <= 127 {
			return false
		}
		// 169.254.0.0/16 cloud metadata is covered by IsLinkLocalUnicast.
	}
	return true
}
