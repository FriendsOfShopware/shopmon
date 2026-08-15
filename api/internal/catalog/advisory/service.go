package advisory

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/shyim/go-composer/repository"
)

const (
	vendorShopware     = "shopware"
	packagePrefix      = "shopware/"
	defaultHTTPTimeout = 30 * time.Second
	// Cap GitHub fallback enrichment per sync (NVD has its own batch/delay).
	githubDetailsBatchSize = 20
	githubDetailsDelay     = 120 * time.Millisecond
)

var tracer = otel.Tracer("shopmon/catalog/advisory")

// Service crawls Packagist security advisories for official shopware/* packages
// and upserts them into composer_advisory. Admin enrichment columns are never
// written by this service. After Packagist upsert it fills CVE description /
// CVSS / CWE details from the public NVD API (preferred), falling back to the
// GitHub Advisories API when only a GHSA is available.
type Service struct {
	// pool is needed alongside queries because the rematch pass replaces an
	// environment's match rows transactionally.
	pool    *pgxpool.Pool
	queries *queries.Queries
	client  *packagistClient
	nvd     *nvdClient
	github  *githubClient
	// notifier alerts environment subscribers about newly matching advisories.
	// Nil in tests that do not exercise notifications.
	notifier *notify.Dispatcher
	now      func() time.Time
}

type Config struct {
	BaseURL     string
	UserAgent   string
	GithubToken string
	NVDAPIKey   string
	// Mailer delivers advisory alert emails. Nil disables the email channel.
	Mailer mail.Sender
}

func NewService(pool *pgxpool.Pool, q *queries.Queries, cfg Config) *Service {
	httpClient := httputil.NewHTTPClient(httputil.WithTimeout(defaultHTTPTimeout))
	return &Service{
		pool:     pool,
		queries:  q,
		client:   newPackagistClient(cfg.BaseURL, cfg.UserAgent, httpClient),
		nvd:      newNVDClient(cfg.NVDAPIKey, cfg.UserAgent, httpClient),
		github:   newGitHubClient(cfg.GithubToken, cfg.UserAgent, httpClient),
		notifier: notify.NewDispatcher(q, cfg.Mailer),
		now:      time.Now,
	}
}

// NewServiceWithClient is for tests that need a custom HTTP client / base URL.
// The same base URL is used for Packagist, NVD, and GitHub mocks.
func NewServiceWithClient(pool *pgxpool.Pool, q *queries.Queries, baseURL string, httpClient *http.Client) *Service {
	base := strings.TrimRight(baseURL, "/")
	nvd := newNVDClient("", defaultUserAgent, httpClient)
	nvd.baseURL = base + "/rest/json/cves/2.0"
	gh := newGitHubClient("", defaultUserAgent, httpClient)
	gh.baseURL = base
	return &Service{
		pool:    pool,
		queries: q,
		client:  newPackagistClient(baseURL, defaultUserAgent, httpClient),
		nvd:     nvd,
		github:  gh,
		now:     time.Now,
	}
}

