package advisory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/ghsa"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/russross/blackfriday/v2"
)

var ErrNotFound = errors.New("advisory not found")

// ErrInvalidInput marks a rejected enrichment value, so the handler can answer
// 400 rather than 500.
var ErrInvalidInput = errors.New("invalid advisory input")

// ErrConflict marks a create/link that collides with an advisory already
// tracked under another id, so the handler can answer 409.
var ErrConflict = errors.New("advisory conflict")

// GHSAFetcher loads one GitHub security advisory for the admin import flow.
// Implemented by the catalog's GitHub client; nil disables the preview.
type GHSAFetcher interface {
	FetchGHSA(ctx context.Context, ref string) (*ghsa.Advisory, error)
}

type Service struct {
	// pool is needed alongside queries because advisory authoring (create,
	// content update, official link/merge) writes the advisory and its
	// package rows transactionally.
	pool    *pgxpool.Pool
	queries *queries.Queries
	ghsa    GHSAFetcher
}

func NewService(pool *pgxpool.Pool, q *queries.Queries, ghsa GHSAFetcher) *Service {
	return &Service{pool: pool, queries: q, ghsa: ghsa}
}

type ListParams struct {
	Limit       int32
	Offset      int32
	PackageName *string
	Severity    *string
	Tag         *string
	Search      *string
	// Origin filters to Shopmon-authored or upstream advisories when non-nil.
	Origin *string
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
		Origin:      emptyToNil(params.Origin),
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
		Origin:      filter.Origin,
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
		if !errors.Is(err, pgx.ErrNoRows) {
			return AdvisoryView{}, fmt.Errorf("get advisory: %w", err)
		}
		// Allow old Packagist PKSA URLs to resolve to the canonical advisory.
		row, err = s.queries.GetComposerAdvisoryByPackagistID(ctx, &advisoryID)
		if err != nil && !errors.Is(err, pgx.ErrNoRows) {
			return AdvisoryView{}, fmt.Errorf("get advisory by packagist id: %w", err)
		}
		if err != nil {
			// A linked Shopmon advisory keeps its SHOPMON-… id, so resolve it
			// by the official identifiers a user may arrive with instead.
			upper := strings.ToUpper(strings.TrimSpace(advisoryID))
			switch {
			case strings.HasPrefix(upper, "CVE-"):
				row, err = s.queries.GetComposerAdvisoryByCve(ctx, &upper)
			case strings.HasPrefix(upper, "GHSA-"):
				row, err = s.queries.GetComposerAdvisoryByGhsa(ctx, &advisoryID)
			}
			if err != nil {
				if errors.Is(err, pgx.ErrNoRows) {
					return AdvisoryView{}, ErrNotFound
				}
				return AdvisoryView{}, fmt.Errorf("get advisory by official id: %w", err)
			}
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

	// CoreFields, when non-nil, rewrites the advisory's synced content. It is
	// accepted only for manually tracked advisories (Shopmon-authored or
	// GitHub-imported) without a linked CVE; upstream rows take content from
	// the sync.
	CoreFields *CoreFieldsUpdate
}

// CoreFieldsUpdate is a partial edit of an advisory's content fields. The
// pointer semantics mirror the enrichment patch: omitted keeps the stored
// value, present-but-blank clears it (except the title, which must not be
// blank). CVSS uses a second pointer to distinguish omission from clearing.
type CoreFieldsUpdate struct {
	Title              *string
	Packages           *[]PackageInput
	Severity           *string
	Summary            *string
	Description        *string
	CvssScore          **float64
	CvssVector         *string
	Cwes               *[]api.AdvisoryCWE
	Link               *string
	ExternalReferences *[]string
	ReportedAt         *time.Time
}

// UpdateEnrichment applies admin enrichment as a partial update: a field the
// caller omitted keeps its stored value, and an explicitly empty string clears
// it. The endpoint is a PATCH with every field optional, so treating omission
// as "clear" would let a request that only sets tags silently wipe an
// advisory's notes — or, worse, republish one an admin had deliberately hidden.
func (s *Service) UpdateEnrichment(ctx context.Context, advisoryID string, update EnrichmentUpdate) (AdvisoryView, error) {
	// Resolve aliases (PKSA, linked CVE/GHSA) to the canonical id first; all
	// writes below key off it.
	view, err := s.Get(ctx, advisoryID, false)
	if err != nil {
		return AdvisoryView{}, err
	}
	advisoryID = view.Advisory.AdvisoryID

	if s.pool == nil {
		return AdvisoryView{}, errors.New("update advisory: database pool is required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("begin advisory update transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := s.queries.WithTx(tx)

	// Lock the row so the omitted-keeps-value patch logic reads the values it
	// is about to overwrite, not a concurrent admin's half-applied edit.
	current, err := txq.GetComposerAdvisoryForUpdate(ctx, advisoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdvisoryView{}, ErrNotFound
		}
		return AdvisoryView{}, fmt.Errorf("lock advisory: %w", err)
	}

	severityOverride := current.SeverityOverride
	if update.SeverityOverride != nil {
		trimmed := strings.ToLower(strings.TrimSpace(*update.SeverityOverride))
		if trimmed == "" {
			// An explicit empty value clears the override.
			severityOverride = nil
		} else {
			// Reject unknown levels rather than storing a value that
			// toSeverity would silently drop on read, leaving writes and
			// reads disagreeing.
			if !api.SeverityLevel(trimmed).Valid() {
				return AdvisoryView{}, fmt.Errorf("%w: severity override %q", ErrInvalidInput, trimmed)
			}
			severityOverride = &trimmed
		}
	}

	isVisible := current.IsVisible
	if update.IsVisible != nil {
		isVisible = *update.IsVisible
	}

	components := current.AffectedComponents
	if update.AffectedComponents != nil {
		components = *update.AffectedComponents
	}
	if components == nil {
		components = []string{}
	}
	tags := current.Tags
	if update.Tags != nil {
		tags = *update.Tags
	}
	if tags == nil {
		tags = []string{}
	}

	notesPublic := patchString(current.NotesPublic, update.NotesPublic)
	notesInternal := patchString(current.NotesInternal, update.NotesInternal)
	remediationSummary := patchString(current.RemediationSummary, update.RemediationSummary)
	remediationURL := patchString(current.RemediationUrl, update.RemediationUrl)
	recommendedUpgrade := patchString(current.RecommendedUpgrade, update.RecommendedUpgrade)
	impact := patchString(current.ShopwareImpactSummary, update.ShopwareImpactSummary)

	row, err := txq.UpdateComposerAdvisoryEnrichment(ctx, queries.UpdateComposerAdvisoryEnrichmentParams{
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

	packages := view.Packages
	if update.CoreFields != nil {
		row, packages, err = s.applyCoreFieldsUpdate(ctx, txq, current, row, update.CoreFields)
		if err != nil {
			return AdvisoryView{}, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return AdvisoryView{}, fmt.Errorf("commit advisory update transaction: %w", err)
	}
	return AdvisoryView{Advisory: row, Packages: packages}, nil
}

// applyCoreFieldsUpdate rewrites a manually tracked advisory's content
// inside the enrichment transaction. The guards keep admin edits and the
// Packagist sync from fighting over the same columns: once an official CVE
// is linked, the sync owns the content and only enrichment stays editable.
func (s *Service) applyCoreFieldsUpdate(ctx context.Context, txq *queries.Queries, current, enriched queries.ComposerAdvisory, core *CoreFieldsUpdate) (queries.ComposerAdvisory, []queries.ComposerAdvisoryPackage, error) {
	if current.Origin != originShopmon && current.Origin != originGitHub {
		return enriched, nil, fmt.Errorf("%w: core fields can only be edited on Shopmon-authored or GitHub-imported advisories", ErrInvalidInput)
	}
	if contentLocked(current) {
		return enriched, nil, fmt.Errorf("%w: advisory %s is linked to an official CVE, so its core fields are maintained by the sync", ErrConflict, current.AdvisoryID)
	}

	content := contentParamsFromRow(enriched)

	if core.Title != nil {
		title := strings.TrimSpace(*core.Title)
		if title == "" {
			return enriched, nil, fmt.Errorf("%w: title must not be empty", ErrInvalidInput)
		}
		content.Title = title
	}
	if core.Severity != nil {
		severity, err := normalizeSeverity(core.Severity)
		if err != nil {
			return enriched, nil, err
		}
		content.Severity = severity
	}
	content.Summary = patchString(content.Summary, core.Summary)
	content.Description = patchString(content.Description, core.Description)
	content.Link = patchString(content.Link, core.Link)
	if content.Link != nil {
		if err := validateHTTPURL("link", *content.Link); err != nil {
			return enriched, nil, err
		}
	}
	if core.CvssScore != nil {
		if err := validateCVSSScore(*core.CvssScore); err != nil {
			return enriched, nil, err
		}
		content.CvssScore = *core.CvssScore
	}
	content.CvssVector = patchString(content.CvssVector, core.CvssVector)
	if core.Cwes != nil {
		cwesJSON, err := encodeCWEs(*core.Cwes)
		if err != nil {
			return enriched, nil, err
		}
		content.Cwes = cwesJSON
	}
	if core.ExternalReferences != nil {
		refsJSON, err := encodeReferences(*core.ExternalReferences)
		if err != nil {
			return enriched, nil, err
		}
		content.ExternalReferences = refsJSON
	}
	if core.ReportedAt != nil {
		content.ReportedAt = pgtype.Timestamp{Time: core.ReportedAt.UTC(), Valid: true}
	}
	// Hand-written details are the disclosure until an official source
	// replaces them after linking. The source follows the row's origin so a
	// GitHub import stays attributed to GitHub after an admin edit.
	if content.Summary != nil || content.Description != nil || content.CvssScore != nil {
		source := detailsSourceShopmon
		if current.Origin == originGitHub {
			source = detailsSourceGitHub
		}
		content.DetailsSource = &source
	}

	row, err := txq.UpdateComposerAdvisoryContent(ctx, content)
	if err != nil {
		return enriched, nil, fmt.Errorf("update advisory content: %w", err)
	}

	if core.Packages == nil {
		pkgs, err := s.listPackagesTx(ctx, txq, row.AdvisoryID)
		if err != nil {
			return enriched, nil, err
		}
		return row, pkgs, nil
	}

	packages, err := normalizePackages(*core.Packages)
	if err != nil {
		return enriched, nil, err
	}
	if err := txq.DeleteComposerAdvisoryPackages(ctx, row.AdvisoryID); err != nil {
		return enriched, nil, fmt.Errorf("replace advisory packages: %w", err)
	}
	out := make([]queries.ComposerAdvisoryPackage, 0, len(packages))
	for _, pkg := range packages {
		if err := txq.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
			AdvisoryID:          row.AdvisoryID,
			PackageName:         pkg.PackageName,
			PackagistAdvisoryID: nil,
			AffectedVersions:    pkg.AffectedVersions,
		}); err != nil {
			return enriched, nil, fmt.Errorf("replace advisory packages: %w", err)
		}
		out = append(out, queries.ComposerAdvisoryPackage{
			AdvisoryID:       row.AdvisoryID,
			PackageName:      pkg.PackageName,
			AffectedVersions: pkg.AffectedVersions,
		})
	}
	return row, out, nil
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
		Origin:                     toOrigin(row.Origin),
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
		Origin:                base.Origin,
		CreatedBy:             view.Advisory.CreatedBy,
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

// patchString resolves one optional PATCH field: omitted keeps the stored
// value, present-but-blank clears it, anything else replaces it.
func patchString(current, update *string) *string {
	if update == nil {
		return current
	}
	return nilIfEmpty(*update)
}

func nilIfEmpty(s string) *string {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	v := s
	return &v
}

// toOrigin maps the stored origin onto the API enum. Unknown values (which
// the CHECK constraint should prevent) fall back to upstream rather than
// leaking an out-of-enum value into the API.
func toOrigin(origin string) api.AdvisoryOrigin {
	if v := api.AdvisoryOrigin(origin); v.Valid() {
		return v
	}
	return api.AdvisoryOriginUpstream
}
