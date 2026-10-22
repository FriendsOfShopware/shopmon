package advisory

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

const (
	originShopmon        = "shopmon"
	originGitHub         = "github"
	detailsSourceShopmon = "shopmon"
	detailsSourceGitHub  = "github"
)

var (
	cvePattern         = regexp.MustCompile(`^CVE-\d{4}-\d{4,}$`)
	ghsaPattern        = regexp.MustCompile(`(?i)^GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}$`)
	cwePattern         = regexp.MustCompile(`(?i)^CWE-\d+$`)
	packageNamePattern = regexp.MustCompile(`^[a-z0-9]([a-z0-9._-]*[a-z0-9])?/[a-z0-9]([a-z0-9._-]*[a-z0-9])?$`)
)

// PackageInput is one affected Composer package of a Shopmon-authored advisory.
type PackageInput struct {
	PackageName      string
	AffectedVersions string
}

// CreateInput carries everything an admin can author on a new Shopmon advisory.
type CreateInput struct {
	Title                 string
	Packages              []PackageInput
	Severity              *string
	Summary               *string
	Description           *string
	CvssScore             *float64
	CvssVector            *string
	Cwes                  *[]api.AdvisoryCWE
	Link                  *string
	ExternalReferences    *[]string
	Cve                   *string
	GhsaID                *string
	ReportedAt            *time.Time
	IsVisible             *bool
	NotesPublic           *string
	NotesInternal         *string
	RemediationSummary    *string
	RemediationUrl        *string
	RecommendedUpgrade    *string
	ShopwareImpactSummary *string
	AffectedComponents    *[]string
	Tags                  *[]string
	CreatedBy             string
}

// Create inserts a Shopmon-authored advisory with a stable SHOPMON-YYYY-NNNN
// id. The id never changes afterwards — linking the official CVE/GHSA later
// only fills the cve/ghsa_id columns — so matches, suppressions, and alert
// markers recorded against it stay valid forever.
func (s *Service) Create(ctx context.Context, input CreateInput) (AdvisoryView, error) {
	n, err := normalizeCreateInput(input)
	if err != nil {
		return AdvisoryView{}, err
	}
	if s.pool == nil {
		return AdvisoryView{}, errors.New("create advisory: database pool is required")
	}

	number, err := s.queries.NextShopmonAdvisoryNumber(ctx)
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("allocate shopmon advisory id: %w", err)
	}
	advisoryID := fmt.Sprintf("SHOPMON-%d-%04d", time.Now().UTC().Year(), number)
	sourcesJSON, err := json.Marshal([]map[string]string{{"name": "Shopmon", "remoteId": advisoryID}})
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("marshal advisory sources: %w", err)
	}

	// Hand-written details are the disclosure until an official source
	// replaces them after linking; the sync markers stay NULL so the
	// NVD/GitHub queues may overwrite them once identifiers exist.
	return s.insertManualAdvisory(ctx, input, n, manualInsert{
		advisoryID:    advisoryID,
		origin:        originShopmon,
		sourcesJSON:   sourcesJSON,
		detailsSource: detailsSourceShopmon,
	})
}

// ImportGHSA inserts an advisory an admin imported from a GitHub security
// advisory after reviewing the fetched data. It keeps the GHSA id as its
// advisory_id with origin 'github': an official advisory, just not a synced
// one. The detail-sync markers are stamped at insert because the content
// already came from GitHub — the enrichment queues would only overwrite the
// admin's reviewed text with the same source. Linking the CVE later re-opens
// enrichment through the normal link path.
func (s *Service) ImportGHSA(ctx context.Context, input CreateInput) (AdvisoryView, error) {
	n, err := normalizeCreateInput(input)
	if err != nil {
		return AdvisoryView{}, err
	}
	if n.ghsa == nil {
		return AdvisoryView{}, fmt.Errorf("%w: ghsaId is required for a GitHub import", ErrInvalidInput)
	}
	if s.pool == nil {
		return AdvisoryView{}, errors.New("import advisory: database pool is required")
	}

	sourcesJSON, err := json.Marshal([]map[string]string{{"name": "GitHub", "remoteId": *n.ghsa}})
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("marshal advisory sources: %w", err)
	}

	return s.insertManualAdvisory(ctx, input, n, manualInsert{
		advisoryID:    *n.ghsa,
		origin:        originGitHub,
		sourcesJSON:   sourcesJSON,
		detailsSource: detailsSourceGitHub,
		detailsSynced: true,
	})
}

