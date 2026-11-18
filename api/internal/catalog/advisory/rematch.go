package advisory

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/friendsofshopware/shopmon/api/internal/version"
)

// rematchEnvironments recomputes the advisory matches for every environment
// that has an SBOM on file.
//
// Scrapes already match an environment's packages when they run, but that only
// covers advisories that existed at scrape time. Running this after each sync
// means a newly published CVE surfaces against already-known package inventory
// immediately, instead of waiting for each shop's next scrape.
func (s *Service) rematchEnvironments(ctx context.Context) (int, error) {
	ctx, span := tracer.Start(ctx, "shopware.advisory.rematch")
	defer span.End()

	environmentIDs, err := s.queries.ListEnvironmentIDsWithSbom(ctx)
	if err != nil {
		return 0, fmt.Errorf("list environments with sbom: %w", err)
	}
	if len(environmentIDs) == 0 {
		return 0, nil
	}

	index, err := LoadIndex(ctx, s.queries)
	if err != nil {
		return 0, fmt.Errorf("load advisory index: %w", err)
	}
	fixes, err := LoadSecurityPluginFixes(ctx, s.queries)
	if err != nil {
		return 0, err
	}
	pluginFixes := checker.NewPluginFixIndex(fixes)

	var matched int
	for _, environmentID := range environmentIDs {
		count, err := s.rematchEnvironment(ctx, environmentID, index, pluginFixes)
		if err != nil {
			// One unreadable environment must not abort the whole pass.
			slog.WarnContext(ctx, "failed to rematch environment advisories",
				"environmentId", environmentID, "error", err)
			continue
		}
		matched += count
	}
	return matched, nil
}

func (s *Service) rematchEnvironment(ctx context.Context, environmentID int32, index *sbom.Index, pluginFixes checker.PluginFixIndex) (int, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin rematch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	txq := s.queries.WithTx(tx)

	env, err := txq.LockEnvironmentForAdvisoryMatch(ctx, environmentID)
	if err != nil {
		return 0, fmt.Errorf("lock environment for rematch: %w", err)
	}
	components, err := txq.GetEnvironmentSbomComponents(ctx, environmentID)
	if err != nil {
		return 0, fmt.Errorf("get sbom components: %w", err)
	}

	installed := make(map[string]string, len(components)+2)
	for _, c := range components {
		installed[c.PackageName] = c.Version
	}
	if env.PhpVersion != nil && *env.PhpVersion != "" {
		installed["php"] = *env.PhpVersion
	}
	if env.DatabaseVersion != nil && *env.DatabaseVersion != "" {
		if pkg, ver := version.ParseDatabaseVersion(*env.DatabaseVersion); pkg != "" && ver != "" {
			installed[pkg] = ver
		}
	}
	if len(installed) == 0 {
		return 0, nil
	}

	matches := index.Match(installed)
	if len(matches) > 0 && len(pluginFixes) > 0 {
		pluginVersion, err := txq.GetActiveEnvironmentSecurityPluginVersion(ctx, environmentID)
		if err != nil {
			return 0, fmt.Errorf("get security plugin version: %w", err)
		}
		if pluginVersion != "" {
			plugin := checker.Extension{Version: pluginVersion, Active: true, Installed: true}
			matches = pluginFixes.ExcludeMitigatedMatches(matches, env.ShopwareVersion, &plugin)
		}
	}

	// "Already known" means "subscribers were already told", not "a match row
	// exists". The match rows below must commit whether or not the alert
	// succeeds — they drive the UI and shop status — so keying dedup on them
	// would make any transient failure in the notify path permanent: the next
	// pass would see the row and stay silent forever.
	notified, err := txq.ListNotifiedAdvisoryIDsForEnvironment(ctx, environmentID)
	if err != nil {
		return 0, fmt.Errorf("list notified advisories: %w", err)
	}
	known := make(map[string]bool, len(notified))
	for _, advisoryID := range notified {
		known[advisoryID] = true
	}

	// Replace rather than merge: a package upgraded out of an affected range
	// must lose its match row. The delete and re-inserts share one transaction
	// so a crash mid-loop cannot leave the environment with an empty match set
	// — which would both hide its vulnerabilities and, by emptying the next
	// pass's known snapshot, re-notify every advisory as if it were new.
	if err := txq.DeleteEnvironmentAdvisoryMatches(ctx, environmentID); err != nil {
		return 0, fmt.Errorf("delete advisory matches: %w", err)
	}
	for _, m := range matches {
		if err := txq.UpsertEnvironmentAdvisoryMatch(ctx, queries.UpsertEnvironmentAdvisoryMatchParams{
			EnvironmentID:    environmentID,
			AdvisoryID:       m.AdvisoryID,
			PackageName:      m.PackageName,
			InstalledVersion: m.InstalledVersion,
			AffectedVersions: m.AffectedVersions,
		}); err != nil {
			return 0, fmt.Errorf("upsert advisory match: %w", err)
		}
	}
	// Retire markers for advisories that no longer match, so one that reappears
	// later (package downgraded, affected range widened) alerts again.
	if err := txq.DeleteStaleEnvironmentAdvisoryNotifications(ctx, environmentID); err != nil {
		return 0, fmt.Errorf("prune stale advisory notification markers: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit rematch transaction: %w", err)
	}

	s.notifyNewAdvisories(ctx, environmentID, matches, known)

	return len(matches), nil
}

// LoadIndex builds a package-keyed index over the visible advisory
// catalog, shared across all environments in one rematch pass.
func LoadIndex(ctx context.Context, q *queries.Queries) (*sbom.Index, error) {
	rows, err := q.ListVisibleComposerAdvisories(ctx)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return sbom.NewIndex(nil), nil
	}

	ids := make([]string, 0, len(rows))
	ghsaByID := make(map[string]string, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AdvisoryID)
		if row.GhsaID != nil {
			ghsaByID[row.AdvisoryID] = *row.GhsaID
		}
	}

	pkgRows, err := q.ListComposerAdvisoryPackagesForAdvisories(ctx, ids)
	if err != nil {
		return nil, err
	}

	advPkgs := make([]sbom.AdvisoryPackage, 0, len(pkgRows))
	for _, p := range pkgRows {
		advPkgs = append(advPkgs, sbom.AdvisoryPackage{
			AdvisoryID:       p.AdvisoryID,
			GHSA:             ghsaByID[p.AdvisoryID],
			PackageName:      p.PackageName,
			AffectedVersions: p.AffectedVersions,
		})
	}
	return sbom.NewIndex(advPkgs), nil
}

// LoadSecurityPluginFixes reads the combined verified and changelog backport map.
func LoadSecurityPluginFixes(ctx context.Context, q *queries.Queries) ([]checker.PluginFix, error) {
	fixRows, err := q.ListSecurityPluginFixes(ctx)
	if err != nil {
		return nil, fmt.Errorf("list security plugin fixes: %w", err)
	}

	fixes := make([]checker.PluginFix, 0, len(fixRows))
	for _, row := range fixRows {
		fixes = append(fixes, checker.PluginFix{
			GhsaID:        row.GhsaID,
			Branch:        row.PluginBranch,
			PluginVersion: row.PluginVersion,
		})
	}
	return fixes, nil
}
