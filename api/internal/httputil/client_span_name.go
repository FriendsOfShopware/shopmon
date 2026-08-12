package httputil

import (
	"net"
	"net/http"
	"strings"
)

// ClientSpanName formats an outbound HTTP client span as "METHOD host/path".
//
// otelhttp's default transport formatter uses only the method ("HTTP GET"),
// which Datadog collapses to a bare GET/POST resource. Including host + path
// (without query string) keeps cardinality low while making worker traces
// actionable — e.g. "GET api.shopware.com/pluginStore/pluginsByName".
//
// The operation argument is ignored; otelhttp's client transport always passes "".
func ClientSpanName(_ string, r *http.Request) string {
	if r == nil {
		return http.MethodGet
	}

	method := r.Method
	if method == "" {
		method = http.MethodGet
	}

	host := requestHost(r)
	path := requestPath(r)

	if host == "" {
		return method + " " + path
	}
	if strings.HasPrefix(path, "/") {
		return method + " " + host + path
	}
	return method + " " + host + "/" + path
}

func requestHost(r *http.Request) string {
	if r.URL != nil {
		if host := r.URL.Hostname(); host != "" {
			return host
		}
		if host := hostnameOnly(r.URL.Host); host != "" {
			return host
		}
	}
	return hostnameOnly(r.Host)
}

func requestPath(r *http.Request) string {
	if r.URL == nil {
		return "/"
	}
	// EscapedPath omits the query string; fall back to Path, then "/".
	path := r.URL.EscapedPath()
	if path == "" {
		path = r.URL.Path
	}
	if path == "" {
		return "/"
	}
	return path
}

func hostnameOnly(hostport string) string {
	if hostport == "" {
		return ""
	}
	// Strip brackets from IPv6 literals without a port: "[::1]".
	if strings.HasPrefix(hostport, "[") && strings.HasSuffix(hostport, "]") {
		return hostport[1 : len(hostport)-1]
	}
	host, _, err := net.SplitHostPort(hostport)
	if err != nil {
		return hostport
	}
	return host
}
