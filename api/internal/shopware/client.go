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

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Client struct {
	baseURL      string
	clientID     string
	clientSecret string
	shopToken    string
	httpClient   *http.Client

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
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) getToken() (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.cachedToken != nil && c.cachedToken.expiresAt.After(time.Now()) {
		return c.cachedToken.accessToken, nil
	}

	body, _ := json.Marshal(map[string]string{
		"grant_type":    "client_credentials",
		"client_id":     c.clientID,
		"client_secret": c.clientSecret,
	})

	req, err := http.NewRequest("POST", c.baseURL+"/api/oauth/token", bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("shopmon-shop-token", c.shopToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("token request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		respBody, _ := io.ReadAll(resp.Body)
		return "", &ApiError{StatusCode: resp.StatusCode, Body: string(respBody)}
	}

	var tokenResp tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("decode token response: %w", err)
	}

	c.cachedToken = &tokenCache{
		accessToken: tokenResp.AccessToken,
		expiresAt:   time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second),
	}

	return c.cachedToken.accessToken, nil
}

// Authenticate tests the connection by fetching an access token.
// Returns nil on success or an error if authentication fails.
func (c *Client) Authenticate() error {
	_, err := c.getToken()
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

	token, err := c.getToken()
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
	defer resp.Body.Close()

	span.SetAttributes(attribute.Int("http.status_code", resp.StatusCode))

	if resp.StatusCode == 301 || resp.StatusCode == 302 {
		return nil, &ApiError{StatusCode: resp.StatusCode, Body: "redirect detected"}
	}

	if resp.StatusCode == 401 && retry {
		c.mu.Lock()
		c.cachedToken = nil
		c.mu.Unlock()
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
