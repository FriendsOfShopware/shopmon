package advisory

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
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

	index, err := s.loadAdvisoryIndex(ctx)
	if err != nil {
		return 0, fmt.Errorf("load advisory index: %w", err)
	}

	var matched int
	for _, environmentID := range environmentIDs {
		count, err := s.rematchEnvironment(ctx, environmentID, index)
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

func (s *Service) rematchEnvironment(ctx context.Context, environmentID int32, index *sbom.Index) (int, error) {
	components, err := s.queries.GetEnvironmentSbomComponents(ctx, environmentID)
	if err != nil {
		return 0, fmt.Errorf("get sbom components: %w", err)
	}
	if len(components) == 0 {
		return 0, nil
	}

	installed := make(map[string]string, len(components))
	for _, c := range components {
		installed[c.PackageName] = c.Version
	}

	matches := index.Match(installed)

	// Snapshot what already matched, before the replace below, so only genuinely
	// new advisories trigger an alert. Without this every sync would re-notify
	// the same findings.
	previous, err := s.queries.GetEnvironmentAdvisoryMatches(ctx, environmentID)
	if err != nil {
		return 0, fmt.Errorf("get existing advisory matches: %w", err)
	}
	known := make(map[string]bool, len(previous))
	for _, row := range previous {
		known[row.AdvisoryID] = true
	}

	// Replace rather than merge: a package upgraded out of an affected range
	// must lose its match row. The delete and re-inserts share one transaction
	// so a crash mid-loop cannot leave the environment with an empty match set
	// — which would both hide its vulnerabilities and, by emptying the next
	// pass's known snapshot, re-notify every advisory as if it were new.
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return 0, fmt.Errorf("begin rematch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txq := s.queries.WithTx(tx)
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
	if err := tx.Commit(ctx); err != nil {
		return 0, fmt.Errorf("commit rematch transaction: %w", err)
	}

	s.notifyNewAdvisories(ctx, environmentID, matches, known)

	return len(matches), nil
}

// loadAdvisoryIndex builds a package-keyed index over the visible advisory
// catalog, shared across all environments in one rematch pass.
func (s *Service) loadAdvisoryIndex(ctx context.Context) (*sbom.Index, error) {
	rows, err := s.queries.ListVisibleComposerAdvisories(ctx)
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

	pkgRows, err := s.queries.ListComposerAdvisoryPackagesForAdvisories(ctx, ids)
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
