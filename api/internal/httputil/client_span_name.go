package httputil

import (
	"net"
	"net/http"
	"strings"
	"unicode"
)

// sharedSpanHosts are outbound peers with stable, deployment-wide cardinality.
// Tenant shop domains, SSO IdPs, and other per-customer hosts are collapsed to
// "{host}" so Datadog resources stay bounded while shared APIs remain readable
// (e.g. "GET api.shopware.com/pluginStore/pluginsByName").
var sharedSpanHosts = map[string]struct{}{
	"api.shopware.com":          {},
	"releases.shopware.com":     {},
	"store.shopware.com":        {},
	"raw.githubusercontent.com": {},
}

// ClientSpanName formats an outbound HTTP client span as "METHOD host/path".
//
// otelhttp's default transport formatter uses only the method ("HTTP GET"),
// which Datadog collapses to a bare GET/POST resource. This formatter keeps
// cardinality low by:
//   - stripping query strings
//   - keeping only known shared hosts literal; other hosts become "{host}"
//   - replacing UUID / numeric / long-hex path segments with "{id}"
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

	host := spanHost(requestHost(r))
	path := spanPath(requestPath(r))

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

func spanHost(host string) string {
	if host == "" {
		return ""
	}
	host = strings.ToLower(host)
	if _, ok := sharedSpanHosts[host]; ok {
		return host
	}
	return "{host}"
}

func spanPath(path string) string {
	if path == "" || path == "/" {
		return "/"
	}

	leading := strings.HasPrefix(path, "/")
	parts := strings.Split(path, "/")
	for i, part := range parts {
		if part == "" {
			continue
		}
		if isHighCardinalityPathSegment(part) {
			parts[i] = "{id}"
		}
	}
	out := strings.Join(parts, "/")
	if leading && !strings.HasPrefix(out, "/") {
		return "/" + out
	}
	return out
}

func isHighCardinalityPathSegment(seg string) bool {
	if isUUIDSegment(seg) {
		return true
	}
	if isAllDigits(seg) {
		return true
	}
	// Long hex tokens (OAuth-ish / opaque IDs), but not short version-like
	// fragments such as "v1".
	if len(seg) >= 16 && isAllHex(seg) {
		return true
	}
	return false
}

func isUUIDSegment(seg string) bool {
	// 8-4-4-4-12 dashed UUID.
	if len(seg) == 36 {
		for i, r := range seg {
			switch i {
			case 8, 13, 18, 23:
				if r != '-' {
					return false
				}
			default:
				if !isHexRune(r) {
					return false
				}
			}
		}
		return true
	}
	// 32-char hex UUID without dashes.
	return len(seg) == 32 && isAllHex(seg)
}

func isAllDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}

func isAllHex(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !isHexRune(r) {
			return false
		}
	}
	return true
}

func isHexRune(r rune) bool {
	return unicode.IsDigit(r) || (r >= 'a' && r <= 'f') || (r >= 'A' && r <= 'F')
}
