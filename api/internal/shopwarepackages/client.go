// Package shopwarepackages is a client for the public packages.shopware.com
// Composer repository. Its per-package feeds list every released version with
// its changelog and require constraints, without the authentication and rate
// limits of the store API (api.shopware.com).
package shopwarepackages

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// DefaultBaseURL is the public Shopware Composer repository base URL.
const DefaultBaseURL = "https://packages.shopware.com"

// maxFeedResponseBytes bounds how much we read from a single feed response, so
// a misbehaving upstream can't exhaust worker memory.
const maxFeedResponseBytes = 10 << 20 // 10 MB

// ErrNotFound marks a feed that packages.shopware.com does not serve (HTTP
// 404) — the normal case for custom plugins that are not distributed through
// the store's Composer repository.
var ErrNotFound = errors.New("package feed not found")

// Client fetches package feeds from packages.shopware.com.
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

// Package is a package feed: every released version with its changelog and
// Composer require constraints.
type Package struct {
	Name     string
	Link     string
	Versions []PackageVersion
}

// PackageVersion is one release in a package feed.
type PackageVersion struct {
	Version    string
	Changelog  string
	ReleasedAt *time.Time
	Require    map[string]string
}

// rawFeed mirrors the feed JSON document served at
// {baseURL}/feeds/package/{name}.json.
type rawFeed struct {
	Name     string `json:"name"`
	Link     string `json:"link"`
	Versions []struct {
		Version   string            `json:"version"`
		Date      string            `json:"date"`
		Changelog string            `json:"changelog"`
		Require   map[string]string `json:"require"`
	} `json:"versions"`
}

// PackageFeed fetches the feed for a Composer package name such as
// "store.shopware.com/swagplatformsecurity".
func (c *Client) PackageFeed(ctx context.Context, name string) (*Package, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/feeds/package/"+name+".json", nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch package feed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("fetch package feed %s: unexpected status %d", name, resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxFeedResponseBytes))
	if err != nil {
		return nil, fmt.Errorf("read package feed: %w", err)
	}

	var raw rawFeed
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse package feed: %w", err)
	}

	pkg := &Package{
		Name:     raw.Name,
		Link:     raw.Link,
		Versions: make([]PackageVersion, 0, len(raw.Versions)),
	}
	for _, v := range raw.Versions {
		pkg.Versions = append(pkg.Versions, PackageVersion{
			Version:    v.Version,
			Changelog:  v.Changelog,
			ReleasedAt: parseFeedDate(v.Date),
			Require:    v.Require,
		})
	}
	return pkg, nil
}

func parseFeedDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	if t, err := time.Parse(time.RFC3339, value); err == nil {
		utc := t.UTC()
		return &utc
	}
	return nil
}
