package scrape

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/jackc/pgx/v5/pgtype"
	"go.opentelemetry.io/otel/attribute"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// froshToolsExtension is the technical name of the plugin that serves the SBOM
// endpoint. Composer package names are only available via that SBOM, so without
// the plugin an environment has no package inventory at all.
const froshToolsExtension = "FroshTools"

// sbomResult carries what a scrape learned about an environment's Composer
// package inventory. Supported is false when the shop cannot serve an SBOM
// (FroshTools absent or too old); in that case packages is nil and any existing
// rows are left untouched rather than deleted, so a temporarily disabled plugin
// does not wipe the inventory.
type sbomResult struct {
	supported    bool
	fetched      bool
	packages     []sbom.Package
	matches      []sbom.Match
	specVersion  string
	serialNumber string
	generatedAt  pgtype.Timestamp
	err          string
	// fetchErr is the typed upstream error behind a failed SBOM fetch (nil for
	// local failures like an unreadable advisory catalog), kept so the scrape
	// can classify it for the owner instead of only logging err.
	fetchErr error
}

// fetchSbom pulls the CycloneDX SBOM from FroshTools and matches it against the
// advisory catalog. Every failure mode is non-fatal: an environment whose SBOM
// cannot be read still completes its scrape.
func (h *Service) fetchSbom(ctx context.Context, env queries.GetAllEnvironmentsRow, client sbom.Fetcher, extensions []extensionEntry) sbomResult {
	if !hasActiveFroshTools(extensions) {
		return sbomResult{supported: false}
	}
	if client == nil {
		return sbomResult{supported: false}
	}

	ctx, span := tracer.Start(ctx, "environment.scrape.sbom")
	defer span.End()

	doc, err := sbom.Fetch(ctx, client)
	if err != nil {
		if unsupported, ok := errors.AsType[*sbom.ErrUnsupported](err); ok {
			// Expected for FroshTools releases without the SBOM route.
			slog.DebugContext(ctx, "sbom endpoint unavailable",
				"environmentId", env.ID, "reason", unsupported.Reason)
			return sbomResult{supported: false}
		}
		slog.WarnContext(ctx, "failed to fetch sbom", "environmentId", env.ID, "error", err)
		span.RecordError(err)
		return sbomResult{supported: true, err: err.Error(), fetchErr: err}
	}

	packages := doc.Packages()
	if len(packages) == 0 {
		// A Shopware shop always has Composer packages (shopware/core at
		// minimum), so an SBOM parsing to zero of them is a malformed or
		// filtered document, not a real inventory. Treat it like a failed
		// fetch: keep the previously collected components and matches rather
		// than wiping the environment's inventory.
		slog.WarnContext(ctx, "sbom parsed to zero composer packages, keeping previous inventory",
			"environmentId", env.ID)
		return sbomResult{supported: true, err: "sbom contained no composer packages"}
	}

	result := sbomResult{
		supported:    true,
		fetched:      true,
		packages:     packages,
		specVersion:  sbom.SanitizeMeta(doc.SpecVersion),
		serialNumber: sbom.SanitizeMeta(doc.SerialNumber),
	}
	if ts, ok := doc.GeneratedAt(); ok {
		result.generatedAt = pgtype.Timestamp{Time: ts, Valid: true}
	}

	matches, err := h.matchAdvisories(ctx, env.ID, doc.VersionMap())
	if err != nil {
		// Keep the previously stored inventory and matches: replacing them from
		// an unreadable catalog would clear this environment's vulnerabilities.
		slog.WarnContext(ctx, "failed to match sbom against advisory catalog, keeping previous matches",
			"environmentId", env.ID, "error", err)
		span.RecordError(err)
		return sbomResult{supported: true, err: "advisory catalog unavailable: " + err.Error()}
	}
	result.matches = matches

	span.SetAttributes(
		attribute.Int("sbom.components", len(packages)),
		attribute.Int("sbom.advisory_matches", len(result.matches)),
	)
	slog.DebugContext(ctx, "collected sbom",
		"environmentId", env.ID, "components", len(packages), "advisoryMatches", len(result.matches))
	return result
}

// matchAdvisories evaluates an environment's installed package versions against
// the advisory catalog.
//
// The error is returned rather than folded into an empty result: persistSbom
// replaces the whole match set, so "the catalog could not be read" and "nothing
// matches" must not look the same — the former would erase every known
// vulnerability for the environment.
func (h *Service) matchAdvisories(ctx context.Context, environmentID int32, installed map[string]string) ([]sbom.Match, error) {
	if len(installed) == 0 {
		return nil, nil
	}

	index, err := loadAdvisoryIndex(ctx, h.queries)
	if err != nil {
		return nil, fmt.Errorf("load advisory index: %w", err)
	}
	return index.Match(installed), nil
}

