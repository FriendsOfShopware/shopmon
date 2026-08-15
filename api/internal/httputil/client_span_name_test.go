package httputil

import (
	"net/http"
	"net/url"
	"testing"
)

func TestClientSpanName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		method string
		rawURL string
		host   string // optional Request.Host override when URL has no host
		want   string
	}{
		{
			name:   "shared host and path",
			method: http.MethodGet,
			rawURL: "https://api.shopware.com/pluginStore/pluginsByName",
			want:   "GET api.shopware.com/pluginStore/pluginsByName",
		},
		{
			name:   "query string stripped",
			method: http.MethodGet,
			rawURL: "https://api.shopware.com/pluginStore/pluginsByName?locale=en-GB&shopId=abc",
			want:   "GET api.shopware.com/pluginStore/pluginsByName",
		},
		{
			name:   "missing host falls back to path",
			method: http.MethodPost,
			rawURL: "/swplatform/autoupdate",
			want:   "POST /swplatform/autoupdate",
		},
		{
			name:   "empty path on shared host",
			method: http.MethodGet,
			rawURL: "https://api.shopware.com",
			want:   "GET api.shopware.com/",
		},
		{
			name:   "tenant host collapsed",
			method: http.MethodHead,
			rawURL: "https://shop.example.com/",
			want:   "HEAD {host}/",
		},
		{
			name:   "host from Request.Host when URL host empty",
			method: http.MethodPut,
			rawURL: "/api/oauth/token",
			host:   "shop.example:443",
			want:   "PUT {host}/api/oauth/token",
		},
		{
			name:   "non-shared IP host collapsed and port stripped",
			method: http.MethodGet,
			rawURL: "http://127.0.0.1:8080/_info/config",
			want:   "GET {host}/_info/config",
		},
		{
			name:   "uuid path segment grouped",
			method: http.MethodPatch,
			rawURL: "https://shop.example.com/api/scheduled-task/550e8400-e29b-41d4-a716-446655440000",
			want:   "PATCH {host}/api/scheduled-task/{id}",
		},
		{
			name:   "numeric path segment grouped",
			method: http.MethodGet,
			rawURL: "https://sitespeed.internal/api/result/42",
			want:   "GET {host}/api/result/{id}",
		},
		{
			name:   "long hex path segment grouped",
			method: http.MethodGet,
			rawURL: "https://shop.example.com/api/token/0123456789abcdef0123456789abcdef",
			want:   "GET {host}/api/token/{id}",
		},
		{
			name:   "shared github host kept literal",
			method: http.MethodGet,
			rawURL: "https://raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/security.json",
			want:   "GET raw.githubusercontent.com/FriendsOfShopware/shopware-static-data/main/data/security.json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			u, err := url.Parse(tt.rawURL)
			if err != nil {
				t.Fatalf("parse URL: %v", err)
			}
			req := &http.Request{
				Method: tt.method,
				URL:    u,
				Host:   tt.host,
			}

			got := ClientSpanName("", req)
			if got != tt.want {
				t.Fatalf("ClientSpanName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestClientSpanName_NilRequest(t *testing.T) {
	t.Parallel()
	if got := ClientSpanName("", nil); got != http.MethodGet {
		t.Fatalf("ClientSpanName(nil) = %q, want %q", got, http.MethodGet)
	}
}

func TestSpanPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		in   string
		want string
	}{
		{"", "/"},
		{"/", "/"},
		{"/pluginStore/pluginsByName", "/pluginStore/pluginsByName"},
		{"/api/result/99", "/api/result/{id}"},
		{"/api/scheduled-task/550e8400-e29b-41d4-a716-446655440000/run", "/api/scheduled-task/{id}/run"},
		{"/_info/config", "/_info/config"},
	}
	for _, tt := range tests {
		if got := spanPath(tt.in); got != tt.want {
			t.Fatalf("spanPath(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}
