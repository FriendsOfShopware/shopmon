package advisory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/shyim/go-composer/repository"
)

const (
	defaultUserAgent        = "shopmon/1.0 (+https://github.com/FriendsOfShopware/shopmon)"
	defaultPackagistListURL = "https://packagist.org"
	maxResponseBytes        = 20 << 20 // 20 MB
	// Packagist security-advisories API accepts a packages[] list; batch to
	// keep request bodies reasonable.
	packageBatchSize = 50
)

// packagistClient loads shopware/* package names from Packagist's list API and
// security advisories via github.com/shyim/go-composer/repository.
type packagistClient struct {
	listBase   string // host for /packages/list.json (packagist.org)
	repo       *repository.Client
	userAgent  string
	httpClient *http.Client
}

type packageListResponse struct {
	PackageNames []string `json:"packageNames"`
}

func newPackagistClient(baseURL, userAgent string, httpClient *http.Client) *packagistClient {
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	// Production: repo.packagist.org for Composer protocol, packagist.org for
	// vendor package listing. Tests pass a single mock base for both.
	listBase := defaultPackagistListURL
	repoURL := repository.PackagistURL
	if baseURL != "" {
		base := strings.TrimRight(baseURL, "/")
		listBase = base
		repoURL = base
	}

	repo := repository.New(repoURL, nil)
	repo.HTTPClient = httpClient

	return &packagistClient{
		listBase:   listBase,
		repo:       repo,
		userAgent:  userAgent,
		httpClient: httpClient,
	}
}

func (c *packagistClient) listVendorPackages(ctx context.Context, vendor string) ([]string, error) {
	endpoint := fmt.Sprintf("%s/packages/list.json?vendor=%s", c.listBase, url.QueryEscape(vendor))
	var resp packageListResponse
	if err := c.getJSON(ctx, endpoint, &resp); err != nil {
		return nil, fmt.Errorf("list packages for vendor %s: %w", vendor, err)
	}
	return resp.PackageNames, nil
}

// advisoriesForPackages fetches full security advisories via go-composer's
// repository client (Composer security-advisories API, POST packages[]).
func (c *packagistClient) advisoriesForPackages(ctx context.Context, packages []string) (repository.Advisories, error) {
	if len(packages) == 0 {
		return repository.Advisories{}, nil
	}

	result := repository.Advisories{}
	for i := 0; i < len(packages); i += packageBatchSize {
		end := i + packageBatchSize
		if end > len(packages) {
			end = len(packages)
		}
		batch := packages[i:end]

		part, err := c.repo.GetSecurityAdvisories(ctx, batch)
		if err != nil {
			return nil, fmt.Errorf("fetch advisories for packages: %w", err)
		}
		for pkg, ads := range part {
			result[pkg] = append(result[pkg], ads...)
		}
	}
	return result, nil
}

func (c *packagistClient) getJSON(ctx context.Context, endpoint string, dst any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("Accept", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("perform request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes))
	if err != nil {
		return fmt.Errorf("read response: %w", err)
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decode response: %w", err)
	}
	return nil
}

func extractGHSAID(adv repository.SecurityAdvisory) string {
	for _, src := range adv.Sources {
		if id := normalizeGHSA(src.RemoteID); id != "" {
			return id
		}
	}
	if id := normalizeGHSA(adv.RemoteID); id != "" {
		return id
	}
	return ""
}

func normalizeGHSA(id string) string {
	id = strings.TrimSpace(id)
	if strings.HasPrefix(strings.ToUpper(id), "GHSA-") {
		return id
	}
	return ""
}

func parseReportedAt(value string) (time.Time, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return time.Time{}, false
	}
	layouts := []string{
		"2006-01-02 15:04:05",
		time.RFC3339,
		"2006-01-02T15:04:05Z",
		"2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.Parse(layout, value); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}
