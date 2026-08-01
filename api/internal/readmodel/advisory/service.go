package advisory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/jackc/pgx/v5"
	"github.com/russross/blackfriday/v2"
)

var ErrNotFound = errors.New("advisory not found")

type Service struct {
	queries *queries.Queries
}

func NewService(q *queries.Queries) *Service {
	return &Service{queries: q}
}

type ListParams struct {
	Limit       int32
	Offset      int32
	PackageName *string
	Severity    *string
	Tag         *string
	Search      *string
	// IsVisible filters by visibility when non-nil. User listing always passes true.
	IsVisible *bool
}

type AdvisoryView struct {
	Advisory queries.ComposerAdvisory
	Packages []queries.ComposerAdvisoryPackage
	// SecurityPluginFixes lists which SwagPlatformSecurity releases backport
	// this advisory, per branch. Nil when not loaded.
	SecurityPluginFixes []api.AdvisorySecurityPluginFix
	// SuppressedEnvironments counts the caller's environments where this
	// advisory has been acknowledged. Nil when not requested or unknown.
	SuppressedEnvironments *int
	// AffectedEnvironments counts the caller's environments whose Composer
	// inventory matches this advisory. Nil when not requested or unknown.
	AffectedEnvironments *int
}

func (s *Service) List(ctx context.Context, params ListParams) ([]AdvisoryView, int32, error) {
	if params.Limit <= 0 {
		params.Limit = 25
	}
	if params.Limit > 100 {
		params.Limit = 100
	}
	if params.Offset < 0 {
		params.Offset = 0
	}

	filter := queries.ListComposerAdvisoriesParams{
		Limit:       params.Limit,
		Offset:      params.Offset,
		IsVisible:   params.IsVisible,
		PackageName: emptyToNil(params.PackageName),
		Severity:    emptyToNil(params.Severity),
		Tag:         emptyToNil(params.Tag),
		Search:      emptyToNil(params.Search),
	}

	rows, err := s.queries.ListComposerAdvisories(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list advisories: %w", err)
	}

	total, err := s.queries.CountComposerAdvisories(ctx, queries.CountComposerAdvisoriesParams{
		IsVisible:   filter.IsVisible,
		PackageName: filter.PackageName,
		Severity:    filter.Severity,
		Tag:         filter.Tag,
		Search:      filter.Search,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("count advisories: %w", err)
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AdvisoryID)
	}
	pkgsByID, err := s.packagesByAdvisoryIDs(ctx, ids)
	if err != nil {
		return nil, 0, err
	}

	out := make([]AdvisoryView, 0, len(rows))
	for _, row := range rows {
		out = append(out, AdvisoryView{
			Advisory: row,
			Packages: pkgsByID[row.AdvisoryID],
		})
	}
	return out, total, nil
}

func (s *Service) Get(ctx context.Context, advisoryID string, visibleOnly bool) (AdvisoryView, error) {
	row, err := s.queries.GetComposerAdvisory(ctx, advisoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Allow old Packagist PKSA URLs to resolve to the canonical advisory.
			row, err = s.queries.GetComposerAdvisoryByPackagistID(ctx, advisoryID)
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return AdvisoryView{}, ErrNotFound
				}
				return AdvisoryView{}, fmt.Errorf("get advisory by packagist id: %w", err)
			}
		} else {
			return AdvisoryView{}, fmt.Errorf("get advisory: %w", err)
		}
	}
	if visibleOnly && !row.IsVisible {
		return AdvisoryView{}, ErrNotFound
	}
	pkgs, err := s.queries.ListComposerAdvisoryPackagesForAdvisory(ctx, row.AdvisoryID)
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("list advisory packages: %w", err)
	}
	packages := make([]queries.ComposerAdvisoryPackage, 0, len(pkgs))
	for _, p := range pkgs {
		packages = append(packages, queries.ComposerAdvisoryPackage{
			AdvisoryID:          row.AdvisoryID,
			PackageName:         p.PackageName,
			PackagistAdvisoryID: p.PackagistAdvisoryID,
			AffectedVersions:    p.AffectedVersions,
			SyncedAt:            p.SyncedAt,
		})
	}
	view := AdvisoryView{Advisory: row, Packages: packages}
	if row.GhsaID != nil && *row.GhsaID != "" {
		if fixes, err := s.securityPluginFixesByGhsa(ctx, []string{*row.GhsaID}); err == nil {
			view.SecurityPluginFixes = fixes[*row.GhsaID]
		}
	}
	return view, nil
}