func (s *Service) Sync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "shopware.advisory.sync")
	defer span.End()

	state, err := s.queries.GetComposerAdvisorySyncState(ctx)
	if err != nil {
		// First deploy before migration seed is unlikely; still treat as empty state.
		state = queries.ComposerAdvisorySyncState{}
	}

	now := s.now().UTC()

	// Always re-query shopware/* packages via go-composer. There are few vendor
	// packages, so a full package-scoped Packagist advisory pull is cheap and
	// avoids the brittle updatedSince response shape.
	upserted, err := s.syncFromPackagist(ctx)
	if err != nil {
		_ = s.queries.UpsertComposerAdvisorySyncState(ctx, queries.UpsertComposerAdvisorySyncStateParams{
			LastUpdatedSince:      state.LastUpdatedSince,
			LastFullSyncAt:        state.LastFullSyncAt,
			LastIncrementalSyncAt: state.LastIncrementalSyncAt,
			LastError:             ptrString(err.Error()),
		})
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	enriched, enrichErr := s.enrichMissingDetails(ctx)
	if enrichErr != nil {
		// Packagist data is already durable; log enrichment failure without
		// failing the whole job so timestamps still advance.
		slog.WarnContext(ctx, "advisory detail enrichment incomplete", "error", enrichErr, "enriched", enriched)
		span.RecordError(enrichErr)
	}

	// Re-evaluate known shop inventories against the refreshed catalog so newly
	// published advisories surface without waiting for each shop's next scrape.
	rematched, rematchErr := s.rematchEnvironments(ctx)
	if rematchErr != nil {
		slog.WarnContext(ctx, "advisory rematch incomplete", "error", rematchErr)
		span.RecordError(rematchErr)
	}

	nowUnix := now.Unix()
	if err := s.queries.UpsertComposerAdvisorySyncState(ctx, queries.UpsertComposerAdvisorySyncStateParams{
		LastUpdatedSince: &nowUnix,
		LastFullSyncAt:   pgtype.Timestamp{Time: now, Valid: true},
		LastError:        nil,
	}); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("update advisory sync state: %w", err)
	}

	span.SetAttributes(
		attribute.Int("advisories.upserted", upserted),
		attribute.Int("advisories.details_enriched", enriched),
		attribute.Int("advisories.environment_matches", rematched),
	)
	slog.InfoContext(ctx, "synced composer advisories",
		"upserted", upserted, "detailsEnriched", enriched, "environmentMatches", rematched)
	return nil
}

func (s *Service) syncFromPackagist(ctx context.Context) (int, error) {
	packages, err := s.client.listVendorPackages(ctx, vendorShopware)
	if err != nil {
		return 0, fmt.Errorf("list shopware packages: %w", err)
	}

	wanted := filterShopwarePackages(packages)

	// Widen coverage to every Composer package actually installed across the
	// monitored shops (collected from their FroshTools SBOMs). Most real
	// vulnerabilities live in transitive dependencies like symfony/* or twig/*,
	// which the shopware/* vendor listing alone never surfaces.
	//
	// Only packages Packagist already publishes are included. A monitored shop
	// may run in-house packages, and their names are the tenant's information:
	// sending "acme/secret-billing" to a third party would disclose it even
	// though no advisory could ever come back. publicPackagesOf establishes
	// publicness from vendor listings, so an unpublished package name never
	// leaves Shopmon.
	installed, err := s.queries.ListDistinctSbomPackageNames(ctx)
	if err != nil {
		// The SBOM inventory is an enhancement; a failure here must not stop
		// the shopware/* sync that shops already depend on.
		slog.WarnContext(ctx, "failed to list sbom packages for advisory sync", "error", err)
	} else {
		public := s.client.publicPackagesOf(ctx, installed)
		if skipped := len(installed) - len(public); skipped > 0 {
			slog.InfoContext(ctx, "excluded non-public packages from advisory sync",
				"skipped", skipped, "checked", len(installed))
		}
		wanted = mergePackageNames(wanted, public)
	}

	advisoriesByPkg, err := s.client.advisoriesForPackages(ctx, wanted)
	if err != nil {
		return 0, fmt.Errorf("fetch package advisories: %w", err)
	}
	return s.upsertAll(ctx, advisoriesByPkg)
}

func (s *Service) upsertAll(ctx context.Context, advisoriesByPkg repository.Advisories) (int, error) {
	var upserted int
	for pkg, ads := range advisoriesByPkg {
		if pkg == "" {
			continue
		}
		for _, adv := range ads {
			if err := s.upsertOne(ctx, adv); err != nil {
				return upserted, err
			}
			upserted++
		}
	}
	return upserted, nil
}

