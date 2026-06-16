package shopwareaccount

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// StorePlugin is a plugin entry returned by the store pluginsByName endpoint.
type StorePlugin struct {
	Name          string           `json:"name"`
	Label         string           `json:"label"`
	Version       string           `json:"version"`
	RatingAverage float64          `json:"ratingAverage"`
	StoreLink     string           `json:"link"`
	Changelogs    []StoreChangelog `json:"changelog"`
}

// StoreChangelog is a single changelog entry for a store plugin.
type StoreChangelog struct {
	Version      string `json:"version"`
	Text         string `json:"text"`
	CreationDate struct {
		Date string `json:"date"`
	} `json:"creationDate"`
}

// PluginsByName fetches store metadata (latest version, rating, changelogs,
// store link) for the given technical names, scoped to a Shopware version.
func (c *Client) PluginsByName(ctx context.Context, locale, shopwareVersion string, technicalNames []string) ([]StorePlugin, error) {
	q := url.Values{}
	q.Set("locale", locale)
	q.Set("shopwareVersion", shopwareVersion)
	for _, name := range technicalNames {
		q.Add("technicalNames[]", name)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/pluginStore/pluginsByName?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch store plugins: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("store api returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	var plugins []StorePlugin
	if err := json.Unmarshal(body, &plugins); err != nil {
		return nil, fmt.Errorf("parse store plugins response: %w", err)
	}
	return plugins, nil
}
