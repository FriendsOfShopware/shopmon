package securityplugin

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/friendsofshopware/shopmon/api/internal/shopwarepackages"
	"github.com/friendsofshopware/shopmon/api/internal/testutil"
)

type stubFeedClient struct {
	pkg *shopwarepackages.Package
	err error
}

func (s stubFeedClient) PackageFeed(context.Context, string) (*shopwarepackages.Package, error) {
	return s.pkg, s.err
}

func feedVersion(version, changelog string) shopwarepackages.PackageVersion {
	return shopwarepackages.PackageVersion{Version: version, Changelog: changelog}
}

// The feed replaces the store API probing: a single response covering every
// branch must derive the same per-branch minimums the merged probes did.
func TestSyncDerivesFixesFromFeed(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	released := time.Date(2026, 5, 19, 13, 25, 35, 0, time.UTC)
	feed := &shopwarepackages.Package{
		Name: PackageName,
		Versions: []shopwarepackages.PackageVersion{
			{Version: "4.0.12", Changelog: "Improved the fix for GHSA-aaaa-bbbb-cccc", ReleasedAt: &released},
			{Version: "4.0.10", Changelog: "Added fix for GHSA-aaaa-bbbb-cccc", ReleasedAt: &released},
			{Version: "3.0.15", Changelog: "Added fix for GHSA-aaaa-bbbb-cccc"},
			{Version: "2.0.20", Changelog: "unrelated improvements"},
		},
	}
	svc := NewService(env.Pool, env.Queries, stubFeedClient{pkg: feed})

	require.NoError(t, svc.Sync(ctx))

	fixes, err := env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	require.Len(t, fixes, 2)

	byBranch := map[string]string{}
	for _, fix := range fixes {
		byBranch[fix.PluginBranch] = fix.PluginVersion
	}
	// The 4.0.12 refinement must not raise the required version past the
	// release that actually shipped the fix.
	assert.Equal(t, "4.0.10", byBranch["4"])
	assert.Equal(t, "3.0.15", byBranch["3"])
	_, hasBranch2 := byBranch["2"]
	assert.False(t, hasBranch2, "branch 2 never named the advisory")

	coverage, err := env.Queries.ListSecurityPluginCoverage(ctx)
	require.NoError(t, err)
	require.Len(t, coverage, 3, "every branch in the feed is recorded as parsed")
}

func TestSyncFeedFailureWithEmptyCatalogErrors(t *testing.T) {
	env := testutil.Setup(t)

	svc := NewService(env.Pool, env.Queries, stubFeedClient{err: errors.New("feed down")})

	require.Error(t, svc.Sync(context.Background()))
}

// When the feed is unreachable, the catalog copy synced from monitored shops
// keeps the map fresh.
func TestSyncFeedFailureFallsBackToCatalog(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	_, err := env.Pool.Exec(ctx, `INSERT INTO store_extension (name) VALUES ($1)`, ExtensionName)
	require.NoError(t, err)
	var versionID int
	require.NoError(t, env.Pool.QueryRow(ctx,
		`INSERT INTO store_extension_version (extension_name, version) VALUES ($1, '4.0.10') RETURNING id`,
		ExtensionName).Scan(&versionID))
	_, err = env.Pool.Exec(ctx,
		`INSERT INTO store_extension_version_translation (extension_version_id, language, changelog) VALUES ($1, 'en', 'Added fix for GHSA-aaaa-bbbb-cccc')`,
		versionID)
	require.NoError(t, err)

	svc := NewService(env.Pool, env.Queries, stubFeedClient{err: errors.New("feed down")})
	require.NoError(t, svc.Sync(ctx))

	fixes, err := env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	require.Len(t, fixes, 1)
	assert.Equal(t, "GHSA-AAAA-BBBB-CCCC", fixes[0].GhsaID)
	assert.Equal(t, "4.0.10", fixes[0].PluginVersion)
}

// A feed that answers but lists no versions must not wipe the existing map.
func TestSyncEmptyFeedKeepsExistingMap(t *testing.T) {
	env := testutil.Setup(t)
	ctx := context.Background()

	svc := NewService(env.Pool, env.Queries, stubFeedClient{pkg: &shopwarepackages.Package{
		Name:     PackageName,
		Versions: []shopwarepackages.PackageVersion{feedVersion("4.0.10", "Added fix for GHSA-aaaa-bbbb-cccc")},
	}})
	require.NoError(t, svc.Sync(ctx))

	svc = NewService(env.Pool, env.Queries, stubFeedClient{pkg: &shopwarepackages.Package{Name: PackageName}})
	require.NoError(t, svc.Sync(ctx), "an empty feed is a warn, not a failure")

	fixes, err := env.Queries.ListSecurityPluginFixes(ctx)
	require.NoError(t, err)
	assert.Len(t, fixes, 1, "existing fixes survive an empty feed")
}