func (s *Service) packagesByAdvisoryIDs(ctx context.Context, ids []string) (map[string][]queries.ComposerAdvisoryPackage, error) {
	out := make(map[string][]queries.ComposerAdvisoryPackage, len(ids))
	if len(ids) == 0 {
		return out, nil
	}
	rows, err := s.queries.ListComposerAdvisoryPackagesForAdvisories(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("list packages for advisories: %w", err)
	}
	for _, row := range rows {
		out[row.AdvisoryID] = append(out[row.AdvisoryID], row)
	}
	return out, nil
}

func (s *Service) ListPackages(ctx context.Context, visibleOnly bool) ([]string, error) {
	var isVisible *bool
	if visibleOnly {
		t := true
		isVisible = &t
	}
	names, err := s.queries.ListComposerAdvisoryPackages(ctx, isVisible)
	if err != nil {
		return nil, fmt.Errorf("list advisory packages: %w", err)
	}
	return names, nil
}

type EnrichmentUpdate struct {
	SeverityOverride      *string
	IsVisible             *bool
	NotesPublic           *string
	NotesInternal         *string
	RemediationSummary    *string
	RemediationUrl        *string
	RecommendedUpgrade    *string
	ShopwareImpactSummary *string
	AffectedComponents    *[]string
	Tags                  *[]string
	EnrichedBy            string
}

// UpdateEnrichment replaces all admin enrichment fields for an advisory.
// Nil string pointers clear the column; nil slices become empty arrays.
// IsVisible defaults to true when omitted.
func (s *Service) UpdateEnrichment(ctx context.Context, advisoryID string, update EnrichmentUpdate) (AdvisoryView, error) {
	view, err := s.Get(ctx, advisoryID, false)
	if err != nil {
		return AdvisoryView{}, err
	}
	advisoryID = view.Advisory.AdvisoryID

	var severityOverride *string
	if update.SeverityOverride != nil && strings.TrimSpace(*update.SeverityOverride) != "" {
		v := strings.ToLower(strings.TrimSpace(*update.SeverityOverride))
		severityOverride = &v
	}

	isVisible := true
	if update.IsVisible != nil {
		isVisible = *update.IsVisible
	}

	components := []string{}
	if update.AffectedComponents != nil && *update.AffectedComponents != nil {
		components = *update.AffectedComponents
	}
	tags := []string{}
	if update.Tags != nil && *update.Tags != nil {
		tags = *update.Tags
	}

	notesPublic := (*string)(nil)
	if update.NotesPublic != nil {
		notesPublic = nilIfEmpty(*update.NotesPublic)
	}
	notesInternal := (*string)(nil)
	if update.NotesInternal != nil {
		notesInternal = nilIfEmpty(*update.NotesInternal)
	}
	remediationSummary := (*string)(nil)
	if update.RemediationSummary != nil {
		remediationSummary = nilIfEmpty(*update.RemediationSummary)
	}
	remediationURL := (*string)(nil)
	if update.RemediationUrl != nil {
		remediationURL = nilIfEmpty(*update.RemediationUrl)
	}
	recommendedUpgrade := (*string)(nil)
	if update.RecommendedUpgrade != nil {
		recommendedUpgrade = nilIfEmpty(*update.RecommendedUpgrade)
	}
	impact := (*string)(nil)
	if update.ShopwareImpactSummary != nil {
		impact = nilIfEmpty(*update.ShopwareImpactSummary)
	}

	row, err := s.queries.UpdateComposerAdvisoryEnrichment(ctx, queries.UpdateComposerAdvisoryEnrichmentParams{
		AdvisoryID:            advisoryID,
		SeverityOverride:      severityOverride,
		IsVisible:             isVisible,
		NotesPublic:           notesPublic,
		NotesInternal:         notesInternal,
		RemediationSummary:    remediationSummary,
		RemediationUrl:        remediationURL,
		RecommendedUpgrade:    recommendedUpgrade,
		ShopwareImpactSummary: impact,
		AffectedComponents:    components,
		Tags:                  tags,
		EnrichedBy:            &update.EnrichedBy,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdvisoryView{}, ErrNotFound
		}
		return AdvisoryView{}, fmt.Errorf("update advisory enrichment: %w", err)
	}
	return AdvisoryView{Advisory: row, Packages: view.Packages}, nil
}

