package shopware

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"golang.org/x/sync/singleflight"
)

// tokenExpiryMargin is subtracted from the OAuth token lifetime so a token is
// treated as expired slightly before the server would reject it, avoiding 401s
// from a token that lapses mid-request.
const tokenExpiryMargin = 30 * time.Second

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	shopToken    string
	httpClient   *http.Client

	// tokenGroup collapses concurrent token fetches (cold cache or post-401
	// re-auth) into a single in-flight request that all callers share.
	tokenGroup singleflight.Group

	mu          sync.Mutex
	cachedToken *tokenCache
}

type tokenCache struct {
	accessToken string
	expiresAt   time.Time
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
}

type ApiError struct {
	StatusCode int
	Body       string
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("shopware api error (status %d): %s", e.StatusCode, e.Body)
}

func NewClient(baseURL, clientID, clientSecret, shopToken string) *Client {
	return &Client{
		baseURL:      strings.TrimRight(baseURL, "/"),
		clientID:     clientID,
		clientSecret: clientSecret,
		shopToken:    shopToken,
		httpClient:   httputil.NewHTTPClient(httputil.WithTimeout(30 * time.Second)),
	}
}

// cachedAccessToken returns a still-valid cached token, or "" if none.
func (c *Client) cachedAccessToken() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cachedToken != nil && c.cachedToken.expiresAt.After(time.Now()) {
		return c.cachedToken.accessToken
	}
	return ""
}

// invalidateToken clears the cache only if it still holds stale, the token the
// caller observed. This prevents a 401 from one goroutine discarding a fresh
// token that another goroutine has already obtained.
func (c *Client) invalidateToken(stale string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cachedToken != nil && c.cachedToken.accessToken == stale {
		c.cachedToken = nil
	}
}

func (c *Client) getToken(ctx context.Context) (string, error) {
	if token := c.cachedAccessToken(); token != "" {
		return token, nil
	}

	// Collapse concurrent fetches: only one goroutine performs the network
	// round-trip; the rest wait and share its result. The HTTP call runs
	// outside c.mu so a slow token endpoint never blocks cache reads.
	token, err, _ := c.tokenGroup.Do("token", func() (interface{}, error) {
		// Another goroutine may have populated the cache while we waited.
		if token := c.cachedAccessToken(); token != "" {
			return token, nil
		}
		return c.fetchToken(ctx)
	})
	if err != nil {
		return "", err
	}
	return token.(string), nil
}

func (c *Client) fetchToken(ctx context.Context) (string, error) {
	body, err := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	})
	if err != nil {
		return "", fmt.Errorf("marshal token request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/oauth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("shopmon-shop-token", c.shopToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", &ApiError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	// Apply a safety margin, but never let it push expiry into the past for
	// short-lived tokens.
	lifetime := time.Duration(tokenResp.ExpiresIn) * time.Second
	if lifetime > tokenExpiryMargin {
		lifetime -= tokenExpiryMargin
	}

	c.mu.Lock()
	c.cachedToken = &tokenCache{
		accessToken: tokenResp.AccessToken,
		expiresAt:   time.Now().Add(lifetime),
	}
	c.mu.Unlock()

	return tokenResp.AccessToken, nil
}

// Authenticate tests the connection by fetching an access token.
// Returns nil on success or an error if authentication fails. The provided
// context bounds the token request so callers can cancel it (e.g. on shutdown
// or a per-scrape deadline).
func (c *Client) Authenticate(ctx context.Context) error {
	_, err := c.getToken(ctx)
	return err
}

var tracer = otel.Tracer("shopmon/shopware")

func (c *Client) request(ctx context.Context, method, path string, body interface{}, retry bool) ([]byte, error) {
	ctx, span := tracer.Start(ctx, method+" "+path,
		trace.WithAttributes(
			attribute.String("http.method", method),
			attribute.String("http.url", c.baseURL+"/api"+path),
		),
	)
	defer span.End()

	token, err := c.getToken(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	var reqBody io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal body: %w", err)
		}
		reqBody = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + "/api" + path
	req, err := http.NewRequestWithContext(ctx, method, url, reqBody)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("shopmon-shop-token", c.shopToken)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, fmt.Errorf("request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		return nil, &ApiError{StatusCode: resp.StatusCode, Body: "redirect detected"}
	}

	if resp.StatusCode == 401 && retry {
		c.invalidateToken(token)
		return c.request(ctx, method, path, body, false)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		apiErr := &ApiError{StatusCode: resp.StatusCode, Body: string(respBody)}
		span.RecordError(apiErr)
		span.SetStatus(codes.Error, apiErr.Error())
		return nil, apiErr
	}

	return respBody, nil
}

func (c *Client) Get(ctx context.Context, path string) ([]byte, error) {
	return c.request(ctx, "GET", path, nil, true)
}

func (c *Client) Post(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.request(ctx, "POST", path, body, true)
}

func (c *Client) Put(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.request(ctx, "PUT", path, body, true)
}

func (c *Client) Patch(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.request(ctx, "PATCH", path, body, true)
}

func (c *Client) Delete(ctx context.Context, path string, body interface{}) ([]byte, error) {
	return c.request(ctx, "DELETE", path, body, true)
}
