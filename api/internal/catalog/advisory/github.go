package advisory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/version"
)

const (
	defaultGitHubAPIBase = "https://api.github.com"
	maxGitHubBodyBytes   = 5 << 20
)

type githubAdvisoryDetails struct {
	Summary     string
	Description string
	CVSSScore   *float64
	CVSSVector  *string
	CWEs        []githubCWE
	References  []string
	HTMLURL     string
	// FirstPatchedVersions maps a Shopware line ("6.7") to the first release
	// that closed the advisory. GitHub can list several patched releases for one
	// line, so the minimum is kept: it is the earliest version that suffices.
	FirstPatchedVersions map[string]string
}

type githubCWE struct {
	ID   string `json:"cwe_id"`
	Name string `json:"name"`
}

type githubAdvisoryResponse struct {
	GHSAID      string `json:"ghsa_id"`
	Summary     string `json:"summary"`
	Description string `json:"description"`
	HTMLURL     string `json:"html_url"`
	CVSS        *struct {
		Score        float64 `json:"score"`
		VectorString string  `json:"vector_string"`
	} `json:"cvss"`
	CVSSSeverities *struct {
		CVSSV3 *struct {
			Score        float64 `json:"score"`
			VectorString string  `json:"vector_string"`
		} `json:"cvss_v3"`
	} `json:"cvss_severities"`
	CWEs []struct {
		CWEID string `json:"cwe_id"`
		Name  string `json:"name"`
	} `json:"cwes"`
	// GitHub returns references either as plain URL strings or as
	// {"url":"..."} objects depending on the advisory/source.
	References githubReferenceList `json:"references"`
	// Vulnerabilities carries the per-package affected range and the release
	// that first fixed it, e.g. 6.7.10.1 for the 6.7 line and 6.6.10.18 for 6.6.
	// This is the authoritative "upgrade to" answer and needs no extra request.
	Vulnerabilities []struct {
		Package struct {
			Name      string `json:"name"`
			Ecosystem string `json:"ecosystem"`
		} `json:"package"`
		VulnerableVersionRange string  `json:"vulnerable_version_range"`
		FirstPatchedVersion    *string `json:"first_patched_version"`
		// The repository-scoped endpoint names the same value differently.
		PatchedVersions string `json:"patched_versions"`
	} `json:"vulnerabilities"`
}

// githubReferenceList decodes GitHub advisory "references" values that may be
// either strings or objects with a "url" field.
type githubReferenceList []string

func (r *githubReferenceList) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*r = nil
		return nil
	}
	var raw []json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	out := make([]string, 0, len(raw))
	for _, item := range raw {
		var asString string
		if err := json.Unmarshal(item, &asString); err == nil {
			if u := strings.TrimSpace(asString); u != "" {
				out = append(out, u)
			}
			continue
		}
		var asObj struct {
			URL string `json:"url"`
		}
		if err := json.Unmarshal(item, &asObj); err != nil {
			// Skip unknown shapes rather than failing the whole advisory.
			continue
		}
		if u := strings.TrimSpace(asObj.URL); u != "" {
			out = append(out, u)
		}
	}
	*r = out
	return nil
}

type githubClient struct {
	baseURL    string
	token      string
	userAgent  string
	httpClient *http.Client
}

func newGitHubClient(token, userAgent string, httpClient *http.Client) *githubClient {
	if userAgent == "" {
		userAgent = defaultUserAgent
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}
	return &githubClient{
		baseURL:    defaultGitHubAPIBase,
		token:      strings.TrimSpace(token),
		userAgent:  userAgent,
		httpClient: httpClient,
	}
}

// ghsaIDPattern constrains the id to GitHub's advisory shape. IDs reach us from
// Packagist source records, which are third-party data: without this check a
// value like "GHSA-x/../../user/repo" would be interpolated into the request
// path and point an authenticated, token-bearing request at another endpoint.
//
// Deliberately shape-only (alphanumeric segments) rather than GitHub's exact
// base32 alphabet: anything that could traverse or reshape a path is already
// excluded, and a stricter rule would silently drop advisories if the id
// encoding ever widens.
var ghsaIDPattern = regexp.MustCompile(`^GHSA(-[A-Za-z0-9]{4,})+$`)

// fetchAdvisory loads a global GitHub advisory. Advisories that were published
// on a repository but never promoted to the global database 404 here; use
// fetchAdvisoryWithLink to fall back to the repository-scoped endpoint.
func (c *githubClient) fetchAdvisory(ctx context.Context, ghsaID string) (*githubAdvisoryDetails, error) {
	return c.fetchAdvisoryWithLink(ctx, ghsaID, "")
}

