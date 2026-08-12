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
			name:   "host and path",
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
			name:   "empty path",
			method: http.MethodGet,
			rawURL: "https://api.shopware.com",
			want:   "GET api.shopware.com/",
		},
		{
			name:   "empty path with trailing slash equivalent",
			method: http.MethodHead,
			rawURL: "https://example.com/",
			want:   "HEAD example.com/",
		},
		{
			name:   "host from Request.Host when URL host empty",
			method: http.MethodPut,
			rawURL: "/api/oauth/token",
			host:   "shop.example:443",
			want:   "PUT shop.example/api/oauth/token",
		},
		{
			name:   "port stripped from URL host",
			method: http.MethodGet,
			rawURL: "http://127.0.0.1:8080/_info/config",
			want:   "GET 127.0.0.1/_info/config",
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