func (s *Service) upsertOne(ctx context.Context, adv repository.SecurityAdvisory) error {
	if adv.AdvisoryID == "" || adv.PackageName == "" || adv.Title == "" {
		slog.WarnContext(ctx, "skipping incomplete packagist advisory",
			"advisoryId", adv.AdvisoryID, "package", adv.PackageName)
		return nil
	}
	sourcesJSON, err := json.Marshal(adv.Sources)
	if err != nil {
		return fmt.Errorf("marshal sources for %s: %w", adv.AdvisoryID, err)
	}
	if len(adv.Sources) == 0 {
		sourcesJSON = []byte("[]")
	}

	var reportedAt pgtype.Timestamp
	if t, ok := parseReportedAt(adv.ReportedAt); ok {
		reportedAt = pgtype.Timestamp{Time: t, Valid: true}
	}

	ghsa := extractGHSAID(adv)
	cve := strings.ToUpper(strings.TrimSpace(adv.CVE))
	var cvePtr *string
	if cve != "" {
		cvePtr = &cve
	}
	var ghsaPtr *string
	if ghsa != "" {
		ghsaPtr = &ghsa
	}
	var link *string
	if strings.TrimSpace(adv.Link) != "" {
		link = ptrString(strings.TrimSpace(adv.Link))
	}
	var severity *string
	if strings.TrimSpace(adv.Severity) != "" {
		severity = ptrString(strings.ToLower(strings.TrimSpace(adv.Severity)))
	}
	var repo *string
	if strings.TrimSpace(adv.ComposerRepository) != "" {
		repo = ptrString(strings.TrimSpace(adv.ComposerRepository))
	}

	affected := adv.AffectedVersions
	if affected == "" {
		affected = "*"
	}

	// One logical row per CVE/GHSA; package ranges live in the child table.
	canonicalID := canonicalAdvisoryID(cve, ghsa, adv.AdvisoryID)

	if err := s.queries.UpsertComposerAdvisory(ctx, queries.UpsertComposerAdvisoryParams{
		AdvisoryID:         canonicalID,
		Title:              adv.Title,
		Link:               link,
		Cve:                cvePtr,
		GhsaID:             ghsaPtr,
		Severity:           severity,
		Sources:            sourcesJSON,
		ReportedAt:         reportedAt,
		ComposerRepository: repo,
	}); err != nil {
		return fmt.Errorf("upsert advisory %s: %w", canonicalID, err)
	}

	if err := s.queries.UpsertComposerAdvisoryPackage(ctx, queries.UpsertComposerAdvisoryPackageParams{
		AdvisoryID:          canonicalID,
		PackageName:         adv.PackageName,
		PackagistAdvisoryID: adv.AdvisoryID,
		AffectedVersions:    affected,
	}); err != nil {
		return fmt.Errorf("upsert advisory package %s/%s: %w", canonicalID, adv.PackageName, err)
	}
	return nil
}

// canonicalAdvisoryID collapses multi-package Packagist rows for the same
// vulnerability onto one primary key (CVE preferred, then GHSA, then PKSA).
func canonicalAdvisoryID(cve, ghsa, packagistID string) string {
	if cve = strings.ToUpper(strings.TrimSpace(cve)); cve != "" {
		return cve
	}
	if ghsa = strings.TrimSpace(ghsa); ghsa != "" {
		return ghsa
	}
	return strings.TrimSpace(packagistID)
}

