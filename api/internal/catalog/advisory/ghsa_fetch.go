package advisory

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/ghsa"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
)

// GHSAFetcher loads a single GitHub security advisory on demand for the
// admin import flow. It shares the sync's GitHub client, so the global
// endpoint, the repository-scoped fallback for advisories that were never
// promoted globally, token handling, and rate-limit errors all behave the
// same as in the enrichment queues.
type GHSAFetcher struct {
	client *githubClient
}

func NewGHSAFetcher(token string) *GHSAFetcher {
	httpClient := httputil.NewHTTPClient(httputil.WithTimeout(defaultHTTPTimeout))
	return &GHSAFetcher{client: newGitHubClient(token, defaultUserAgent, httpClient)}
}

// FetchGHSA loads the advisory named by ref, which may be a bare GHSA id or
// any GitHub advisory URL (global or repository-scoped).
func (f *GHSAFetcher) FetchGHSA(ctx context.Context, ref string) (*ghsa.Advisory, error) {
	ghsaID, link, err := ParseGHSAReference(ref)
	if err != nil {
		return nil, err
	}

	raw, err := f.client.fetchRawAdvisoryWithLink(ctx, ghsaID, link)
	if err != nil {
		return nil, fmt.Errorf("fetch github advisory %s: %w", ghsaID, err)
	}

	details := detailsFromResponse(*raw)
	fetched := &ghsa.Advisory{
		GHSAID:      canonicalGHSAID(raw.GHSAID),
		CVE:         strings.ToUpper(strings.TrimSpace(raw.CVEID)),
		Title:       details.Summary,
		Description: details.Description,
		Severity:    normalizeGitHubSeverity(raw.Severity),
		CVSSScore:   details.CVSSScore,
		CVSSVector:  details.CVSSVector,
		References:  details.References,
		HTMLURL:     details.HTMLURL,
		Withdrawn:   strings.TrimSpace(raw.WithdrawnAt) != "",
	}
	if fetched.GHSAID == "" {
		fetched.GHSAID = ghsaID
	}
	for _, cwe := range details.CWEs {
		fetched.CWEs = append(fetched.CWEs, ghsa.CWE{ID: cwe.ID, Name: cwe.Name})
	}
	if published, parseErr := time.Parse(time.RFC3339, strings.TrimSpace(raw.PublishedAt)); parseErr == nil {
		fetched.PublishedAt = &published
	}
	for _, vuln := range raw.Vulnerabilities {
		patched := strings.TrimSpace(vuln.PatchedVersions)
		if vuln.FirstPatchedVersion != nil {
			patched = strings.TrimSpace(string(*vuln.FirstPatchedVersion))
		}
		fetched.Packages = append(fetched.Packages, ghsa.Package{
			Name:                   strings.TrimSpace(vuln.Package.Name),
			Ecosystem:              strings.ToLower(strings.TrimSpace(vuln.Package.Ecosystem)),
			VulnerableVersionRange: strings.TrimSpace(vuln.VulnerableVersionRange),
			PatchedVersion:         patched,
		})
	}
	return fetched, nil
}

// ParseGHSAReference extracts the GHSA id (and, for repository-scoped URLs,
// the link the repository fallback needs) from an admin-supplied reference.
func ParseGHSAReference(ref string) (ghsaID, link string, err error) {
	ref = strings.TrimSpace(ref)
	if ref == "" {
		return "", "", fmt.Errorf("%w: empty reference", ghsa.ErrInvalidReference)
	}
	if !strings.Contains(ref, "/") {
		if !ghsaIDPattern.MatchString(strings.ToUpper(ref)) {
			return "", "", fmt.Errorf("%w: %q", ghsa.ErrInvalidReference, ref)
		}
		return canonicalGHSAID(ref), "", nil
	}
	if strings.Contains(ref, "://") {
		parsed, parseErr := url.Parse(ref)
		if parseErr != nil || parsed.Host != "github.com" {
			return "", "", fmt.Errorf("%w: %q", ghsa.ErrInvalidReference, ref)
		}
	}
	id := ghsaFromLink(ref)
	if id == "" {
		return "", "", fmt.Errorf("%w: %q", ghsa.ErrInvalidReference, ref)
	}
	return canonicalGHSAID(id), ref, nil
}

// canonicalGHSAID renders the spelling GitHub itself uses (uppercase prefix,
// lowercase segments) so stored ids compare equal to synced ones.
func canonicalGHSAID(id string) string {
	id = strings.TrimSpace(id)
	if len(id) <= len("GHSA-") {
		return id
	}
	return "GHSA-" + strings.ToLower(id[len("GHSA-"):])
}

func normalizeGitHubSeverity(severity string) string {
	switch strings.ToLower(strings.TrimSpace(severity)) {
	case "critical", "high", "low":
		return strings.ToLower(strings.TrimSpace(severity))
	case "medium", "moderate":
		return "medium"
	default:
		return ""
	}
}
