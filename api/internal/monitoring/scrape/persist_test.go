package scrape

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	catalogsync "github.com/friendsofshopware/shopmon/api/internal/catalog/sync"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPersistTest returns a scrape handler backed by a clean test database and
// a seeded environment.
func setupPersistTest(t *testing.T) (*Service, *pgxpool.Pool, queries.GetAllEnvironmentsRow) {
	t.Helper()

	pool := testdb.Setup(t)
	q := queries.New(pool)
	cfg := &config.Config{}
	h := NewService(pool, q, cfg, nil, catalogsync.NewService(pool, q, cfg), nil)

	envID := seedEnvironment(t, pool, "Production", "6.5.0.0")
	return h, pool, reloadEnv(t, h, envID)
}

// seedEnvironment creates the organization/shop scaffolding (once) and one
// environment, returning its id.
func seedEnvironment(t *testing.T, pool *pgxpool.Pool, name, shopwareVersion string) int32 {
	t.Helper()
	ctx := context.Background()

	_, err := pool.Exec(ctx, `INSERT INTO organization (id, name, slug, created_at) VALUES ('org-1', 'Test Org', 'test-org', NOW()) ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err, "seed organization")

	var shopID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO shop (organization_id, name, created_at, updated_at) VALUES ('org-1', 'Test Shop ' || $1::text, NOW(), NOW()) RETURNING id
	`, name).Scan(&shopID)
	require.NoError(t, err, "seed shop")

	var envID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO environment (organization_id, shop_id, name, url, client_id, client_secret, shopware_version, environment_token, created_at)
		VALUES ('org-1', $1, $2, 'https://shop.example.com', 'client', 'secret', $3, 'token', NOW())
		RETURNING id
	`, shopID, name, shopwareVersion).Scan(&envID)
	require.NoError(t, err, "seed environment")

	return envID
}

func reloadEnv(t *testing.T, h *Service, envID int32) queries.GetAllEnvironmentsRow {
	t.Helper()
	row, err := h.queries.GetEnvironmentForScrape(context.Background(), envID)
	require.NoError(t, err, "reload environment")
	return queries.GetAllEnvironmentsRow(row)
}

// seedStoreCatalog populates the shared catalog with one extension (FroshTools)
// including translations, versions, changelogs, images, a compatibility entry
// for Shopware 6.5.0.0 and fresh sync bookkeeping. The persist step only links
// against the catalog; it is maintained by the StoreExtensionSync job.
func seedStoreCatalog(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	statements := []string{
		`INSERT INTO store_extension (name, store_id, icon_url, producer_name, rating_average, store_link, latest_version)
		 VALUES ('FroshTools', 42, 'https://store.shopware.com/icon.png', 'FriendsOfShopware', 4, 'https://store.shopware.com/frosh-tools', '1.2.0')`,
		`INSERT INTO store_extension_translation (extension_name, language, label, short_description, description)
		 VALUES ('FroshTools', 'en', 'Frosh Tools', 'Dev tools', 'A long description'),
		        ('FroshTools', 'de', 'Frosh Werkzeuge', 'Entwickler-Werkzeuge', 'Eine lange Beschreibung')`,
		`INSERT INTO store_extension_version (extension_name, version, released_at)
		 VALUES ('FroshTools', '1.0.0', '2023-01-01T00:00:00Z'), ('FroshTools', '1.2.0', '2023-06-01T00:00:00Z')`,
		`INSERT INTO store_extension_version_translation (extension_version_id, language, changelog)
		 SELECT id, 'en', 'changelog en ' || version FROM store_extension_version WHERE extension_name = 'FroshTools'`,
		`INSERT INTO store_extension_image (extension_name, url, preview, priority)
		 VALUES ('FroshTools', 'https://img.example.com/1.png', true, 1), ('FroshTools', 'https://img.example.com/2.png', false, 2)`,
		`INSERT INTO store_extension_compatibility (extension_name, shopware_version, latest_version)
		 VALUES ('FroshTools', '6.5.0.0', '1.2.0')`,
		`INSERT INTO store_extension_sync (extension_name) VALUES ('FroshTools'), ('CustomPlugin')`,
	}
	for _, stmt := range statements {
		_, err := pool.Exec(ctx, stmt)
		require.NoError(t, err, "seed store catalog")
	}
}

// baseScrapeResult builds a representative scrape result: one store-known
// extension (linked against the seeded catalog), one unknown extension, plus
// tasks, queues, cache info and checks.
func baseScrapeResult() scrapeResult {
	return scrapeResult{
		status:          "yellow",
		shopwareVersion: "6.5.0.0",
		extensions: []extensionEntry{
			{
				Name: "FroshTools", Label: "Frosh Tools", Active: true, Version: "1.0.0",
				LatestVersion: new("1.2.0"), Installed: true, InstalledAt: new("2024-01-01T00:00:00Z"),
				isStore: true,
			},
			{
				Name: "CustomPlugin", Label: "Custom Plugin", Active: true, Version: "2.0.0",
				Installed: true, InstalledAt: new("2024-02-01T00:00:00Z"),
			},
		},
		scheduledTasks: []shopwareScheduledTask{
			{ID: "task-1", Name: "log_entry.cleanup", Status: "scheduled", RunInterval: 86400,
				LastExecutionTime: new("2024-01-01T00:00:00Z"), NextExecutionTime: new("2999-01-01T00:00:00Z")},
		},
		queueEntries: []shopwareQueueEntry{
			{Name: "async", Size: json.Number("42")},
		},
		cacheInfo: shopwareCacheInfo{Environment: "prod", HttpCache: true, CacheAdapter: "redis"},
		checks: []checker.Check{
			{ID: "check.a", Level: checker.StatusGreen, MessageKey: "check.a", Source: "env"},
			{ID: "check.b", Level: checker.StatusYellow, MessageKey: "check.b", Source: "worker", Link: "https://docs.example.com"},
		},
		favicon: new("https://shop.example.com/favicon.ico"),
	}
}

// snapshotRowVersions captures the xmin (row version) of every row written by
// the scrape persist or the catalog sync, keyed by table and natural key. Two
// equal snapshots prove the rows between them were not rewritten, i.e. produced
// no WAL.
func snapshotRowVersions(t *testing.T, pool *pgxpool.Pool) map[string]string {
	t.Helper()

	tableKeys := map[string]string{
		"environment_extension":               `SELECT environment_id || ':' || name, xmin::text FROM environment_extension`,
		"environment_store_extension":         `SELECT environment_id || ':' || extension_name, xmin::text FROM environment_store_extension`,
		"environment_scheduled_task":          `SELECT environment_id || ':' || task_id, xmin::text FROM environment_scheduled_task`,
		"environment_queue":                   `SELECT environment_id || ':' || name, xmin::text FROM environment_queue`,
		"environment_cache":                   `SELECT environment_id::text, xmin::text FROM environment_cache`,
		"environment_check":                   `SELECT environment_id || ':' || check_id, xmin::text FROM environment_check`,
		"store_extension":                     `SELECT name, xmin::text FROM store_extension`,
		"store_extension_translation":         `SELECT extension_name || ':' || language, xmin::text FROM store_extension_translation`,
		"store_extension_version":             `SELECT extension_name || ':' || version, xmin::text FROM store_extension_version`,
		"store_extension_version_translation": `SELECT extension_version_id || ':' || language, xmin::text FROM store_extension_version_translation`,
		"store_extension_image":               `SELECT extension_name || ':' || url, xmin::text FROM store_extension_image`,
		"store_extension_compatibility":       `SELECT extension_name || ':' || shopware_version, xmin::text FROM store_extension_compatibility`,
	}

	snapshot := make(map[string]string)
	for table, query := range tableKeys {
		rows, err := pool.Query(context.Background(), query)
		require.NoError(t, err, "snapshot %s", table)
		for rows.Next() {
			var key, xmin string
			require.NoError(t, rows.Scan(&key, &xmin), "scan %s", table)
			snapshot[table+"/"+key] = xmin
		}
		rows.Close()
		require.NoError(t, rows.Err(), "iterate %s", table)
	}
	return snapshot
}

func countRows(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var n int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&n), "count %s", table)
	return n
}

func TestPersistScrapeResultWritesAllData(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	require.NoError(t, h.persistScrapeResult(ctx, env, baseScrapeResult()))

	for table, want := range map[string]int{
		"environment_extension":       1,
		"environment_store_extension": 1,
		"environment_scheduled_task":  1,
		"environment_queue":           1,
		"environment_cache":           1,
		"environment_check":           2,
		"environment_changelog":       0, // same shopware version, no diff
	} {
		assert.Equalf(t, want, countRows(t, pool, table), "row count for %s", table)
	}

	var status string
	var favicon *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT status, favicon FROM environment WHERE id = $1`, env.ID).Scan(&status, &favicon))
	assert.Equal(t, "yellow", status, "environment status")
	require.NotNil(t, favicon, "environment favicon")
	assert.Equal(t, "https://shop.example.com/favicon.ico", *favicon)

	var linkLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM environment_store_extension WHERE environment_id = $1 AND extension_name = 'FroshTools'`, env.ID).Scan(&linkLatest))
	require.NotNil(t, linkLatest, "link latest_version")
	assert.Equal(t, "1.2.0", *linkLatest)
}

// TestPersistScrapeResultIdenticalRunRewritesNothing is the core WAL regression
// test: persisting the exact same scrape result a second time must not produce
// a single new row version in any scrape-managed table, and must never touch
// the shared catalog.
func TestPersistScrapeResultIdenticalRunRewritesNothing(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	res := baseScrapeResult()
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "first persist")

	before := snapshotRowVersions(t, pool)
	require.NotEmpty(t, before, "snapshot is empty, seeding failed")

	env = reloadEnv(t, h, env.ID)
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "second persist")

	after := snapshotRowVersions(t, pool)
	for key, xmin := range before {
		got, ok := after[key]
		if assert.Truef(t, ok, "row %s disappeared on identical re-persist", key) {
			assert.Equalf(t, xmin, got, "row %s was rewritten on identical re-persist", key)
		}
	}
	for key := range after {
		assert.Containsf(t, before, key, "row %s appeared on identical re-persist", key)
	}

	assert.Equal(t, 0, countRows(t, pool, "environment_changelog"))
}

func TestPersistScrapeResultUpdatesChangedRowsOnly(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	require.NoError(t, h.persistScrapeResult(ctx, env, baseScrapeResult()), "first persist")
	before := snapshotRowVersions(t, pool)

	res := baseScrapeResult()
	res.queueEntries[0].Size = json.Number("100")
	res.checks[1].Level = checker.StatusRed

	env = reloadEnv(t, h, env.ID)
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "second persist")
	after := snapshotRowVersions(t, pool)

	queueKey := fmt.Sprintf("environment_queue/%d:async", env.ID)
	checkKey := fmt.Sprintf("environment_check/%d:check.b", env.ID)
	changed := map[string]bool{queueKey: true, checkKey: true}

	for key, xmin := range before {
		if changed[key] {
			assert.NotEqualf(t, xmin, after[key], "row %s should have been rewritten but was not", key)
			continue
		}
		assert.Equalf(t, xmin, after[key], "unchanged row %s was rewritten", key)
	}

	var size int
	require.NoError(t, pool.QueryRow(ctx, `SELECT size FROM environment_queue WHERE environment_id = $1 AND name = 'async'`, env.ID).Scan(&size))
	assert.Equal(t, 100, size, "queue size")

	var level string
	require.NoError(t, pool.QueryRow(ctx, `SELECT level FROM environment_check WHERE environment_id = $1 AND check_id = 'check.b'`, env.ID).Scan(&level))
	assert.Equal(t, "red", level, "check level")
}

func TestPersistScrapeResultPrunesStaleRows(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	require.NoError(t, h.persistScrapeResult(ctx, env, baseScrapeResult()), "first persist")
	before := snapshotRowVersions(t, pool)

	res := baseScrapeResult()
	res.checks = res.checks[1:] // drop check.a, keep check.b

	env = reloadEnv(t, h, env.ID)
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "second persist")

	require.Equal(t, 1, countRows(t, pool, "environment_check"))
	var checkID string
	require.NoError(t, pool.QueryRow(ctx, `SELECT check_id FROM environment_check WHERE environment_id = $1`, env.ID).Scan(&checkID))
	assert.Equal(t, "check.b", checkID, "remaining check")

	// The surviving check must not have been rewritten by the prune.
	after := snapshotRowVersions(t, pool)
	checkKey := fmt.Sprintf("environment_check/%d:check.b", env.ID)
	assert.Equal(t, before[checkKey], after[checkKey], "surviving check row was rewritten")

	// No checks reported at all removes every row.
	res.checks = nil
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "third persist")
	assert.Equal(t, 0, countRows(t, pool, "environment_check"), "checks not cleared on empty result")
}

// TestPersistScrapeResultEmptyExtensionsKeepsData guards the existing safety
// behavior: a scrape that returned no extensions (likely a fetch error) must
// not wipe previously stored extensions.
func TestPersistScrapeResultEmptyExtensionsKeepsData(t *testing.T) {
	h, pool, env := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	require.NoError(t, h.persistScrapeResult(ctx, env, baseScrapeResult()), "first persist")

	res := baseScrapeResult()
	res.extensions = nil

	env = reloadEnv(t, h, env.ID)
	require.NoError(t, h.persistScrapeResult(ctx, env, res), "second persist")

	assert.Equal(t, 1, countRows(t, pool, "environment_store_extension"))
	assert.Equal(t, 1, countRows(t, pool, "environment_extension"))
}

// TestResolveExtensionsFromCatalog verifies classification and latest-version
// resolution against the local catalog: store membership sets isStore, the
// compatibility table wins over the shop-reported upgrade version, a NULL
// compatibility entry keeps the shop value, and a missing entry falls back to
// the prior link value.
func TestResolveExtensionsFromCatalog(t *testing.T) {
	h, pool, _ := setupPersistTest(t)
	ctx := context.Background()
	seedStoreCatalog(t, pool)

	_, err := pool.Exec(ctx, `
		INSERT INTO store_extension (name, latest_version) VALUES ('AcmeNull', '3.0.0'), ('AcmePending', '4.0.0');
		INSERT INTO store_extension_compatibility (extension_name, shopware_version, latest_version)
		VALUES ('AcmeNull', '6.5.0.0', NULL);
	`)
	require.NoError(t, err, "seed extra catalog rows")

	extensions := []extensionEntry{
		{Name: "FroshTools", Version: "1.0.0", LatestVersion: new("1.1.0")}, // compat row 1.2.0 must win
		{Name: "AcmeNull", Version: "2.0.0", LatestVersion: new("2.5.0")},   // NULL compat row keeps shop value
		{Name: "AcmePending", Version: "3.0.0"},                             // no compat row: prior link value
		{Name: "CustomPlugin", Version: "2.0.0"},                            // not in catalog
	}
	oldExtensions := []existingExtension{
		{Name: "AcmePending", Version: "3.0.0", IsStore: true, LatestVersion: new("3.9.0")},
	}

	catalog, err := h.resolveExtensionsFromCatalog(ctx, extensions, oldExtensions, "6.5.0.0")
	require.NoError(t, err, "resolve")

	// The returned state carries the reads the dispatch path reuses instead of
	// re-querying: membership and the compatibility rows that exist (AcmeNull's
	// NULL row counts as known).
	assert.Equal(t, map[string]bool{"FroshTools": true, "AcmeNull": true, "AcmePending": true}, catalog.member, "member")
	assert.Contains(t, catalog.compatKnown, "FroshTools")
	assert.Contains(t, catalog.compatKnown, "AcmeNull")
	assert.NotContains(t, catalog.compatKnown, "AcmePending", "no compat row yet")

	assertExt := func(i int, wantStore bool, wantLatest *string) {
		t.Helper()
		ext := extensions[i]
		assert.Equalf(t, wantStore, ext.isStore, "%s: isStore", ext.Name)
		assert.Equalf(t, wantLatest, ext.LatestVersion, "%s: latest version", ext.Name)
	}
	assertExt(0, true, new("1.2.0"))
	assertExt(1, true, new("2.5.0"))
	assertExt(2, true, new("3.9.0"))
	assertExt(3, false, nil)
}
