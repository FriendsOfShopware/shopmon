package advisory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/ghsa"
)

// ErrGHSAImportUnavailable marks a preview request on an instance without a
// GitHub fetcher configured, so the handler can answer 503.
var ErrGHSAImportUnavailable = errors.New("ghsa import is not available")

// ErrGHSAFetchFailed marks an upstream GitHub failure (rate limit, outage),
// so the handler can answer 502 instead of blaming the admin's input.
var ErrGHSAFetchFailed = errors.New("failed to fetch github advisory")

// GHSAPreviewPackage is one affected Composer package from the fetched
// advisory, including GitHub's patched-version claim for the review form.
type GHSAPreviewPackage struct {
	PackageName      string
	AffectedVersions string
	PatchedVersion   string
}

// GHSAPreview is a fetched GitHub advisory projected onto the create form's
// fields. Nothing is persisted by the preview; the admin reviews and submits
// through the import endpoint, which re-validates everything.
type GHSAPreview struct {
	GhsaID             string
	Cve                *string
	Title              string
	Description        *string
	Severity           *string
	CvssScore          *float64
	CvssVector         *string
	Cwes               []api.AdvisoryCWE
	Link               *string
	ExternalReferences []string
	ReportedAt         *time.Time
	Packages           []GHSAPreviewPackage
	// ExistingAdvisoryID names the catalog row already tracking this
	// GHSA/CVE when there is one; importing would duplicate it.
	ExistingAdvisoryID string
}

// PreviewGHSAImport fetches a GitHub advisory by id or URL and projects it
// for the admin create form. Only Composer packages are carried over: they
// are the only ones the SBOM matcher can ever match, and keeping e.g. an npm
// entry would seed a package row that looks authoritative but never fires.
func (s *Service) PreviewGHSAImport(ctx context.Context, ref string) (GHSAPreview, error) {
	if s.ghsa == nil {
		return GHSAPreview{}, ErrGHSAImportUnavailable
	}

	fetched, err := s.ghsa.FetchGHSA(ctx, ref)
	if err != nil {
		switch {
		case errors.Is(err, ghsa.ErrInvalidReference):
			return GHSAPreview{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
		case errors.Is(err, ghsa.ErrAdvisoryNotFound):
			return GHSAPreview{}, fmt.Errorf("%w: %v", ErrNotFound, err)
		default:
			return GHSAPreview{}, fmt.Errorf("%w: %w", ErrGHSAFetchFailed, err)
		}
	}
	if fetched.Withdrawn {
		return GHSAPreview{}, fmt.Errorf("%w: advisory %s was withdrawn", ErrInvalidInput, fetched.GHSAID)
	}

	preview := GHSAPreview{
		GhsaID:      fetched.GHSAID,
		Title:       fetched.Title,
		Description: nilIfEmpty(fetched.Description),
		CvssScore:   fetched.CVSSScore,
		CvssVector:  fetched.CVSSVector,
		Link:        nilIfEmpty(fetched.HTMLURL),
		ReportedAt:  fetched.PublishedAt,
	}
	if fetched.CVE != "" {
		if cve, err := normalizeCVE(&fetched.CVE); err == nil {
			preview.Cve = cve
		}
	}
	if severity, err := normalizeSeverity(&fetched.Severity); err == nil {
		preview.Severity = severity
	}
	for _, cwe := range fetched.CWEs {
		if strings.TrimSpace(cwe.ID) == "" {
			continue
		}
		preview.Cwes = append(preview.Cwes, api.AdvisoryCWE{Id: cwe.ID, Name: cwe.Name})
	}
	// The advisory page itself is the disclosure link, not a reference.
	for _, ref := range fetched.References {
		if ref == fetched.HTMLURL {
			continue
		}
		preview.ExternalReferences = append(preview.ExternalReferences, ref)
	}
	for _, pkg := range fetched.Packages {
		if pkg.Ecosystem != "composer" {
			continue
		}
		preview.Packages = append(preview.Packages, GHSAPreviewPackage{
			PackageName:      strings.ToLower(pkg.Name),
			AffectedVersions: pkg.VulnerableVersionRange,
			PatchedVersion:   pkg.PatchedVersion,
		})
	}
	if len(preview.Packages) == 0 {
		return GHSAPreview{}, fmt.Errorf("%w: advisory %s lists no affected Composer packages", ErrInvalidInput, fetched.GHSAID)
	}

	ghsaID := fetched.GHSAID
	existing, err := s.queries.ListComposerAdvisoriesByIdentifiers(ctx, queries.ListComposerAdvisoriesByIdentifiersParams{
		Cve:    preview.Cve,
		GhsaID: &ghsaID,
	})
	if err != nil {
		return GHSAPreview{}, fmt.Errorf("check advisory identifiers: %w", err)
	}
	if len(existing) > 0 {
		preview.ExistingAdvisoryID = existing[0].AdvisoryID
	}

	return preview, nil
}
