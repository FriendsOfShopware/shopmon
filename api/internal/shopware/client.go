// Package shopware wires the Shopware Admin API client
// (github.com/FriendsOfShopware/go-shopware-http-client) into Shopmon.
package shopware

import (
	"time"

	swclient "github.com/FriendsOfShopware/go-shopware-http-client"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// ShopTokenHeader carries the environment token the Shopmon app plugin uses to
// identify requests from Shopmon. It is sent with every request, token
// requests included.
const ShopTokenHeader = "shopmon-shop-token"

// Client is the Shopware Admin API client. OAuth token caching, concurrent
// token fetch collapsing and the 401 re-auth retry are handled by the library.
type Client = swclient.Client

// Response is a raw Admin API response.
type Response = swclient.Response

// APIError describes a non-2xx response from the Admin API.
type APIError = swclient.APIError

// NewClient builds an Admin API client for one environment. Requests go
// through httputil's instrumented HTTP client so they are traced like every
// other outbound call.
func NewClient(baseURL, clientID, clientSecret, shopToken string) *Client {
	return swclient.NewClient(swclient.Config{
		BaseURL:     baseURL,
		Credentials: swclient.NewIntegrationCredentials(clientID, clientSecret),
		HTTPClient:  httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second)),
		Headers:     map[string]string{ShopTokenHeader: shopToken},
		UserAgent:   httputil.UserAgentString(),
	})
}
