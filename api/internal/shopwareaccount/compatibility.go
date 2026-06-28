package shopwareaccount

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// CompatibilityExtension is a single extension in a compatibility check.
type CompatibilityExtension struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// compatibilityRequest is the JSON body the Shopware auto-update API expects.
// The current Shopware version and language are passed as query parameters, not
// in the body (see CheckExtensionCompatibility).
type compatibilityRequest struct {
	FutureVersion string                   `json:"futureShopwareVersion"`
	Plugins       []CompatibilityExtension `json:"plugins"`
}

// CompatibilityResponse is the raw result of an autoupdate compatibility check.
// Body holds the unmodified response so callers can pass it through verbatim.
type CompatibilityResponse struct {
	StatusCode int
	Body       []byte
}

// CheckExtensionCompatibility calls the Shopware auto-update API to determine
// whether the given extensions are compatible with a target Shopware version.
// It returns the raw response so the caller controls response semantics.
func (c *Client) CheckExtensionCompatibility(ctx context.Context, currentVersion, futureVersion string, extensions []CompatibilityExtension) (*CompatibilityResponse, error) {
	reqBody, err := json.Marshal(compatibilityRequest{
		FutureVersion: futureVersion,
		Plugins:       extensions,
	})
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/swplatform/autoupdate", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	// The autoupdate API expects the current Shopware version and language as
	// query parameters (the body only carries futureShopwareVersion + plugins);
	// without `language` the store rejects the request with
	// REQUEST_PARAMETER_LANGUAGE_NOT_GIVEN. The response is version compatibility
	// data, not localized prose, so a fixed English locale is sufficient.
	q := req.URL.Query()
	q.Set("language", "en-GB")
	q.Set("shopwareVersion", currentVersion)
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call autoupdate api: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	return &CompatibilityResponse{
		StatusCode: resp.StatusCode,
		Body:       body,
	}, nil
}
