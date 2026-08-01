package securityplugin

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"

	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
)

// ExtensionName is the technical name of the Shopware Security Plugin.
const ExtensionName = "SwagPlatformSecurity"

// anchorVersions ensure every known branch is probed even when no monitored
// shop runs that Shopware line yet. Without them a fleet that is entirely on
// 6.7 would never learn what 3.x backports, and a shop upgrading onto 6.6 would
// start with an empty map.
var anchorVersions = []string{"6.7.0.0", "6.6.0.0", "6.5.0.0"}

// coverageDropTolerance guards against markup drift. If a re-parse finds fewer
// than this fraction of the identifiers a branch had before, the existing rows
// are kept: a store template change must not silently convert "covered" shops
// into "not backported" ones, or vice versa.
const coverageDropTolerance = 0.5

var tracer = otel.Tracer("shopmon/catalog/securityplugin")

// StoreClient fetches plugin metadata from the Shopware account API.
type StoreClient interface {
	PluginsByName(ctx context.Context, locale, shopwareVersion string, technicalNames []string) ([]shopwareaccount.StorePlugin, error)
}

// Service refreshes the GHSA -> plugin version map.
//
// It calls the store directly rather than routing through the extension catalog
// sync: that sync is driven by installed extensions, and writing catalog and
// compatibility rows for a plugin nobody has installed would pollute the
// extension catalog and its orphan-cleanup queries.
type Service struct {
	queries *queries.Queries
	client  StoreClient
	now     func() time.Time
}

func NewService(q *queries.Queries, client StoreClient) *Service {
	return &Service{queries: q, client: client, now: time.Now}
}

