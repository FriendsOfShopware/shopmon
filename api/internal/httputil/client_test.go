package httputil

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsPublicIP(t *testing.T) {
	cases := []struct {
		ip     string
		public bool
	}{
		{"8.8.8.8", true},
		{"1.1.1.1", true},
		{"93.184.216.34", true},
		{"127.0.0.1", false},           // loopback
		{"10.0.0.1", false},            // RFC1918
		{"172.16.0.1", false},          // RFC1918
		{"192.168.1.1", false},         // RFC1918
		{"169.254.169.254", false},     // link-local (cloud metadata)
		{"100.64.0.1", false},          // carrier-grade NAT
		{"0.0.0.0", false},             // unspecified
		{"::1", false},                 // IPv6 loopback
		{"fe80::1", false},             // IPv6 link-local
		{"fc00::1", false},             // IPv6 unique local
		{"2606:4700:4700::1111", true}, // public IPv6
	}

	for _, c := range cases {
		ip := net.ParseIP(c.ip)
		require.NotNil(t, ip, "failed to parse IP %q", c.ip)
		assert.Equal(t, c.public, isPublicIP(ip), "isPublicIP(%s)", c.ip)
	}
}

func TestValidateHTTPSEndpoint(t *testing.T) {
	cases := []struct {
		url string
		ok  bool
	}{
		{"https://idp.example.com/token", true},
		{"https://idp.example.com:8443/jwks", true},
		{"http://idp.example.com/token", false},  // not HTTPS
		{"https://localhost/token", false},       // loopback hostname
		{"https://localhost:8443/token", false},  // loopback hostname w/ port
		{"https://LOCALHOST/token", false},       // loopback hostname, uppercase
		{"https://127.0.0.1/token", false},       // loopback IP literal
		{"https://10.0.0.1/token", false},        // private IP literal
		{"https://169.254.169.254/token", false}, // metadata IP literal
		{"https://[::1]/token", false},           // IPv6 loopback literal
		{"ftp://idp.example.com", false},         // wrong scheme
		{"https://", false},                      // no host
		{"://bad", false},                        // unparseable
	}

	for _, c := range cases {
		err := ValidateHTTPSEndpoint(c.url)
		if c.ok {
			assert.NoError(t, err, "ValidateHTTPSEndpoint(%q)", c.url)
		} else {
			assert.Error(t, err, "ValidateHTTPSEndpoint(%q)", c.url)
		}
	}
}

// TestClientForURL_PerURLSelection documents that the client is chosen per URL:
// a loopback URL gets the plain client while a public URL gets the SSRF-safe
// one. This guards against the mixed-endpoint loopback-downgrade SSRF where one
// loopback endpoint would otherwise weaken requests to other hosts.
func TestClientForURL_PerURLSelection(t *testing.T) {
	cases := []struct {
		url      string
		loopback bool
	}{
		{"http://localhost:9000/token", true},
		{"http://127.0.0.1:9000/jwks", true},
		{"http://[::1]:9000/userinfo", true},
		{"https://idp.example.com/token", false},
		{"https://10.0.0.1/jwks", false}, // private, but not loopback: SSRF-safe client guards it at dial time
	}

	for _, c := range cases {
		assert.Equal(t, c.loopback, IsLoopbackURL(c.url), "IsLoopbackURL(%q)", c.url)
		// ClientForURL must return a non-nil client for every URL.
		assert.NotNil(t, ClientForURL(c.url, 0), "ClientForURL(%q) returned nil", c.url)
	}
}

func TestBuildUserAgent(t *testing.T) {
	assert.Equal(t, "Shopmon", buildUserAgent(""))
	assert.Equal(t, "Shopmon/1.2.3", buildUserAgent("1.2.3"))
	assert.Equal(t, "Shopmon/abcdef012345", buildUserAgent("abcdef012345"))
}

func TestUserAgentStringFormat(t *testing.T) {
	ua := UserAgentString()
	require.True(t, strings.HasPrefix(ua, "Shopmon"), "User-Agent must start with Shopmon, got %q", ua)
	// Either bare product name or Shopmon/<version> — never the Go default.
	assert.NotContains(t, ua, "Go-http-client")
	if strings.Contains(ua, "/") {
		parts := strings.SplitN(ua, "/", 2)
		require.Len(t, parts, 2)
		assert.Equal(t, "Shopmon", parts[0])
		assert.NotEmpty(t, parts[1])
	} else {
		assert.Equal(t, "Shopmon", ua)
	}
}

func TestNewHTTPClientSetsUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	resp, err := NewHTTPClient().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.NotEmpty(t, gotUA, "expected User-Agent header on outbound request")
	assert.True(t, strings.HasPrefix(gotUA, "Shopmon"), "got User-Agent %q", gotUA)
	assert.NotContains(t, gotUA, "Go-http-client")
	assert.Equal(t, UserAgentString(), gotUA)
}

func TestNewHTTPClientPreservesExplicitUserAgent(t *testing.T) {
	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		w.WriteHeader(http.StatusNoContent)
	}))
	defer srv.Close()

	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, srv.URL, nil)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "CustomAgent/9.9")

	resp, err := NewHTTPClient().Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, "CustomAgent/9.9", gotUA)
}

func TestUserAgentTransportInjectsHeader(t *testing.T) {
	var gotUA string
	base := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		gotUA = req.Header.Get("User-Agent")
		return &http.Response{
			StatusCode: http.StatusNoContent,
			Body:       http.NoBody,
			Header:     make(http.Header),
			Request:    req,
		}, nil
	})

	rt := &userAgentTransport{base: base, ua: "Shopmon/test"}
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, "https://example.com/", nil)
	require.NoError(t, err)
	resp, err := rt.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, "Shopmon/test", gotUA)
	// Original request must not be mutated.
	assert.Empty(t, req.Header.Get("User-Agent"))
}

// roundTripFunc is a test helper implementing http.RoundTripper.
type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
