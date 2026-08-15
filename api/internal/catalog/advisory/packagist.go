package advisory

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"regexp"
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

// publicPackagesOf returns the subset of names that Packagist actually
// publishes, established by listing the vendor's public packages and
// intersecting locally.
//
// This is deliberately not a per-package existence probe: monitored shops run
// in-house Composer packages, and asking Packagist about "acme/secret-billing"
// would disclose that name whatever the answer. A vendor listing reveals only
// the vendor segment — which for a private package is typically the customer's
// own already-public brand — and never the package segment.
func (c *packagistClient) publicPackagesOf(ctx context.Context, names []string) []string {
	byVendor := make(map[string][]string)
	for _, name := range names {
		vendor, _, ok := strings.Cut(name, "/")
		if !ok || vendor == "" {
			continue
		}
		byVendor[vendor] = append(byVendor[vendor], name)
	}

	out := make([]string, 0, len(names))
	for vendor, candidates := range byVendor {
		published, err := c.listVendorPackages(ctx, vendor)
		if err != nil {
			// Unverifiable means excluded: the whole point is to not transmit a
			// name we have not confirmed is already public.
			slog.WarnContext(ctx, "failed to verify vendor packages, skipping",
				"vendor", vendor, "error", err)
			continue
		}

		public := make(map[string]bool, len(published))
		for _, name := range published {
			public[strings.ToLower(name)] = true
		}
		for _, name := range candidates {
			if public[strings.ToLower(name)] {
				out = append(out, name)
			}
		}
	}
	return out
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
	// Advisories sourced from FriendsOfPHP carry a remoteId like
	// "mcp/sdk/CVE-2026-53965.yaml" and no GHSA anywhere in Sources, but their
	// link still points at the GitHub advisory page. Without this the row never
	// enters the GHSA enrichment queue and stays permanently bare when NVD has
	// not published the CVE either.
	return ghsaFromLink(adv.Link)
}

// ghsaLinkPattern matches the GHSA id in a GitHub advisory URL, covering both
// the global (/advisories/GHSA-...) and repository-scoped
// (/owner/repo/security/advisories/GHSA-...) forms.
var ghsaLinkPattern = regexp.MustCompile(`(?i)/advisories/(GHSA(?:-[A-Za-z0-9]{4,})+)`)

// ghsaFromLink recovers a GHSA id from an advisory link. The result is matched
// against the same shape rule fetchAdvisory enforces, so a crafted link cannot
// smuggle path segments into a token-bearing request.
func ghsaFromLink(link string) string {
	m := ghsaLinkPattern.FindStringSubmatch(strings.TrimSpace(link))
	if m == nil {
		return ""
	}
	id := m[1]
	if !ghsaIDPattern.MatchString(strings.ToUpper(id)) {
		return ""
	}
	return id
}

// repoFromAdvisoryLink extracts owner/repo from a repository-scoped GitHub
// advisory URL. Global advisory links (github.com/advisories/GHSA-...) have no
// repository and yield empty strings.
func repoFromAdvisoryLink(link string) (owner, repo string) {
	m := repoAdvisoryLinkPattern.FindStringSubmatch(strings.TrimSpace(link))
	if m == nil {
		return "", ""
	}
	return m[1], m[2]
}

// Owner and repo are constrained to GitHub's own naming rules so neither can
// introduce a path segment when interpolated into the API URL.
var repoAdvisoryLinkPattern = regexp.MustCompile(
	`(?i)^https://github\.com/([A-Za-z0-9._-]+)/([A-Za-z0-9._-]+)/security/advisories/GHSA(?:-[A-Za-z0-9]{4,})+/?$`)

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
