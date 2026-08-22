package packagesapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
)

const maxResponseSize = 2 << 20

type Client struct {
	baseURL string
	token   string
	http    *http.Client
}

var _ packagesmirror.Gateway = (*Client)(nil)

func NewClient(baseURL, token string, httpClient *http.Client) *Client {
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), token: token, http: httpClient}
}

func (c *Client) List(ctx context.Context, source string) ([]packagesmirror.Token, error) {
	endpoint := c.baseURL + "/api/tokens?source=" + url.QueryEscape(source)
	body, err := c.request(ctx, http.MethodGet, endpoint, nil, http.StatusOK)
	if err != nil {
		return nil, err
	}
	var tokens []packagesmirror.Token
	if err := json.Unmarshal(body, &tokens); err != nil {
		return nil, fmt.Errorf("decode packages tokens: %w", err)
	}
	return tokens, nil
}

func (c *Client) Create(ctx context.Context, token, source string) error {
	payload, err := json.Marshal(map[string]string{"token": token, "source": source})
	if err != nil {
		return fmt.Errorf("encode packages token: %w", err)
	}
	_, err = c.request(ctx, http.MethodPost, c.baseURL+"/api/tokens", bytes.NewReader(payload), http.StatusOK, http.StatusCreated, http.StatusNoContent)
	return err
}

func (c *Client) Delete(ctx context.Context, tokenID int) error {
	endpoint := fmt.Sprintf("%s/api/tokens/%d", c.baseURL, tokenID)
	_, err := c.request(ctx, http.MethodDelete, endpoint, nil, http.StatusOK, http.StatusNoContent)
	return err
}

func (c *Client) Sync(ctx context.Context, tokenID int) error {
	endpoint := fmt.Sprintf("%s/api/tokens/%d/sync", c.baseURL, tokenID)
	_, err := c.request(ctx, http.MethodPost, endpoint, nil, http.StatusOK, http.StatusNoContent)
	return err
}

func (c *Client) request(ctx context.Context, method, endpoint string, body io.Reader, successCodes ...int) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("create packages API request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: execute request: %w", packagesmirror.ErrRemote, err)
	}
	defer func() { _ = resp.Body.Close() }()

	responseBody, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseSize+1))
	if err != nil {
		return nil, fmt.Errorf("%w: read response: %w", packagesmirror.ErrRemote, err)
	}
	if len(responseBody) > maxResponseSize {
		return nil, fmt.Errorf("%w: response exceeds %d bytes", packagesmirror.ErrRemote, maxResponseSize)
	}
	if slices.Contains(successCodes, resp.StatusCode) {
		return responseBody, nil
	}
	return nil, &packagesmirror.RemoteError{StatusCode: resp.StatusCode, Body: responseBody}
}
