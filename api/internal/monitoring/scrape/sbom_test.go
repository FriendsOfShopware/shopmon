package scrape

import (
	"context"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/shopware/sbom"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHasActiveFroshTools(t *testing.T) {
	t.Parallel()

	assert.True(t, hasActiveFroshTools([]extensionEntry{
		{Name: froshToolsExtension, Active: true, Installed: true},
	}))
	assert.False(t, hasActiveFroshTools([]extensionEntry{
		{Name: froshToolsExtension, Active: false, Installed: true},
	}))
	assert.False(t, hasActiveFroshTools([]extensionEntry{
		{Name: froshToolsExtension, Active: true, Installed: false},
	}))
	assert.False(t, hasActiveFroshTools([]extensionEntry{
		{Name: "OtherPlugin", Active: true, Installed: true},
	}))
}

func TestSbomPackageVersions(t *testing.T) {
	t.Parallel()

	assert.Nil(t, sbomPackageVersions(sbomResult{supported: false}))
	assert.Nil(t, sbomPackageVersions(sbomResult{supported: true, fetched: false}))

	got := sbomPackageVersions(sbomResult{
		supported: true,
		fetched:   true,
		packages: []sbom.Package{
			{Name: "shopware/core", Version: "6.6.10.0"},
			{Name: "symfony/http-kernel", Version: "6.4.2"},
		},
	})
	assert.Equal(t, map[string]string{
		"shopware/core":       "6.6.10.0",
		"symfony/http-kernel": "6.4.2",
	}, got)
}

func TestPersistSbomStoresComponentsAndMatches(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	seedAdvisory(t, pool, "CVE-2026-100", "symfony/http-kernel", ">=6.4.0,<6.4.3")

	res := sbomResult{
		supported:   true,
		fetched:     true,
		specVersion: "1.7",
		packages: []sbom.Package{
			{Name: "shopware/core", Version: "6.6.10.0", Type: "library", Purl: "pkg:composer/shopware/core@6.6.10.0"},
			{Name: "symfony/http-kernel", Version: "6.4.2", Type: "library"},
		},
		matches: []sbom.Match{
			{AdvisoryID: "CVE-2026-100", PackageName: "symfony/http-kernel", InstalledVersion: "6.4.2", AffectedVersions: ">=6.4.0,<6.4.3"},
		},
	}

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, res))

	components, err := h.queries.GetEnvironmentSbomComponents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, components, 2)
	assert.Equal(t, "shopware/core", components[0].PackageName)
	assert.Equal(t, "6.6.10.0", components[0].Version)

	state, err := h.queries.GetEnvironmentSbomState(ctx, env.ID)
	require.NoError(t, err)
	assert.True(t, state.Supported)
	assert.Equal(t, int32(2), state.ComponentCount)

	matches, err := h.queries.GetEnvironmentAdvisoryMatches(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, matches, 1)
	assert.Equal(t, "CVE-2026-100", matches[0].AdvisoryID)
	assert.Equal(t, "6.4.2", matches[0].InstalledVersion)
}

// A package removed from composer.lock must disappear from the inventory, and a
// package upgraded out of an affected range must lose its match row.
func TestPersistSbomPrunesStaleData(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	seedAdvisory(t, pool, "CVE-2026-200", "symfony/http-kernel", ">=6.4.0,<6.4.3")

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{
		supported: true,
		fetched:   true,
		packages: []sbom.Package{
			{Name: "shopware/core", Version: "6.6.10.0"},
			{Name: "symfony/http-kernel", Version: "6.4.2"},
			{Name: "old/removed", Version: "1.0.0"},
		},
		matches: []sbom.Match{
			{AdvisoryID: "CVE-2026-200", PackageName: "symfony/http-kernel", InstalledVersion: "6.4.2", AffectedVersions: ">=6.4.0,<6.4.3"},
		},
	}))

	// Second scrape: dependency patched, one package removed.
	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{
		supported: true,
		fetched:   true,
		packages: []sbom.Package{
			{Name: "shopware/core", Version: "6.6.10.0"},
			{Name: "symfony/http-kernel", Version: "6.4.3"},
		},
	}))

	components, err := h.queries.GetEnvironmentSbomComponents(ctx, env.ID)
	require.NoError(t, err)
	require.Len(t, components, 2)
	names := []string{components[0].PackageName, components[1].PackageName}
	assert.NotContains(t, names, "old/removed")

	matches, err := h.queries.GetEnvironmentAdvisoryMatches(ctx, env.ID)
	require.NoError(t, err)
	assert.Empty(t, matches, "patched dependency should no longer match")
}