// Sync re-derives the whole map from the store changelog.
//
// The changelog is small (under a hundred entries), so a full recompute is
// cheaper than tracking deltas and is self-correcting: a fix to the parser
// takes effect on the next run without a backfill.
func (s *Service) Sync(ctx context.Context) error {
	ctx, span := tracer.Start(ctx, "shopware.securityplugin.sync")
	defer span.End()

	entries, err := s.collectChangelog(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	if len(entries) == 0 {
		slog.WarnContext(ctx, "security plugin changelog empty, keeping existing map")
		return nil
	}

	fixes, coverage := DeriveFixes(entries)

	previous, err := s.previousCoverage(ctx)
	if err != nil {
		return err
	}

	branches := make([]string, 0, len(coverage))
	for _, cov := range coverage {
		if prev, ok := previous[cov.Branch]; ok && prev > 0 {
			if float64(cov.GhsaCount) < float64(prev)*coverageDropTolerance {
				slog.ErrorContext(ctx, "security plugin advisory references dropped sharply, keeping existing fixes",
					"branch", cov.Branch, "previous", prev, "parsed", cov.GhsaCount)
				reason := fmt.Sprintf("parse rejected: ghsa references dropped from %d to %d", prev, cov.GhsaCount)
				if err := s.queries.SetSecurityPluginCoverageError(ctx, queries.SetSecurityPluginCoverageErrorParams{
					PluginBranch: cov.Branch,
					LastError:    &reason,
				}); err != nil {
					slog.WarnContext(ctx, "failed to record coverage rejection", "branch", cov.Branch, "error", err)
				}
				continue
			}
		}
		branches = append(branches, cov.Branch)
	}

	if err := s.persist(ctx, fixes, coverage, branches); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetAttributes(
		attribute.Int("securityplugin.fixes", len(fixes)),
		attribute.Int("securityplugin.branches", len(coverage)),
	)
	slog.InfoContext(ctx, "synced security plugin backport map",
		"fixes", len(fixes), "branches", len(coverage))
	return nil
}

// collectChangelog probes every relevant Shopware version and merges the
// results. One probe already returns entries for all branches, but the endpoint
// omits the plugin entirely when no release supports the requested version, so
// probing several versions is what makes the map robust.
func (s *Service) collectChangelog(ctx context.Context) ([]ChangelogEntry, error) {
	versionsToProbe := map[string]bool{}
	for _, v := range anchorVersions {
		versionsToProbe[v] = true
	}
	inUse, err := s.queries.GetDistinctEnvironmentShopwareVersions(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to list environment shopware versions", "error", err)
	}
	for _, v := range inUse {
		if v != "" {
			versionsToProbe[v] = true
		}
	}

	byVersion := map[string]ChangelogEntry{}
	var probeSucceeded bool

	for shopwareVersion := range versionsToProbe {
		plugins, err := s.client.PluginsByName(ctx, "en_GB", shopwareVersion, []string{ExtensionName})
		if err != nil {
			slog.WarnContext(ctx, "failed to probe security plugin",
				"shopwareVersion", shopwareVersion, "error", err)
			continue
		}
		probeSucceeded = true

		for _, plugin := range plugins {
			if plugin.Name != ExtensionName {
				continue
			}
			for _, entry := range plugin.Changelogs {
				if entry.Version == "" {
					continue
				}
				if _, seen := byVersion[entry.Version]; seen {
					continue
				}
				byVersion[entry.Version] = ChangelogEntry{
					Version:    entry.Version,
					Changelog:  entry.Text,
					ReleasedAt: parseStoreDate(entry.CreationDate.Date),
				}
			}
		}
	}

	// Fall back to the catalog copy, which exists whenever a monitored shop has
	// the plugin installed. Free, and keeps the map fresh if the store is down.
	if rows, err := s.queries.GetStoreExtensionChangelogs(ctx, ExtensionName); err == nil {
		for _, row := range rows {
			if row.Version == "" || row.ChangelogEn == nil {
				continue
			}
			if _, seen := byVersion[row.Version]; seen {
				continue
			}
			byVersion[row.Version] = ChangelogEntry{Version: row.Version, Changelog: *row.ChangelogEn}
		}
	}

	if !probeSucceeded && len(byVersion) == 0 {
		return nil, fmt.Errorf("no security plugin changelog could be fetched")
	}

	out := make([]ChangelogEntry, 0, len(byVersion))
	for _, entry := range byVersion {
		out = append(out, entry)
	}
	return out, nil
}

func (s *Service) previousCoverage(ctx context.Context) (map[string]int32, error) {
	rows, err := s.queries.ListSecurityPluginCoverage(ctx)
	if err != nil {
		return nil, fmt.Errorf("list security plugin coverage: %w", err)
	}
	out := make(map[string]int32, len(rows))
	for _, row := range rows {
		out[row.PluginBranch] = row.GhsaCount
	}
	return out, nil
}

func (s *Service) persist(ctx context.Context, fixes []Fix, coverage []Coverage, branches []string) error {
	accepted := make(map[string]bool, len(branches))
	for _, branch := range branches {
		accepted[branch] = true
	}

	ghsaIDs := make([]string, 0, len(fixes))
	for _, fix := range fixes {
		if !accepted[fix.Branch] {
			continue
		}
		if err := s.queries.UpsertSecurityPluginFix(ctx, queries.UpsertSecurityPluginFixParams{
			GhsaID:         fix.GhsaID,
			PluginBranch:   fix.Branch,
			PluginVersion:  fix.PluginVersion,
			ShopwareBranch: ShopwareBranch(fix.Branch),
			ReleasedAt:     database.TimestampFromPtr(fix.ReleasedAt),
		}); err != nil {
			return fmt.Errorf("upsert security plugin fix: %w", err)
		}
		ghsaIDs = append(ghsaIDs, fix.GhsaID)
	}

	if len(branches) > 0 {
		if err := s.queries.DeleteSecurityPluginFixesNotIn(ctx, queries.DeleteSecurityPluginFixesNotInParams{
			Branches: branches,
			GhsaIds:  ghsaIDs,
		}); err != nil {
			return fmt.Errorf("prune security plugin fixes: %w", err)
		}
	}

	for _, cov := range coverage {
		if !accepted[cov.Branch] {
			continue
		}
		if err := s.queries.UpsertSecurityPluginCoverage(ctx, queries.UpsertSecurityPluginCoverageParams{
			PluginBranch:     cov.Branch,
			MinParsedVersion: cov.MinParsedVersion,
			MaxParsedVersion: cov.MaxParsedVersion,
			GhsaCount:        int32(cov.GhsaCount),
		}); err != nil {
			return fmt.Errorf("upsert security plugin coverage: %w", err)
		}
	}
	return nil
}

func parseStoreDate(value string) *time.Time {
	if value == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02 15:04:05.000000", "2006-01-02 15:04:05", time.RFC3339} {
		if t, err := time.Parse(layout, value); err == nil {
			utc := t.UTC()
			return &utc
		}
	}
	return nil
}
