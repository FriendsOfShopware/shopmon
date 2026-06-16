// Package shopwareaccount provides a client for the Shopware account/store API
// (api.shopware.com). This is distinct from internal/shopware, which talks to a
// single shop's admin API.
package shopwareaccount

import (
	"net/http"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// DefaultBaseURL is the public Shopware account/store API base URL.
const DefaultBaseURL = "https://api.shopware.com"

// Client talks to the Shopware account/store API.
type Client struct {
	baseURL    string
	httpClient *http.Client
}

// NewClient creates a Client for the given base URL. If httpClient is nil a
// default instrumented client with a 30s timeout is used.
func NewClient(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}
	if httpClient == nil {
		httpClient = httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second))
	}
	return &Client{
		baseURL:    strings.TrimRight(baseURL, "/"),
		httpClient: httpClient,
	}
}