func ToUserAdvisory(view AdvisoryView) (api.Advisory, error) {
	row := view.Advisory
	sources, err := parseSources(row.Sources)
	if err != nil {
		return api.Advisory{}, err
	}
	cwes, err := parseCWEs(row.Cwes)
	if err != nil {
		return api.Advisory{}, err
	}
	refs, err := parseStringList(row.ExternalReferences)
	if err != nil {
		return api.Advisory{}, err
	}

	var descriptionHTML *string
	if row.Description != nil && strings.TrimSpace(*row.Description) != "" {
		html := renderMarkdown(*row.Description)
		descriptionHTML = &html
	}

	packages := make([]api.AdvisoryAffectedPackage, 0, len(view.Packages))
	for _, p := range view.Packages {
		packages = append(packages, api.AdvisoryAffectedPackage{
			PackageName:         p.PackageName,
			AffectedVersions:    p.AffectedVersions,
			PackagistAdvisoryId: p.PackagistAdvisoryID,
		})
	}

	return api.Advisory{
		AdvisoryId:                 row.AdvisoryID,
		Packages:                   packages,
		Title:                      row.Title,
		Link:                       row.Link,
		Cve:                        row.Cve,
		GhsaId:                     row.GhsaID,
		Severity:                   toSeverity(row.Severity),
		SeverityOverride:           toSeverity(row.SeverityOverride),
		EffectiveSeverity:          effectiveSeverity(row),
		Sources:                    sources,
		ReportedAt:                 database.TimePtr(row.ReportedAt),
		ComposerRepository:         row.ComposerRepository,
		Summary:                    row.Summary,
		Description:                row.Description,
		DescriptionHtml:            descriptionHTML,
		CvssScore:                  row.CvssScore,
		CvssVector:                 row.CvssVector,
		Cwes:                       &cwes,
		ExternalReferences:         &refs,
		DetailsSource:              row.DetailsSource,
		IsVisible:                  row.IsVisible,
		NotesPublic:                row.NotesPublic,
		RemediationSummary:         row.RemediationSummary,
		RemediationUrl:             row.RemediationUrl,
		RecommendedUpgrade:         row.RecommendedUpgrade,
		ShopwareImpactSummary:      row.ShopwareImpactSummary,
		AffectedComponents:         nonNilSlice(row.AffectedComponents),
		Tags:                       nonNilSlice(row.Tags),
		AffectedEnvironmentCount:   view.AffectedEnvironments,
		SuppressedEnvironmentCount: view.SuppressedEnvironments,
		SecurityPluginFixes:        &view.SecurityPluginFixes,
		FirstPatchedVersions:       parseFirstPatchedVersions(row.FirstPatchedVersions),
		SyncedAt:                   database.Time(row.SyncedAt),
		CreatedAt:                  database.Time(row.CreatedAt),
		UpdatedAt:                  database.Time(row.UpdatedAt),
	}, nil
}