// fetchAdvisoryWithLink tries the global advisories endpoint first and, when
// that reports the advisory unknown, retries against the repository named in
// link. Repository advisories are not mirrored globally and are not indexed by
// cve_id or affects, so the link is the only route to them.
func (c *githubClient) fetchAdvisoryWithLink(ctx context.Context, ghsaID, link string) (*githubAdvisoryDetails, error) {
	ghsaID = strings.TrimSpace(ghsaID)
	if ghsaID == "" {
		return nil, fmt.Errorf("empty ghsa id")
	}
	if !ghsaIDPattern.MatchString(strings.ToUpper(ghsaID)) {
		return nil, fmt.Errorf("malformed ghsa id %q", ghsaID)
	}

	// Escaped as well as validated: defence in depth if the pattern is ever
	// loosened.
	endpoint := fmt.Sprintf("%s/advisories/%s", c.baseURL, url.PathEscape(ghsaID))
	details, err := c.fetchAdvisoryURL(ctx, ghsaID, endpoint)
	if err == nil || !errors.Is(err, errGitHubAdvisoryNotFound) {
		return details, err
	}

	owner, repo := repoFromAdvisoryLink(link)
	if owner == "" || repo == "" {
		return nil, err
	}
	repoEndpoint := fmt.Sprintf("%s/repos/%s/%s/security-advisories/%s",
		c.baseURL, url.PathEscape(owner), url.PathEscape(repo), url.PathEscape(ghsaID))
	repoDetails, repoErr := c.fetchAdvisoryURL(ctx, ghsaID, repoEndpoint)
	if repoErr != nil {
		return nil, repoErr
	}
	return repoDetails, nil
}

// errGitHubAdvisoryNotFound signals a 404 so callers can decide whether another
// endpoint is worth trying.
var errGitHubAdvisoryNotFound = errors.New("github advisory not found")

func (c *githubClient) fetchAdvisoryURL(ctx context.Context, ghsaID, endpoint string) (*githubAdvisoryDetails, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create github advisory request: %w", err)
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", c.userAgent)
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch github advisory %s: %w", ghsaID, err)
	}
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(io.LimitReader(resp.Body, maxGitHubBodyBytes))
	if err != nil {
		return nil, fmt.Errorf("read github advisory %s: %w", ghsaID, err)
	}
	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("github advisory %s: %w", ghsaID, errGitHubAdvisoryNotFound)
	}
	if resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusTooManyRequests {
		return nil, fmt.Errorf("github rate limited or forbidden (%d) for %s", resp.StatusCode, ghsaID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("github advisory %s status %d", ghsaID, resp.StatusCode)
	}

	var raw githubAdvisoryResponse
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode github advisory %s: %w", ghsaID, err)
	}

	details := &githubAdvisoryDetails{
		Summary:     strings.TrimSpace(raw.Summary),
		Description: strings.TrimSpace(raw.Description),
		HTMLURL:     strings.TrimSpace(raw.HTMLURL),
	}

	if raw.CVSSSeverities != nil && raw.CVSSSeverities.CVSSV3 != nil && raw.CVSSSeverities.CVSSV3.Score > 0 {
		score := raw.CVSSSeverities.CVSSV3.Score
		details.CVSSScore = &score
		if v := strings.TrimSpace(raw.CVSSSeverities.CVSSV3.VectorString); v != "" {
			details.CVSSVector = &v
		}
	} else if raw.CVSS != nil && raw.CVSS.Score > 0 {
		score := raw.CVSS.Score
		details.CVSSScore = &score
		if v := strings.TrimSpace(raw.CVSS.VectorString); v != "" {
			details.CVSSVector = &v
		}
	}

	for _, cwe := range raw.CWEs {
		details.CWEs = append(details.CWEs, githubCWE{
			ID:   cwe.CWEID,
			Name: cwe.Name,
		})
	}
	details.References = append(details.References, raw.References...)
	if details.HTMLURL != "" {
		// Prefer GHSA page as a stable first reference.
		found := false
		for _, r := range details.References {
			if r == details.HTMLURL {
				found = true
				break
			}
		}
		if !found {
			details.References = append([]string{details.HTMLURL}, details.References...)
		}
	}

	details.FirstPatchedVersions = firstPatchedByLine(raw)

	return details, nil
}

// firstPatchedByLine folds GitHub's per-package patched versions into one entry
// per Shopware line. Only shopware/* packages are considered; other ecosystems
// in the same advisory are irrelevant to a core upgrade recommendation.
//
// The minimum per line is kept because GitHub may list several patched releases
// for one branch, and the earliest is the least disruptive upgrade that works.
func firstPatchedByLine(raw githubAdvisoryResponse) map[string]string {
	out := map[string]string{}
	for _, vuln := range raw.Vulnerabilities {
		patched := strings.TrimSpace(vuln.PatchedVersions)
		if vuln.FirstPatchedVersion != nil {
			patched = strings.TrimSpace(*vuln.FirstPatchedVersion)
		}
		if patched == "" {
			continue
		}
		name := strings.ToLower(strings.TrimSpace(vuln.Package.Name))
		if !strings.HasPrefix(name, packagePrefix) {
			continue
		}
		line := shopwareLine(patched)
		if line == "" {
			continue
		}
		if current, ok := out[line]; !ok || version.Compare(patched, current) < 0 {
			out[line] = patched
		}
	}
	return out
}

// shopwareLine reduces a version to its major.minor branch, the granularity at
// which Shopware ships security releases.
func shopwareLine(v string) string {
	parts := strings.SplitN(strings.TrimSpace(v), ".", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return ""
	}
	return parts[0] + "." + parts[1]
}