type normalizedCreate struct {
	title       string
	packages    []PackageInput
	severity    *string
	cve         *string
	ghsa        *string
	link        *string
	summary     *string
	description *string
	cwesJSON    json.RawMessage
	refsJSON    json.RawMessage
}

func normalizeCreateInput(input CreateInput) (normalizedCreate, error) {
	var n normalizedCreate
	n.title = strings.TrimSpace(input.Title)
	if n.title == "" {
		return n, fmt.Errorf("%w: title must not be empty", ErrInvalidInput)
	}

	var err error
	if n.packages, err = normalizePackages(input.Packages); err != nil {
		return n, err
	}
	if n.severity, err = normalizeSeverity(input.Severity); err != nil {
		return n, err
	}
	if n.cve, err = normalizeCVE(input.Cve); err != nil {
		return n, err
	}
	if n.ghsa, err = normalizeGHSA(input.GhsaID); err != nil {
		return n, err
	}
	if err := validateCVSSScore(input.CvssScore); err != nil {
		return n, err
	}

	n.link = optTrimmed(input.Link)
	if n.link != nil {
		if err := validateHTTPURL("link", *n.link); err != nil {
			return n, err
		}
	}
	n.summary = patchNilIfEmpty(input.Summary)
	n.description = patchNilIfEmpty(input.Description)

	n.cwesJSON = []byte("[]")
	if input.Cwes != nil {
		if n.cwesJSON, err = encodeCWEs(*input.Cwes); err != nil {
			return n, err
		}
	}
	n.refsJSON = []byte("[]")
	if input.ExternalReferences != nil {
		if n.refsJSON, err = encodeReferences(*input.ExternalReferences); err != nil {
			return n, err
		}
	}
	return n, nil
}

// ensureIdentifiersUnused refuses identifiers another row already claims:
// splitting one vulnerability across two advisories would double-match and
// double-alert shops. The admin links or enriches the existing row instead.
func (s *Service) ensureIdentifiersUnused(ctx context.Context, txq *queries.Queries, cve, ghsa *string) error {
	if cve == nil && ghsa == nil {
		return nil
	}
	existing, err := txq.ListComposerAdvisoriesByIdentifiers(ctx, queries.ListComposerAdvisoriesByIdentifiersParams{
		Cve:    cve,
		GhsaID: ghsa,
	})
	if err != nil {
		return fmt.Errorf("check advisory identifiers: %w", err)
	}
	if len(existing) > 0 {
		return fmt.Errorf("%w: identifier is already tracked as advisory %s", ErrConflict, existing[0].AdvisoryID)
	}
	return nil
}

type manualInsert struct {
	advisoryID    string
	origin        string
	sourcesJSON   json.RawMessage
	detailsSource string
	detailsSynced bool
}

