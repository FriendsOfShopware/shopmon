package uptime

import (
	"context"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newLocalListener reserves a loopback TCP port for the refused-connection test.
func newLocalListener() (net.Listener, error) {
	var lc net.ListenConfig
	return lc.Listen(context.Background(), "tcp", "127.0.0.1:0")
}

func TestProberSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, probeUserAgent, r.Header.Get("User-Agent"))
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("welcome to the shop"))
	}))
	defer server.Close()

	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL})

	assert.True(t, res.OK)
	assert.Equal(t, http.StatusOK, res.StatusCode)
	assert.Empty(t, res.Err)
	assert.GreaterOrEqual(t, res.Latency, time.Duration(0))
}

func TestProberAccepts2xxAnd3xxByDefault(t *testing.T) {
	for _, code := range []int{200, 204, 301, 399} {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			// Redirect codes are followed by the client, so serve a terminal body.
			w.WriteHeader(code)
		}))

		res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL})
		assert.True(t, res.OK, "status %d should be accepted", code)
		server.Close()
	}
}

func TestProberRejects4xxAnd5xx(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer server.Close()

	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL})

	assert.False(t, res.OK)
	assert.Equal(t, http.StatusServiceUnavailable, res.StatusCode)
	assert.Equal(t, "status 503", res.Err)
}

func TestProberExpectedStatusExact(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// A shop that intentionally serves 404 on this URL can pin the expectation.
	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL, ExpectedStatus: 404})
	assert.True(t, res.OK)

	res = newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL, ExpectedStatus: 200})
	assert.False(t, res.OK)
}

func TestProberContentMatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("<html>Powered by Shopware</html>"))
	}))
	defer server.Close()

	match := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL, ContentMatch: "Shopware"})
	assert.True(t, match.OK)

	miss := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL, ContentMatch: "Magento"})
	assert.False(t, miss.OK)
	assert.Equal(t, "content mismatch", miss.Err)
}

func TestProberFollowsRedirects(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/final", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/final", http.StatusFound)
	})
	server := httptest.NewServer(mux)
	defer server.Close()

	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL})
	assert.True(t, res.OK)
	assert.Equal(t, http.StatusOK, res.StatusCode)
}

func TestProberTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(300 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: server.URL, Timeout: 50 * time.Millisecond})

	assert.False(t, res.OK)
	assert.Equal(t, "timeout", res.Err)
}

func TestProberConnectionRefused(t *testing.T) {
	// Reserve a port and close the listener so nothing accepts on it.
	listener, err := newLocalListener()
	require.NoError(t, err)
	addr := listener.Addr().String()
	require.NoError(t, listener.Close())

	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: "http://" + addr})

	assert.False(t, res.OK)
	assert.NotEmpty(t, res.Err)
}

func TestProberDNSFailure(t *testing.T) {
	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{
		URL:     "https://this-host-does-not-exist.shopmon.invalid",
		Timeout: 5 * time.Second,
	})

	assert.False(t, res.OK)
	assert.True(t, strings.HasPrefix(res.Err, "dns"), "expected dns error, got %q", res.Err)
}

func TestProberInvalidURL(t *testing.T) {
	res := newHTTPProber(true).Probe(context.Background(), ProbeConfig{URL: "://broken"})
	assert.False(t, res.OK)
	assert.Contains(t, res.Err, "invalid url")
}

func TestStatusMatches(t *testing.T) {
	assert.True(t, statusMatches(200, 0))
	assert.True(t, statusMatches(301, 0))
	assert.False(t, statusMatches(404, 0))
	assert.False(t, statusMatches(500, 0))
	assert.True(t, statusMatches(404, 404))
	assert.False(t, statusMatches(200, 404))
}

func TestProberBlocksPrivateTargets(t *testing.T) {
	// httptest binds to 127.0.0.1, which the production prober must refuse.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	res := NewHTTPProber().Probe(context.Background(), ProbeConfig{URL: server.URL, Timeout: 5 * time.Second})

	assert.False(t, res.OK)
	assert.Zero(t, res.StatusCode)
	assert.Equal(t, "target address is not allowed", res.Err)
}

func TestBlockPrivateAddresses(t *testing.T) {
	blocked := []string{
		"127.0.0.1:80",       // loopback
		"10.0.0.5:80",        // private
		"192.168.1.1:80",     // private
		"169.254.169.254:80", // cloud metadata
		"0.0.0.0:80",         // unspecified
		"[::1]:80",           // IPv6 loopback
		"[fd00::1]:80",       // IPv6 unique local
		"[fe80::1]:80",       // IPv6 link local
		"224.0.0.1:80",       // multicast
		"255.255.255.255:80", // broadcast
	}
	for _, addr := range blocked {
		assert.Error(t, blockPrivateAddresses("tcp", addr, nil), "%s must be blocked", addr)
	}

	allowed := []string{"93.184.216.34:443", "[2606:2800:220:1:248:1893:25c8:1946]:443"}
	for _, addr := range allowed {
		assert.NoError(t, blockPrivateAddresses("tcp", addr, nil), "%s must be allowed", addr)
	}
}