// loadAdvisoryIndex builds a package-name-keyed index over the visible advisory
// catalog.
func loadAdvisoryIndex(ctx context.Context, q *queries.Queries) (*sbom.Index, error) {
	rows, err := q.ListVisibleComposerAdvisories(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return sbom.NewIndex(nil), nil
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AdvisoryID)
	}

	pkgRows, err := q.ListComposerAdvisoryPackagesForAdvisories(ctx, ids)
	if err != nil {
		return nil, err
	}

	advPkgs := make([]sbom.AdvisoryPackage, 0, len(pkgRows))
	for _, p := range pkgRows {
		advPkgs = append(advPkgs, sbom.AdvisoryPackage{
			AdvisoryID:       p.AdvisoryID,
			PackageName:      p.PackageName,
			AffectedVersions: p.AffectedVersions,
		})
	}
	return sbom.NewIndex(advPkgs), nil
}

func hasActiveFroshTools(extensions []extensionEntry) bool {
	for _, ext := range extensions {
		if ext.Name == froshToolsExtension && ext.Active && ext.Installed {
			return true
		}
	}
	return false
}

// persistSbom writes the component inventory, the advisory matches, and the
// per-environment SBOM state. It runs inside the scrape transaction.
func persistSbom(ctx context.Context, q *queries.Queries, environmentID int32, res sbomResult) error {
	if !res.supported {
		// Record that this shop cannot serve an SBOM, but keep any previously
		// collected components (the plugin may just be temporarily inactive)
		// along with their count and provenance.
		return q.UpsertEnvironmentSbomStateStatus(ctx, queries.UpsertEnvironmentSbomStateStatusParams{
			EnvironmentID: environmentID,
			Supported:     false,
			LastError:     nil,
		})
	}

	// A fetched result with zero packages must never reach the pruning below:
	// DeleteEnvironmentSbomComponentsNotIn with an empty keep-list deletes the
	// environment's entire inventory. fetchSbom already downgrades that case,
	// but the guard belongs on the write path.
	if !res.fetched || len(res.packages) == 0 {
		errText := res.err
		if errText == "" {
			errText = "sbom contained no composer packages"
		}
		return q.UpsertEnvironmentSbomStateStatus(ctx, queries.UpsertEnvironmentSbomStateStatusParams{
			EnvironmentID: environmentID,
			Supported:     true,
			LastError:     new(sbom.SanitizeError(errText)),
		})
	}

	names := make([]string, 0, len(res.packages))
	for _, p := range res.packages {
		if err := q.UpsertEnvironmentSbomComponent(ctx, queries.UpsertEnvironmentSbomComponentParams{
			EnvironmentID: environmentID,
			PackageName:   p.Name,
			Version:       p.Version,
			PackageType:   emptyToNil(p.Type),
			Purl:          emptyToNil(p.Purl),
			IsDev:         false,
		}); err != nil {
			return err
		}
		names = append(names, p.Name)
	}

	if err := q.DeleteEnvironmentSbomComponentsNotIn(ctx, queries.DeleteEnvironmentSbomComponentsNotInParams{
		EnvironmentID: environmentID,
		Column2:       names,
	}); err != nil {
		return err
	}

	// Matches are fully recomputed from the current inventory, so replacing the
	// set is simpler (and safer) than diffing: a package that was upgraded out
	// of range must lose its match row. Runtime matches (php, mysql, mariadb)
	// are managed separately by persistRuntimeAdvisories so an SBOM write never
	// wipes them.
	if err := q.DeleteEnvironmentComposerAdvisoryMatches(ctx, environmentID); err != nil {
		return err
	}
	for _, m := range res.matches {
		if err := q.UpsertEnvironmentAdvisoryMatch(ctx, queries.UpsertEnvironmentAdvisoryMatchParams{
			EnvironmentID:    environmentID,
			AdvisoryID:       m.AdvisoryID,
			PackageName:      m.PackageName,
			InstalledVersion: m.InstalledVersion,
			AffectedVersions: m.AffectedVersions,
		}); err != nil {
			return err
		}
	}

	return q.UpsertEnvironmentSbomState(ctx, queries.UpsertEnvironmentSbomStateParams{
		EnvironmentID:  environmentID,
		Supported:      true,
		ComponentCount: int32(len(res.packages)),
		SpecVersion:    emptyToNil(res.specVersion),
		SerialNumber:   emptyToNil(res.serialNumber),
		GeneratedAt:    res.generatedAt,
		LastError:      nil,
	})
}

type runtimeMatchResult struct {
	matched bool
	matches []sbom.Match
}