func (s *Service) enrichMissingDetails(ctx context.Context) (int, error) {
	ctx, span := tracer.Start(ctx, "shopware.advisory.enrich_details")
	defer span.End()

	var total int
	var firstErr error

	// Prefer NVD for official CVE descriptions when a CVE id is present.
	if s.nvd != nil {
		n, err := s.enrichFromNVD(ctx)
		total += n
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// GitHub fills rows that only have a GHSA, or still lack details after NVD.
	if s.github != nil {
		n, err := s.enrichFromGitHub(ctx)
		total += n
		if err != nil && firstErr == nil {
			firstErr = err
		}
	}

	span.SetAttributes(attribute.Int("advisories.details_enriched", total))
	return total, firstErr
}

func (s *Service) enrichFromNVD(ctx context.Context) (int, error) {
	ids, err := s.queries.ListComposerAdvisoryCvePendingDetails(ctx, s.nvd.batchSize())
	if err != nil {
		return 0, fmt.Errorf("list cve pending details: %w", err)
	}

	var enriched int
	var firstErr error
	delay := s.nvd.fetchDelay()
	for i, rawID := range ids {
		if rawID == nil || strings.TrimSpace(*rawID) == "" {
			continue
		}
		cveID := strings.TrimSpace(*rawID)
		if i > 0 {
			select {
			case <-ctx.Done():
				return enriched, ctx.Err()
			case <-time.After(delay):
			}
		}

		// Stamp the attempt regardless of outcome so an identifier that keeps
		// failing drifts to the back of the queue instead of starving it.
		if err := s.queries.MarkComposerAdvisoryDetailsAttemptedByCve(ctx, &cveID); err != nil {
			slog.WarnContext(ctx, "failed to mark cve details attempt", "cve", cveID, "error", err)
		}

		details, err := s.nvd.fetchCVE(ctx, cveID)
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch nvd cve details", "cve", cveID, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			if strings.Contains(err.Error(), "rate limited") || strings.Contains(err.Error(), "forbidden") {
				break
			}
			continue
		}

		if err := s.applyDetailsByCVE(ctx, cveID, details, "nvd"); err != nil {
			return enriched, err
		}
		enriched++
	}
	return enriched, firstErr
}

func (s *Service) enrichFromGitHub(ctx context.Context) (int, error) {
	rows, err := s.queries.ListComposerAdvisoryGhsaPendingDetails(ctx, githubDetailsBatchSize)
	if err != nil {
		return 0, fmt.Errorf("list ghsa pending details: %w", err)
	}

	var enriched int
	var firstErr error
	for i, row := range rows {
		if row.GhsaID == nil || strings.TrimSpace(*row.GhsaID) == "" {
			continue
		}
		ghsaID := strings.TrimSpace(*row.GhsaID)
		if i > 0 {
			select {
			case <-ctx.Done():
				return enriched, ctx.Err()
			case <-time.After(githubDetailsDelay):
			}
		}

		// Stamp the attempt regardless of outcome so an identifier that keeps
		// failing drifts to the back of the queue instead of starving it.
		if err := s.queries.MarkComposerAdvisoryDetailsAttemptedByGhsa(ctx, &ghsaID); err != nil {
			slog.WarnContext(ctx, "failed to mark ghsa details attempt", "ghsa", ghsaID, "error", err)
		}

		details, err := s.github.fetchAdvisory(ctx, ghsaID)
		if err != nil {
			slog.WarnContext(ctx, "failed to fetch github advisory details", "ghsa", ghsaID, "error", err)
			if firstErr == nil {
				firstErr = err
			}
			if strings.Contains(err.Error(), "rate limited") || strings.Contains(err.Error(), "forbidden") {
				break
			}
			continue
		}

		// Rows NVD already enriched keep their details; GitHub still owns the
		// patched-versions data, which is why they re-enter this queue at all.
		if err := s.applyDetailsByGHSA(ctx, ghsaID, details, "github", row.NeedsDetails); err != nil {
			return enriched, err
		}
		enriched++
	}
	return enriched, firstErr
}

func (s *Service) applyDetailsByCVE(ctx context.Context, cveID string, details *githubAdvisoryDetails, source string) error {
	cwesJSON, refsJSON, summary, description, err := encodeDetails(details)
	if err != nil {
		return err
	}
	id := cveID
	if err := s.queries.UpdateComposerAdvisoryDetailsByCve(ctx, queries.UpdateComposerAdvisoryDetailsByCveParams{
		Cve:                &id,
		Summary:            summary,
		Description:        description,
		CvssScore:          details.CVSSScore,
		CvssVector:         details.CVSSVector,
		Cwes:               cwesJSON,
		ExternalReferences: refsJSON,
		DetailsSource:      ptrString(source),
	}); err != nil {
		return fmt.Errorf("update details for %s: %w", cveID, err)
	}
	return nil
}