// An environment whose FroshTools cannot serve an SBOM keeps whatever inventory
// was previously collected: a temporarily disabled plugin must not wipe it.
func TestPersistSbomUnsupportedKeepsComponents(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{
		supported: true,
		fetched:   true,
		packages:  []sbom.Package{{Name: "shopware/core", Version: "6.6.10.0"}},
	}))

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{supported: false}))

	components, err := h.queries.GetEnvironmentSbomComponents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, components, 1, "components should survive an unsupported scrape")

	state, err := h.queries.GetEnvironmentSbomState(ctx, env.ID)
	require.NoError(t, err)
	assert.False(t, state.Supported)
}

// A fetched result carrying zero packages (malformed or filtered SBOM) must
// not wipe the environment's inventory: DeleteEnvironmentSbomComponentsNotIn
// with an empty keep-list would delete every row.
func TestPersistSbomZeroPackagesKeepsInventory(t *testing.T) {
	h, _, env := setupPersistTest(t)
	ctx := context.Background()

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{
		supported: true,
		fetched:   true,
		packages:  []sbom.Package{{Name: "shopware/core", Version: "6.6.10.0"}},
	}))

	require.NoError(t, persistSbom(ctx, h.queries, env.ID, sbomResult{
		supported: true,
		fetched:   true,
	}))

	components, err := h.queries.GetEnvironmentSbomComponents(ctx, env.ID)
	require.NoError(t, err)
	assert.Len(t, components, 1, "components should survive a zero-package scrape")
}

func TestLoadAdvisoryIndexMatchesInstalledPackages(t *testing.T) {
	h, pool, _ := setupPersistTest(t)
	ctx := context.Background()

	seedAdvisory(t, pool, "CVE-2026-300", "twig/twig", "<3.9.0")

	index, err := loadAdvisoryIndex(ctx, h.queries)
	require.NoError(t, err)

	matches := index.Match(map[string]string{"twig/twig": "3.8.0"})
	require.Len(t, matches, 1)
	assert.Equal(t, "CVE-2026-300", matches[0].AdvisoryID)

	assert.Empty(t, index.Match(map[string]string{"twig/twig": "3.9.0"}))
}

// seedAdvisory inserts a visible advisory covering one package version range.
func seedAdvisory(t *testing.T, pool *pgxpool.Pool, advisoryID, packageName, affectedVersions string) {
	t.Helper()
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		INSERT INTO composer_advisory (advisory_id, title, cve, severity, is_visible)
		VALUES ($1, $1 || ' title', $1, 'high', true)
		ON CONFLICT (advisory_id) DO NOTHING
	`, advisoryID)
	require.NoError(t, err, "seed advisory")

	_, err = pool.Exec(ctx, `
		INSERT INTO composer_advisory_package (advisory_id, package_name, packagist_advisory_id, affected_versions)
		VALUES ($1, $2, 'PKSA-' || $1, $3)
		ON CONFLICT (advisory_id, package_name) DO UPDATE SET affected_versions = EXCLUDED.affected_versions
	`, advisoryID, packageName, affectedVersions)
	require.NoError(t, err, "seed advisory package")
}

// A suppressed advisory must stop escalating the environment status. The check
// itself is still emitted, matching how legacy ignores behave.
func TestSuppressedCheckIDsCoversShopAndEnvironmentScope(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()

	seedAdvisory(t, pool, "CVE-2026-900", "shopware/core", "<7.0.0")

	// No suppression yet.
	assert.Empty(t, h.suppressedCheckIDs(ctx, env.ID))

	var orgID string
	var shopID int32
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT organization_id, shop_id FROM environment WHERE id=$1`, env.ID).Scan(&orgID, &shopID))

	// Shop-wide suppression (environment_id NULL) covers this environment.
	_, err := pool.Exec(ctx, `
		INSERT INTO advisory_suppression (organization_id, shop_id, advisory_id, reason, created_at)
		VALUES ($1, $2, 'CVE-2026-900', 'accepted', NOW())`, orgID, shopID)
	require.NoError(t, err)

	ids := h.suppressedCheckIDs(ctx, env.ID)
	require.Len(t, ids, 1)
	// The id must match what the checker builds, or the union silences nothing.
	assert.Equal(t, "security.CVE-2026-900", ids[0])

	// An expired suppression stops covering it.
	_, err = pool.Exec(ctx, `UPDATE advisory_suppression SET expires_at = NOW() - interval '1 hour'`)
	require.NoError(t, err)
	assert.Empty(t, h.suppressedCheckIDs(ctx, env.ID))

	// A revoked suppression likewise.
	_, err = pool.Exec(ctx, `UPDATE advisory_suppression SET expires_at = NULL, revoked_at = NOW()`)
	require.NoError(t, err)
	assert.Empty(t, h.suppressedCheckIDs(ctx, env.ID))
}