// matchRuntimeAdvisories evaluates an environment's PHP and database versions
// against the advisory catalog. It runs independently of the Composer SBOM so
// shops without the SBOM endpoint (or with a transient SBOM error) still
// refresh their runtime advisory matches on every scrape. On a catalog read
// error, matched is false and persistRuntimeAdvisories leaves existing runtime
// match rows untouched.
func (h *Service) matchRuntimeAdvisories(ctx context.Context, environmentID int32, phpVersion, databaseVersion *string) runtimeMatchResult {
	installed := withRuntimePackageVersions(nil, phpVersion, databaseVersion)
	if len(installed) == 0 {
		return runtimeMatchResult{}
	}

	ctx, span := tracer.Start(ctx, "environment.scrape.runtime_advisories")
	defer span.End()

	matches, err := h.matchAdvisories(ctx, environmentID, installed)
	if err != nil {
		slog.WarnContext(ctx, "failed to match runtime versions against advisory catalog, keeping previous runtime matches",
			"environmentId", environmentID, "error", err)
		span.RecordError(err)
		return runtimeMatchResult{}
	}

	span.SetAttributes(attribute.Int("runtime.advisory_matches", len(matches)))
	return runtimeMatchResult{
		matched: true,
		matches: matches,
	}
}

func persistRuntimeAdvisories(ctx context.Context, q *queries.Queries, environmentID int32, res runtimeMatchResult) error {
	if !res.matched {
		return nil
	}
	if err := q.DeleteEnvironmentRuntimeAdvisoryMatches(ctx, environmentID); err != nil {
		return err
	}
	for _, m := range res.matches {
		if err := q.UpsertEnvironmentAdvisoryMatch(ctx, queries.UpsertEnvironmentAdvisoryMatchParams{
			EnvironmentID:    environmentID,
			AdvisoryID:       m.AdvisoryID,
			PackageName:      m.PackageName,
			InstalledVersion: m.InstalledVersion,
			AffectedVersions: m.AffectedVersions,
		}); err != nil {
			return err
		}
	}
	return nil
}

// sbomPackageVersions flattens a scrape's SBOM into the package name -> version
// map the security checker consumes. Nil when no SBOM was collected.
func sbomPackageVersions(res sbomResult) map[string]string {
	if !res.fetched || len(res.packages) == 0 {
		return nil
	}
	out := make(map[string]string, len(res.packages))
	for _, p := range res.packages {
		out[p.Name] = p.Version
	}
	return out
}

func withRuntimePackageVersions(packages map[string]string, phpVersion, databaseVersion *string) map[string]string {
	hasPhp := phpVersion != nil && *phpVersion != ""
	hasDatabase := databaseVersion != nil && *databaseVersion != ""
	if !hasPhp && !hasDatabase {
		return packages
	}
	out := make(map[string]string, len(packages)+2)
	for k, v := range packages {
		out[k] = v
	}
	if hasPhp {
		out["php"] = *phpVersion
	}
	if hasDatabase {
		if pkg, ver := version.ParseDatabaseVersion(*databaseVersion); pkg != "" && ver != "" {
			out[pkg] = ver
		}
	}
	return out
}

func emptyToNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// suppressedCheckIDs returns the check ids the environment's active advisory
// suppressions cover. A failure degrades to "nothing suppressed" rather than
// failing the scrape: showing an acknowledged advisory is noisy, but hiding a
// real one because a query failed would be unsafe.
func (h *Service) suppressedCheckIDs(ctx context.Context, environmentID int32) []string {
	rows, err := h.queries.ListSuppressedAdvisoriesForEnvironment(ctx, environmentID)
	if err != nil {
		slog.WarnContext(ctx, "failed to load advisory suppressions",
			"environmentId", environmentID, "error", err)
		return nil
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		var cve, ghsa string
		if row.Cve != nil {
			cve = *row.Cve
		}
		if row.GhsaID != nil {
			ghsa = *row.GhsaID
		}
		out = append(out, checker.SecurityCheckID(cve, ghsa, row.AdvisoryID))
	}
	return out
}

// loadSecurityPluginFixes reads the SwagPlatformSecurity backport map. A failure
// degrades to an empty map, which makes every advisory resolve to "unknown" and
// therefore be reported — the safe direction, since the alternative would be
// crediting shops with fixes we cannot prove they have.
func loadSecurityPluginFixes(ctx context.Context, q *queries.Queries) ([]checker.PluginFix, checker.PluginCoverage) {
	fixRows, err := q.ListSecurityPluginFixes(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to load security plugin fixes", "error", err)
		return nil, nil
	}
	coverageRows, err := q.ListSecurityPluginCoverage(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to load security plugin coverage", "error", err)
		return nil, nil
	}

	fixes := make([]checker.PluginFix, 0, len(fixRows))
	for _, row := range fixRows {
		fixes = append(fixes, checker.PluginFix{
			GhsaID:        row.GhsaID,
			Branch:        row.PluginBranch,
			PluginVersion: row.PluginVersion,
		})
	}

	coverage := make(checker.PluginCoverage, len(coverageRows))
	for _, row := range coverageRows {
		coverage[row.PluginBranch] = int(row.GhsaCount)
	}
	return fixes, coverage
}