func (s *Service) applyDetailsByGHSA(ctx context.Context, ghsaID string, details *githubAdvisoryDetails, source string, needsDetails bool) error {
	id := ghsaID
	if needsDetails {
		cwesJSON, refsJSON, summary, description, err := encodeDetails(details)
		if err != nil {
			return err
		}
		if err := s.queries.UpdateComposerAdvisoryDetailsByGhsa(ctx, queries.UpdateComposerAdvisoryDetailsByGhsaParams{
			GhsaID:             &id,
			Summary:            summary,
			Description:        description,
			CvssScore:          details.CVSSScore,
			CvssVector:         details.CVSSVector,
			Cwes:               cwesJSON,
			ExternalReferences: refsJSON,
			DetailsSource:      ptrString(source),
		}); err != nil {
			return fmt.Errorf("update details for %s: %w", ghsaID, err)
		}
	}

	// Patched versions come only from GitHub, so they are written separately
	// from the shared details update that the NVD path also uses.
	if len(details.FirstPatchedVersions) > 0 {
		patchedJSON, err := json.Marshal(details.FirstPatchedVersions)
		if err != nil {
			return fmt.Errorf("marshal first patched versions for %s: %w", ghsaID, err)
		}
		if err := s.queries.UpdateComposerAdvisoryFirstPatchedByGhsa(ctx, queries.UpdateComposerAdvisoryFirstPatchedByGhsaParams{
			GhsaID:               &id,
			FirstPatchedVersions: patchedJSON,
		}); err != nil {
			return fmt.Errorf("update first patched versions for %s: %w", ghsaID, err)
		}
	}

	// Marked last so a failure above leaves the row queued for a retry.
	if err := s.queries.MarkComposerAdvisoryGithubSyncedByGhsa(ctx, &id); err != nil {
		return fmt.Errorf("mark github synced for %s: %w", ghsaID, err)
	}
	return nil
}

func encodeDetails(details *githubAdvisoryDetails) (cwesJSON, refsJSON json.RawMessage, summary, description *string, err error) {
	cwesJSON, err = json.Marshal(details.CWEs)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("marshal cwes: %w", err)
	}
	refsJSON, err = json.Marshal(details.References)
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("marshal references: %w", err)
	}
	if len(details.CWEs) == 0 {
		cwesJSON = []byte("[]")
	}
	if len(details.References) == 0 {
		refsJSON = []byte("[]")
	}
	if details.Summary != "" {
		summary = &details.Summary
	}
	if details.Description != "" {
		description = &details.Description
	}
	return cwesJSON, refsJSON, summary, description, nil
}

func filterShopwarePackages(packages []string) []string {
	out := make([]string, 0, len(packages))
	for _, pkg := range packages {
		if strings.HasPrefix(pkg, packagePrefix) {
			out = append(out, pkg)
		}
	}
	return out
}

// mergePackageNames unions the vendor listing with the packages observed in
// shop SBOMs, deduplicating case-insensitively (Composer package names are
// case-insensitive) while preserving a stable order for batching.
func mergePackageNames(base, extra []string) []string {
	seen := make(map[string]bool, len(base)+len(extra))
	out := make([]string, 0, len(base)+len(extra))

	add := func(names []string) {
		for _, name := range names {
			name = strings.TrimSpace(name)
			if name == "" {
				continue
			}
			key := strings.ToLower(name)
			if seen[key] {
				continue
			}
			seen[key] = true
			out = append(out, name)
		}
	}

	add(base)
	add(extra)
	return out
}

func ptrString(s string) *string {
	return &s
}
