package httputil

import (
	"net"
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