func (s *Service) insertManualAdvisory(ctx context.Context, input CreateInput, n normalizedCreate, ins manualInsert) (AdvisoryView, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("begin advisory create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := s.queries.WithTx(tx)
	if err := txq.LockComposerAdvisoryCatalog(ctx); err != nil {
		return AdvisoryView{}, fmt.Errorf("lock advisory catalog: %w", err)
	}
	if err := s.ensureIdentifiersUnused(ctx, txq, n.cve, n.ghsa); err != nil {
		return AdvisoryView{}, err
	}

	now := time.Now().UTC()
	reportedAt := pgtype.Timestamp{Time: now, Valid: true}
	if input.ReportedAt != nil {
		reportedAt = pgtype.Timestamp{Time: input.ReportedAt.UTC(), Valid: true}
	}

	var detailsSource *string
	if ins.detailsSynced {
		source := ins.detailsSource
		detailsSource = &source
	} else if n.summary != nil || n.description != nil || input.CvssScore != nil {
		source := ins.detailsSource
		detailsSource = &source
	}
	var syncedAt pgtype.Timestamp
	if ins.detailsSynced {
		syncedAt = pgtype.Timestamp{Time: now, Valid: true}
	}

	isVisible := false
	if input.IsVisible != nil {
		isVisible = *input.IsVisible
	}

	if _, err := txq.InsertManualAdvisory(ctx, queries.InsertManualAdvisoryParams{
		AdvisoryID:            ins.advisoryID,
		Title:                 n.title,
		Link:                  n.link,
		Cve:                   n.cve,
		GhsaID:                n.ghsa,
		Severity:              n.severity,
		Sources:               ins.sourcesJSON,
		ReportedAt:            reportedAt,
		Summary:               n.summary,
		Description:           n.description,
		CvssScore:             input.CvssScore,
		CvssVector:            optTrimmed(input.CvssVector),
		Cwes:                  n.cwesJSON,
		ExternalReferences:    n.refsJSON,
		DetailsSource:         detailsSource,
		DetailsSyncedAt:       syncedAt,
		GithubSyncedAt:        syncedAt,
		IsVisible:             isVisible,
		NotesPublic:           patchNilIfEmpty(input.NotesPublic),
		NotesInternal:         patchNilIfEmpty(input.NotesInternal),
		RemediationSummary:    patchNilIfEmpty(input.RemediationSummary),
		RemediationUrl:        optTrimmed(input.RemediationUrl),
		RecommendedUpgrade:    optTrimmed(input.RecommendedUpgrade),
		ShopwareImpactSummary: patchNilIfEmpty(input.ShopwareImpactSummary),
		AffectedComponents:    derefOrEmpty(input.AffectedComponents),
		Tags:                  derefOrEmpty(input.Tags),
		Origin:                ins.origin,
		CreatedBy:             &input.CreatedBy,
	}); err != nil {
		return AdvisoryView{}, fmt.Errorf("insert manual advisory: %w", err)
	}

	for _, pkg := range n.packages {
		if err := txq.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
			AdvisoryID:          ins.advisoryID,
			PackageName:         pkg.PackageName,
			PackagistAdvisoryID: nil,
			AffectedVersions:    pkg.AffectedVersions,
		}); err != nil {
			return AdvisoryView{}, fmt.Errorf("insert manual advisory package: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return AdvisoryView{}, fmt.Errorf("commit advisory create transaction: %w", err)
	}
	return s.Get(ctx, ins.advisoryID, false)
}

// LinkOfficialInput carries the official identifiers a Shopmon advisory is
// linked to once the registry publishes them.
type LinkOfficialInput struct {
	Cve    *string
	GhsaID *string
	Link   *string
}

// LinkOfficial records the official CVE/GHSA on a Shopmon-authored advisory.
//
// Two cases: the official advisory is not in the catalog yet, in which case
// the identifiers are simply stored (the next Packagist sync merges upstream
// data into this row instead of creating a canonical CVE row). Or it was
// already synced as its own row, in which case that row is merged into the
// Shopmon row here — upstream content and package ranges win, while matches,
// notification markers, and suppressions move to the survivor so affected
// shops are neither double-listed nor re-alerted.
func (s *Service) LinkOfficial(ctx context.Context, advisoryID string, input LinkOfficialInput) (AdvisoryView, error) {
	cve, err := normalizeCVE(input.Cve)
	if err != nil {
		return AdvisoryView{}, err
	}
	ghsa, err := normalizeGHSA(input.GhsaID)
	if err != nil {
		return AdvisoryView{}, err
	}
	if cve == nil && ghsa == nil {
		return AdvisoryView{}, fmt.Errorf("%w: cve or ghsaId is required", ErrInvalidInput)
	}
	link := optTrimmed(input.Link)
	if link != nil {
		if err := validateHTTPURL("link", *link); err != nil {
			return AdvisoryView{}, err
		}
	}

	view, err := s.Get(ctx, advisoryID, false)
	if err != nil {
		return AdvisoryView{}, err
	}
	advisoryID = view.Advisory.AdvisoryID

	if s.pool == nil {
		return AdvisoryView{}, errors.New("link advisory: database pool is required")
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("begin advisory link transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := s.queries.WithTx(tx)
	if err := txq.LockComposerAdvisoryCatalog(ctx); err != nil {
		return AdvisoryView{}, fmt.Errorf("lock advisory catalog: %w", err)
	}

	current, err := txq.GetComposerAdvisoryForUpdate(ctx, advisoryID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return AdvisoryView{}, ErrNotFound
		}
		return AdvisoryView{}, fmt.Errorf("lock advisory: %w", err)
	}
	if current.Origin != originShopmon && current.Origin != originGitHub {
		return AdvisoryView{}, fmt.Errorf("%w: only manually tracked advisories (Shopmon-authored or GitHub-imported) can be linked to an official advisory", ErrInvalidInput)
	}

	candidates, err := txq.ListComposerAdvisoriesByIdentifiers(ctx, queries.ListComposerAdvisoriesByIdentifiersParams{
		Cve:    cve,
		GhsaID: ghsa,
	})
	if err != nil {
		return AdvisoryView{}, fmt.Errorf("list advisories by identifiers: %w", err)
	}
	others := candidates[:0]
	for _, candidate := range candidates {
		if candidate.AdvisoryID != current.AdvisoryID {
			others = append(others, candidate)
		}
	}
	if len(others) > 1 {
		return AdvisoryView{}, fmt.Errorf("%w: identifiers match multiple advisories (%s, %s)", ErrConflict, others[0].AdvisoryID, others[1].AdvisoryID)
	}

	// Once upstream data has been merged in (synced details or Packagist
	// package rows), the recorded identifiers are the identity of record;
	// silently swapping them would orphan the sync's view of this row.
	// Adding a still-missing identifier kind (the CVE for a GHSA-only row,
	// the normal lifecycle of a GitHub import) and re-linking the same
	// identifiers stay allowed so the operation is idempotent for the admin.
	currentPackages, err := s.listPackagesTx(ctx, txq, current.AdvisoryID)
	if err != nil {
		return AdvisoryView{}, err
	}
	alreadyMerged := current.DetailsSyncedAt.Valid
	for _, pkg := range currentPackages {
		if pkg.PackagistAdvisoryID != nil && *pkg.PackagistAdvisoryID != "" {
			alreadyMerged = true
		}
	}
	if alreadyMerged {
		if cve != nil && nonEmptyPtr(current.Cve) && !equalStringPtr(cve, current.Cve) ||
			ghsa != nil && nonEmptyPtr(current.GhsaID) && !equalStringPtr(ghsa, current.GhsaID) {
			return AdvisoryView{}, fmt.Errorf("%w: advisory %s is already linked and merged with upstream data", ErrConflict, current.AdvisoryID)
		}
	}

	content := contentParamsFromRow(current)
	if len(others) == 1 {
		official := others[0]
		if official.Origin == originShopmon || official.Origin == originGitHub {
			return AdvisoryView{}, fmt.Errorf("%w: identifier is already linked to manually tracked advisory %s", ErrConflict, official.AdvisoryID)
		}
		official, err = txq.GetComposerAdvisoryForUpdate(ctx, official.AdvisoryID)
		if err != nil {
			return AdvisoryView{}, fmt.Errorf("lock official advisory: %w", err)
		}
		content = mergedContent(current, official, cve, ghsa, link)
		if err := s.mergeOfficialRow(ctx, txq, current, official); err != nil {
			return AdvisoryView{}, err
		}
	} else {
		if cve != nil {
			content.Cve = cve
		}
		if ghsa != nil {
			content.GhsaID = ghsa
		}
		if link != nil {
			content.Link = link
		}
		// Re-open official enrichment: the NVD/GitHub queues now have
		// identifiers to fetch, and their details replace the Shopmon text.
		content.DetailsSyncedAt = pgtype.Timestamp{}
		content.GithubSyncedAt = pgtype.Timestamp{}
	}

	if _, err := txq.UpdateComposerAdvisoryContent(ctx, content); err != nil {
		return AdvisoryView{}, fmt.Errorf("update linked advisory content: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return AdvisoryView{}, fmt.Errorf("commit advisory link transaction: %w", err)
	}
	return s.Get(ctx, advisoryID, false)
}

// mergeOfficialRow moves everything recorded against the official row onto
// the surviving Shopmon row, adopts the official package set, and deletes the
// official row. Runs inside the link transaction.
func (s *Service) mergeOfficialRow(ctx context.Context, txq *queries.Queries, survivor, official queries.ComposerAdvisory) error {
	// The official package set replaces the survivor's: delete its rows, then
	// re-point the official rows (copying them would collide on the unique
	// packagist_advisory_id while both rows exist).
	if err := txq.DeleteComposerAdvisoryPackages(ctx, survivor.AdvisoryID); err != nil {
		return fmt.Errorf("replace advisory packages on merge: %w", err)
	}
	if err := txq.MergeComposerAdvisoryPackages(ctx, queries.MergeComposerAdvisoryPackagesParams{
		AdvisoryID:   survivor.AdvisoryID,
		AdvisoryID_2: official.AdvisoryID,
	}); err != nil {
		return fmt.Errorf("adopt official advisory packages: %w", err)
	}

	if err := txq.MergeEnvironmentAdvisoryMatches(ctx, queries.MergeEnvironmentAdvisoryMatchesParams{
		AdvisoryID:   survivor.AdvisoryID,
		AdvisoryID_2: official.AdvisoryID,
	}); err != nil {
		return fmt.Errorf("merge advisory matches: %w", err)
	}
	if err := txq.MergeEnvironmentAdvisoryNotifications(ctx, queries.MergeEnvironmentAdvisoryNotificationsParams{
		AdvisoryID:   survivor.AdvisoryID,
		AdvisoryID_2: official.AdvisoryID,
	}); err != nil {
		return fmt.Errorf("merge advisory notification markers: %w", err)
	}
	if err := txq.RevokeExpiredSuppressionsForAdvisoryMerge(ctx, []string{survivor.AdvisoryID, official.AdvisoryID}); err != nil {
		return fmt.Errorf("retire expired advisory suppressions: %w", err)
	}
	if err := txq.MergeRevokedAdvisorySuppressions(ctx, queries.MergeRevokedAdvisorySuppressionsParams{
		AdvisoryID:   survivor.AdvisoryID,
		AdvisoryID_2: official.AdvisoryID,
	}); err != nil {
		return fmt.Errorf("merge revoked advisory suppressions: %w", err)
	}
	if err := txq.MergeActiveAdvisorySuppressions(ctx, queries.MergeActiveAdvisorySuppressionsParams{
		AdvisoryID:   survivor.AdvisoryID,
		AdvisoryID_2: official.AdvisoryID,
	}); err != nil {
		return fmt.Errorf("merge active advisory suppressions: %w", err)
	}

	// Anything still pointing at the official row (duplicate matches or
	// suppressions the survivor already has) goes away with the cascade.
	if err := txq.DeleteComposerAdvisory(ctx, official.AdvisoryID); err != nil {
		return fmt.Errorf("delete merged official advisory: %w", err)
	}
	return nil
}

// mergedContent computes the surviving row's content after a merge: upstream
// is authoritative for everything the sync maintains, the admin-supplied link
// identifiers fill gaps, and Shopmon values remain only as fallback.
func mergedContent(current, official queries.ComposerAdvisory, cve, ghsa, link *string) queries.UpdateComposerAdvisoryContentParams {
	content := contentParamsFromRow(current)
	content.Title = official.Title
	content.Link = firstNonEmptyPtr(official.Link, link, current.Link)
	content.Cve = firstNonEmptyPtr(official.Cve, cve, current.Cve)
	content.GhsaID = firstNonEmptyPtr(official.GhsaID, ghsa, current.GhsaID)
	content.Severity = firstNonEmptyPtr(official.Severity, current.Severity)
	if len(official.Sources) > 0 && string(official.Sources) != "[]" {
		content.Sources = official.Sources
	}
	if official.ReportedAt.Valid {
		content.ReportedAt = official.ReportedAt
	}
	content.ComposerRepository = firstNonEmptyPtr(official.ComposerRepository, current.ComposerRepository)
	if official.DetailsSyncedAt.Valid {
		content.Summary = official.Summary
		content.Description = official.Description
		content.CvssScore = official.CvssScore
		content.CvssVector = official.CvssVector
		content.Cwes = official.Cwes
		content.ExternalReferences = official.ExternalReferences
		content.DetailsSource = official.DetailsSource
		content.DetailsSyncedAt = official.DetailsSyncedAt
		content.GithubSyncedAt = official.GithubSyncedAt
	} else {
		content.DetailsSyncedAt = pgtype.Timestamp{}
		content.GithubSyncedAt = pgtype.Timestamp{}
	}
	if len(official.FirstPatchedVersions) > 0 && string(official.FirstPatchedVersions) != "{}" {
		content.FirstPatchedVersions = official.FirstPatchedVersions
	}
	return content
}

func contentParamsFromRow(row queries.ComposerAdvisory) queries.UpdateComposerAdvisoryContentParams {
	return queries.UpdateComposerAdvisoryContentParams{
		AdvisoryID:           row.AdvisoryID,
		Title:                row.Title,
		Link:                 row.Link,
		Cve:                  row.Cve,
		GhsaID:               row.GhsaID,
		Severity:             row.Severity,
		Sources:              nonEmptyJSON(row.Sources, "[]"),
		ReportedAt:           row.ReportedAt,
		ComposerRepository:   row.ComposerRepository,
		Summary:              row.Summary,
		Description:          row.Description,
		CvssScore:            row.CvssScore,
		CvssVector:           row.CvssVector,
		Cwes:                 nonEmptyJSON(row.Cwes, "[]"),
		ExternalReferences:   nonEmptyJSON(row.ExternalReferences, "[]"),
		DetailsSource:        row.DetailsSource,
		DetailsSyncedAt:      row.DetailsSyncedAt,
		GithubSyncedAt:       row.GithubSyncedAt,
		FirstPatchedVersions: nonEmptyJSON(row.FirstPatchedVersions, "{}"),
	}
}

func (s *Service) listPackagesTx(ctx context.Context, txq *queries.Queries, advisoryID string) ([]queries.ComposerAdvisoryPackage, error) {
	rows, err := txq.ListComposerAdvisoryPackagesForAdvisory(ctx, advisoryID)
	if err != nil {
		return nil, fmt.Errorf("list advisory packages: %w", err)
	}
	packages := make([]queries.ComposerAdvisoryPackage, 0, len(rows))
	for _, p := range rows {
		packages = append(packages, queries.ComposerAdvisoryPackage{
			AdvisoryID:          advisoryID,
			PackageName:         p.PackageName,
			PackagistAdvisoryID: p.PackagistAdvisoryID,
			AffectedVersions:    p.AffectedVersions,
			SyncedAt:            p.SyncedAt,
		})
	}
	return packages, nil
}

// contentLocked reports whether the sync owns an advisory's core fields.
// The CVE is the full official identity: once one is linked, upstream data
// wins. A GHSA-only row (every GitHub import, Shopmon advisories linked
// early) stays editable, because the registry entry an admin must correct
// or complete is exactly what exists so far.
func contentLocked(row queries.ComposerAdvisory) bool {
	return nonEmptyPtr(row.Cve)
}

func nonEmptyPtr(v *string) bool {
	return v != nil && strings.TrimSpace(*v) != ""
}

func normalizeSeverity(v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	trimmed := strings.ToLower(strings.TrimSpace(*v))
	if trimmed == "" {
		return nil, nil
	}
	if !api.SeverityLevel(trimmed).Valid() {
		return nil, fmt.Errorf("%w: severity %q", ErrInvalidInput, trimmed)
	}
	return &trimmed, nil
}

func normalizeCVE(v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	trimmed := strings.ToUpper(strings.TrimSpace(*v))
	if trimmed == "" {
		return nil, nil
	}
	if !cvePattern.MatchString(trimmed) {
		return nil, fmt.Errorf("%w: invalid CVE id %q", ErrInvalidInput, trimmed)
	}
	return &trimmed, nil
}

// ghsaInURLPattern finds a GHSA id inside a GitHub advisory URL, so admins
// can paste the link they have open instead of extracting the id by hand.
var ghsaInURLPattern = regexp.MustCompile(`(?i)GHSA-[a-z0-9]{4}-[a-z0-9]{4}-[a-z0-9]{4}`)

func normalizeGHSA(v *string) (*string, error) {
	if v == nil {
		return nil, nil
	}
	trimmed := strings.TrimSpace(*v)
	if trimmed == "" {
		return nil, nil
	}
	if strings.Contains(trimmed, "/") {
		if found := ghsaInURLPattern.FindString(trimmed); found != "" {
			trimmed = found
		}
	}
	if !ghsaPattern.MatchString(trimmed) {
		return nil, fmt.Errorf("%w: invalid GHSA id %q", ErrInvalidInput, trimmed)
	}
	// Canonical GHSA spelling: uppercase prefix, lowercase segments. Upstream
	// sources use it too, so identifier lookups compare equal byte-for-byte.
	canonical := "GHSA-" + strings.ToLower(trimmed[len("GHSA-"):])
	return &canonical, nil
}

func validateCVSSScore(v *float64) error {
	if v == nil {
		return nil
	}
	if *v < 0 || *v > 10 {
		return fmt.Errorf("%w: cvssScore %v is outside 0–10", ErrInvalidInput, *v)
	}
	return nil
}

func validateHTTPURL(field, raw string) error {
	parsed, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Host == "" {
		return fmt.Errorf("%w: %s %q is not an http(s) URL", ErrInvalidInput, field, raw)
	}
	return nil
}

func normalizePackages(packages []PackageInput) ([]PackageInput, error) {
	if len(packages) == 0 {
		return nil, fmt.Errorf("%w: at least one affected package is required", ErrInvalidInput)
	}
	seen := make(map[string]bool, len(packages))
	out := make([]PackageInput, 0, len(packages))
	for _, pkg := range packages {
		name := strings.ToLower(strings.TrimSpace(pkg.PackageName))
		if !packageNamePattern.MatchString(name) {
			return nil, fmt.Errorf("%w: invalid Composer package name %q", ErrInvalidInput, pkg.PackageName)
		}
		if seen[name] {
			return nil, fmt.Errorf("%w: duplicate package %q", ErrInvalidInput, name)
		}
		seen[name] = true
		constraint := strings.TrimSpace(pkg.AffectedVersions)
		if err := sbom.ValidateAffectedVersions(constraint); err != nil {
			return nil, fmt.Errorf("%w: package %s: %v", ErrInvalidInput, name, err)
		}
		out = append(out, PackageInput{PackageName: name, AffectedVersions: constraint})
	}
	return out, nil
}

// encodeCWEs writes the {cwe_id, name} shape the NVD/GitHub enrichment
// produces, so manually authored CWEs round-trip through the same parser.
func encodeCWEs(cwes []api.AdvisoryCWE) (json.RawMessage, error) {
	type storedCWE struct {
		ID   string `json:"cwe_id"`
		Name string `json:"name"`
	}
	stored := make([]storedCWE, 0, len(cwes))
	for _, cwe := range cwes {
		id := strings.ToUpper(strings.TrimSpace(cwe.Id))
		if !cwePattern.MatchString(id) {
			return nil, fmt.Errorf("%w: invalid CWE id %q", ErrInvalidInput, cwe.Id)
		}
		stored = append(stored, storedCWE{ID: id, Name: strings.TrimSpace(cwe.Name)})
	}
	encoded, err := json.Marshal(stored)
	if err != nil {
		return nil, fmt.Errorf("marshal cwes: %w", err)
	}
	return encoded, nil
}

func encodeReferences(references []string) (json.RawMessage, error) {
	out := make([]string, 0, len(references))
	for _, ref := range references {
		ref = strings.TrimSpace(ref)
		if ref == "" {
			continue
		}
		if err := validateHTTPURL("externalReferences", ref); err != nil {
			return nil, err
		}
		out = append(out, ref)
	}
	encoded, err := json.Marshal(out)
	if err != nil {
		return nil, fmt.Errorf("marshal external references: %w", err)
	}
	return encoded, nil
}

func optTrimmed(v *string) *string {
	if v == nil {
		return nil
	}
	return nilIfEmpty(strings.TrimSpace(*v))
}

func patchNilIfEmpty(v *string) *string {
	if v == nil {
		return nil
	}
	return nilIfEmpty(*v)
}

func derefOrEmpty(v *[]string) []string {
	if v == nil {
		return []string{}
	}
	return *v
}

func firstNonEmptyPtr(values ...*string) *string {
	for _, v := range values {
		if v != nil && strings.TrimSpace(*v) != "" {
			return v
		}
	}
	return nil
}

func equalStringPtr(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func nonEmptyJSON(raw json.RawMessage, fallback string) json.RawMessage {
	if len(raw) == 0 || string(raw) == "null" {
		return json.RawMessage(fallback)
	}
	return raw
}