func ToAdminAdvisory(view AdvisoryView) (api.AdminAdvisory, error) {
	base, err := ToUserAdvisory(view)
	if err != nil {
		return api.AdminAdvisory{}, err
	}
	return api.AdminAdvisory{
		AdvisoryId:            base.AdvisoryId,
		Packages:              base.Packages,
		Title:                 base.Title,
		Link:                  base.Link,
		Cve:                   base.Cve,
		GhsaId:                base.GhsaId,
		Severity:              base.Severity,
		SeverityOverride:      base.SeverityOverride,
		EffectiveSeverity:     base.EffectiveSeverity,
		Sources:               base.Sources,
		ReportedAt:            base.ReportedAt,
		ComposerRepository:    base.ComposerRepository,
		Summary:               base.Summary,
		Description:           base.Description,
		DescriptionHtml:       base.DescriptionHtml,
		CvssScore:             base.CvssScore,
		CvssVector:            base.CvssVector,
		Cwes:                  base.Cwes,
		ExternalReferences:    base.ExternalReferences,
		DetailsSource:         base.DetailsSource,
		IsVisible:             base.IsVisible,
		NotesPublic:           base.NotesPublic,
		NotesInternal:         view.Advisory.NotesInternal,
		RemediationSummary:    base.RemediationSummary,
		RemediationUrl:        base.RemediationUrl,
		RecommendedUpgrade:    base.RecommendedUpgrade,
		ShopwareImpactSummary: base.ShopwareImpactSummary,
		AffectedComponents:    base.AffectedComponents,
		Tags:                  base.Tags,
		SyncedAt:              base.SyncedAt,
		CreatedAt:             base.CreatedAt,
		UpdatedAt:             base.UpdatedAt,
		EnrichedAt:            database.TimePtr(view.Advisory.EnrichedAt),
		EnrichedBy:            view.Advisory.EnrichedBy,
	}, nil
}

func parseSources(raw json.RawMessage) ([]api.AdvisorySource, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []api.AdvisorySource{}, nil
	}
	var sources []struct {
		Name     string `json:"name"`
		RemoteID string `json:"remoteId"`
	}
	if err := json.Unmarshal(raw, &sources); err != nil {
		return nil, fmt.Errorf("decode advisory sources: %w", err)
	}
	out := make([]api.AdvisorySource, 0, len(sources))
	for _, s := range sources {
		out = append(out, api.AdvisorySource{Name: s.Name, RemoteId: s.RemoteID})
	}
	return out, nil
}

func parseCWEs(raw json.RawMessage) ([]api.AdvisoryCWE, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []api.AdvisoryCWE{}, nil
	}
	var items []struct {
		ID   string `json:"cwe_id"`
		Name string `json:"name"`
	}
	// Also accept API-shaped {id,name} after round-trips.
	if err := json.Unmarshal(raw, &items); err != nil {
		var alt []api.AdvisoryCWE
		if err2 := json.Unmarshal(raw, &alt); err2 != nil {
			return nil, fmt.Errorf("decode advisory cwes: %w", err)
		}
		return alt, nil
	}
	out := make([]api.AdvisoryCWE, 0, len(items))
	for _, item := range items {
		id := item.ID
		if id == "" {
			continue
		}
		out = append(out, api.AdvisoryCWE{Id: id, Name: item.Name})
	}
	return out, nil
}

func parseStringList(raw json.RawMessage) ([]string, error) {
	if len(raw) == 0 || string(raw) == "null" {
		return []string{}, nil
	}
	var out []string
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, fmt.Errorf("decode string list: %w", err)
	}
	if out == nil {
		return []string{}, nil
	}
	return out, nil
}

func renderMarkdown(md string) string {
	// CommonExtensions covers fenced code, tables, autolinks, etc. that GHSA
	// write-ups commonly use.
	return string(blackfriday.Run(
		[]byte(md),
		blackfriday.WithExtensions(blackfriday.CommonExtensions),
	))
}

func effectiveSeverity(row queries.ComposerAdvisory) *api.SeverityLevel {
	if row.SeverityOverride != nil && *row.SeverityOverride != "" {
		return toSeverity(row.SeverityOverride)
	}
	return toSeverity(row.Severity)
}

func toSeverity(s *string) *api.SeverityLevel {
	if s == nil || *s == "" {
		return nil
	}
	v := api.SeverityLevel(strings.ToLower(*s))
	if !v.Valid() {
		return nil
	}
	return &v
}

func nonNilSlice(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func emptyToNil(s *string) *string {
	if s == nil {
		return nil
	}
	if strings.TrimSpace(*s) == "" {
		return nil
	}
	v := strings.TrimSpace(*s)
	return &v
}

func nilIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := s
	return &v
}
