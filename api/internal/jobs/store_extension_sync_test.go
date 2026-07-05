package jobs

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockExtension describes what the mock store returns for one technical name.
type mockExtension struct {
	// latestByShopwareVersion maps the shopwareVersion query param to the Version
	// field the store reports. A version absent from the map means the store
	// returns no plugin for it (no compatible release) — modelling the endpoint's
	// version scoping.
	latestByShopwareVersion map[string]string
	// changelogVersions is the full changelog list, newest first.
	changelogVersions []string
}

// mockStoreServer serves the pluginsByName endpoint for a configurable set of
// extensions whose availability and reported latest version depend on the
// requested Shopware version, and counts requests.
type mockStoreServer struct {
	*httptest.Server
	requests   atomic.Int64
	extensions map[string]*mockExtension
}

func newMockStoreServer(t *testing.T) *mockStoreServer {
	t.Helper()
	m := &mockStoreServer{
		extensions: map[string]*mockExtension{
			"FroshTools": {
				latestByShopwareVersion: map[string]string{"6.5.0.0": "1.1.0", "6.6.0.0": "1.2.0"},
				changelogVersions:       []string{"1.2.0", "1.1.0", "1.0.0"},
			},
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pluginStore/pluginsByName", func(w http.ResponseWriter, r *http.Request) {
		m.requests.Add(1)
		locale := r.URL.Query().Get("locale")
		swv := r.URL.Query().Get("shopwareVersion")

		var plugins []map[string]any
		for _, name := range r.URL.Query()["technicalNames[]"] {
			ext, ok := m.extensions[name]
			if !ok {
				continue
			}
			latest, ok := ext.latestByShopwareVersion[swv]
			if !ok {
				continue // no release compatible with this Shopware version
			}
			changelog := make([]map[string]any, 0, len(ext.changelogVersions))
			for _, v := range ext.changelogVersions {
				changelog = append(changelog, map[string]any{
					"version":      v,
					"text":         locale + " changelog " + v,
					"creationDate": map[string]string{"date": "2023-01-01 00:00:00.000000"},
				})
			}
			plugins = append(plugins, map[string]any{
				"id":            42,
				"name":          name,
				"label":         locale + " " + name,
				"description":   locale + " description",
				"version":       latest,
				"ratingAverage": 4.0,
				"link":          "http://store.shopware.com:80/" + name,
				"iconPath":      "https://store.shopware.com/icon.png",
				"producer":      map[string]string{"name": "FriendsOfShopware", "website": "https://friendsofshopware.com"},
				"infos":         []map[string]string{{"shortDescription": locale + " short"}},
				"pictures": []map[string]any{
					{"remoteLink": "https://img.example.com/1.png", "preview": true, "priority": 1},
					{"remoteLink": "https://img.example.com/2.png", "preview": false, "priority": 2},
				},
				"changelog": changelog,
			})
		}

		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(plugins), "encode mock response")
	})

	m.Server = httptest.NewServer(mux)
	t.Cleanup(m.Close)
	return m
}

// setupSyncTest returns a sync handler wired to a mock store server, with two
// environments on different Shopware versions (6.5.0.0 and 6.6.0.0).
func setupSyncTest(t *testing.T) (*StoreExtensionSyncHandler, *pgxpool.Pool, *mockStoreServer) {
	t.Helper()

	pool := testdb.Setup(t)
	q := queries.New(pool)
	store := newMockStoreServer(t)
	h := NewStoreExtensionSyncHandler(pool, q, &config.Config{ShopwareAPIURL: store.URL})

	seedEnvironment(t, pool, "Production", "6.5.0.0")
	seedEnvironment(t, pool, "Staging", "6.6.0.0")

	return h, pool, store
}

func TestStoreExtensionSyncPersistsCatalog(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false))

	// en + de for each in-use Shopware version (6.5.0.0 and 6.6.0.0).
	assert.Equal(t, int64(4), store.requests.Load(), "store requests")

	for table, want := range map[string]int{
		"store_extension":                     1,
		"store_extension_translation":         2, // en + de
		"store_extension_version":             3,
		"store_extension_version_translation": 6, // 3 versions x en+de
		"store_extension_image":               2,
		"store_extension_compatibility":       2, // 6.5.0.0 + 6.6.0.0
		"store_extension_sync":                2, // FroshTools + CustomPlugin (miss)
	} {
		assert.Equalf(t, want, countRows(t, pool, table), "row count for %s", table)
	}

	// Compatible latest per Shopware version, and the uncapped global latest.
	for swv, want := range map[string]string{"6.5.0.0": "1.1.0", "6.6.0.0": "1.2.0"} {
		var latest *string
		require.NoErrorf(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = $1`, swv).Scan(&latest), "read compatibility %s", swv)
		require.NotNilf(t, latest, "compatible latest for %s", swv)
		assert.Equalf(t, want, *latest, "compatible latest for %s", swv)
	}
	var globalLatest *string
	var storeLink *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version, store_link FROM store_extension WHERE name = 'FroshTools'`).Scan(&globalLatest, &storeLink), "read catalog row")
	require.NotNil(t, globalLatest, "global latest")
	assert.Equal(t, "1.2.0", *globalLatest, "global latest")
	require.NotNil(t, storeLink, "store link")
	assert.Equal(t, "https://store.shopware.com/FroshTools", *storeLink, "store link normalized")

	// The miss (CustomPlugin) must not create a catalog row but is recorded.
	var missSynced int
	require.NoError(t, pool.QueryRow(ctx, `SELECT COUNT(*) FROM store_extension_sync WHERE extension_name = 'CustomPlugin'`).Scan(&missSynced), "read sync state")
	assert.Equal(t, 1, missSynced, "store miss recorded in sync bookkeeping")
}

// TestStoreExtensionSyncPersistsOlderOnlyExtension is a regression test for the
// mixed-version scoping bug: an extension whose latest release is only
// compatible with an older environment's Shopware version must still be
// persisted, even when a newer environment exists and a sync is requested from
// the older one. The store's pluginsByName endpoint omits the plugin for the
// newer version, so building the catalog from a fixed "newest" version would
// drop it while still marking it synced.
func TestStoreExtensionSyncPersistsOlderOnlyExtension(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	// LegacyExt only has a release compatible with 6.5.0.0, not the newer
	// 6.6.0.0 environment that also exists.
	store.extensions["LegacyExt"] = &mockExtension{
		latestByShopwareVersion: map[string]string{"6.5.0.0": "3.4.0"},
		changelogVersions:       []string{"3.4.0", "3.3.0"},
	}

	require.NoError(t, h.SyncNames(ctx, []string{"LegacyExt"}, "6.5.0.0", false), "sync")

	// The catalog and its changelog history are persisted despite the plugin
	// being absent from the 6.6.0.0 probe.
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "catalog row")
	assert.Equal(t, 2, countRows(t, pool, "store_extension_version"), "version rows")

	var globalLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension WHERE name = 'LegacyExt'`).Scan(&globalLatest), "read catalog row")
	require.NotNil(t, globalLatest, "global latest")
	assert.Equal(t, "3.4.0", *globalLatest)

	// 6.5.0.0 gets a compatible latest; 6.6.0.0 records that it was checked and
	// has no compatible release (NULL), so it is not re-probed every scrape.
	var latest65 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'LegacyExt' AND shopware_version = '6.5.0.0'`).Scan(&latest65), "read 6.5 compat")
	require.NotNil(t, latest65, "6.5.0.0 compatible latest")
	assert.Equal(t, "3.4.0", *latest65)

	var latest66 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'LegacyExt' AND shopware_version = '6.6.0.0'`).Scan(&latest66), "read 6.6 compat")
	assert.Nil(t, latest66, "6.6.0.0 has no compatible release")

	// It is marked synced, so the classification is stable rather than retried.
	assert.Equal(t, 1, countRows(t, pool, "store_extension_sync"))
}

// TestStoreExtensionSyncFreshIsNoop: a second sync right after the first must
// be answered from the bookkeeping without any HTTP request or row write. This
// is what collapses concurrent dispatches from many environments sharing the
// same extensions into a single hourly fetch.
func TestStoreExtensionSyncFreshIsNoop(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()
	before := snapshotRowVersions(t, pool)

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "second sync")

	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "fresh re-sync must make no extra store requests")
	after := snapshotRowVersions(t, pool)
	for key, xmin := range before {
		assert.Equalf(t, xmin, after[key], "row %s was rewritten by a fresh re-sync", key)
	}
}

// TestStoreExtensionSyncForceUnchangedRewritesNothing: a forced sync re-fetches
// from the store, but identical data must not produce a single new row version
// thanks to the change guards.
func TestStoreExtensionSyncForceUnchangedRewritesNothing(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	before := snapshotRowVersions(t, pool)
	requestsAfterFirst := store.requests.Load()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", true), "forced sync")

	assert.Greater(t, store.requests.Load(), requestsAfterFirst, "forced sync must hit the store")
	after := snapshotRowVersions(t, pool)
	for key, xmin := range before {
		// The catalog row itself is rewritten (last_refreshed_at), everything
		// else must be untouched.
		if key == "store_extension/FroshTools" {
			continue
		}
		assert.Equalf(t, xmin, after[key], "row %s was rewritten by a forced sync of unchanged data", key)
	}
}

// TestStoreExtensionSyncPicksUpNewRelease: once the bookkeeping is stale, a
// sync fetches again and only writes the new version rows.
func TestStoreExtensionSyncPicksUpNewRelease(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	before := snapshotRowVersions(t, pool)

	// A new release appears and the bookkeeping ages out.
	frosh := store.extensions["FroshTools"]
	frosh.latestByShopwareVersion["6.6.0.0"] = "1.3.0"
	frosh.changelogVersions = append([]string{"1.3.0"}, frosh.changelogVersions...)
	_, err := pool.Exec(ctx, `UPDATE store_extension_sync SET last_synced_at = NOW() - INTERVAL '2 days'`)
	require.NoError(t, err, "age sync state")

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "second sync")

	var globalLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension WHERE name = 'FroshTools'`).Scan(&globalLatest), "read catalog row")
	require.NotNil(t, globalLatest, "global latest after release")
	assert.Equal(t, "1.3.0", *globalLatest, "global latest after release")
	assert.Equal(t, 4, countRows(t, pool, "store_extension_version"), "version rows")

	var latest65 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.5.0.0'`).Scan(&latest65), "read compatibility")
	require.NotNil(t, latest65, "6.5.0.0 compatible latest")
	assert.Equal(t, "1.1.0", *latest65, "6.5.0.0 compatible latest must be unchanged")

	// Existing version rows and their changelogs must not have been rewritten.
	after := snapshotRowVersions(t, pool)
	for _, key := range []string{
		"store_extension_version/FroshTools:1.0.0",
		"store_extension_version/FroshTools:1.1.0",
		"store_extension_version/FroshTools:1.2.0",
	} {
		require.NotEmptyf(t, before[key], "row %s missing from first snapshot", key)
		assert.Equalf(t, before[key], after[key], "unchanged row %s was rewritten during release sync", key)
	}
}

// TestNamesNeedingSyncCompatibilityGap: fresh bookkeeping does not suppress a
// sync when a store-known extension lacks a compatibility entry for a new
// Shopware version (e.g. right after an environment upgraded).
func TestNamesNeedingSyncCompatibilityGap(t *testing.T) {
	h, _, _ := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "sync")

	// Known version: nothing to do.
	needing, err := h.namesNeedingSync(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0")
	require.NoError(t, err, "namesNeedingSync")
	assert.Empty(t, needing, "needing sync for known version")

	// New version: only the store-known extension needs a compatibility probe;
	// the store miss stays quiet until its bookkeeping ages out.
	needing, err = h.namesNeedingSync(ctx, []string{"FroshTools", "CustomPlugin"}, "6.7.0.0")
	require.NoError(t, err, "namesNeedingSync")
	assert.Equal(t, []string{"FroshTools"}, needing, "needing sync for new version")
}

// TestOldDataCleanupRetainsCatalog verifies the catalog itself — including its
// full changelog history — is kept even once no environment links it, while
// the internal bookkeeping (orphaned sync state, compatibility rows for unused
// Shopware versions) is pruned.
func TestOldDataCleanupRetainsCatalog(t *testing.T) {
	h, pool, _ := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "sync")

	// Drop every environment link, so the catalog is fully orphaned, and age the
	// bookkeeping past the 7-day window. The 6.5.0.0 and 6.6.0.0 environments
	// are removed, so their compatibility rows become unused.
	for _, stmt := range []string{
		`DELETE FROM environment_store_extension`,
		`DELETE FROM environment_extension`,
		`DELETE FROM environment`,
		`UPDATE store_extension SET last_refreshed_at = NOW() - INTERVAL '8 days'`,
		`UPDATE store_extension_sync SET last_synced_at = NOW() - INTERVAL '8 days'`,
	} {
		_, err := pool.Exec(ctx, stmt)
		require.NoErrorf(t, err, "prepare orphaned state (%q)", stmt)
	}

	cleanup := NewCleanupHandler(queries.New(pool))
	require.NoError(t, cleanup.HandleOldDataCleanup(ctx, OldDataCleanup{}), "cleanup")

	// The catalog and its changelog history survive.
	for table, want := range map[string]int{
		"store_extension":                     1,
		"store_extension_translation":         2,
		"store_extension_version":             3,
		"store_extension_version_translation": 6,
		"store_extension_image":               2,
	} {
		assert.Equalf(t, want, countRows(t, pool, table), "%s must be retained after cleanup", table)
	}

	// Orphaned bookkeeping is pruned.
	assert.Equal(t, 0, countRows(t, pool, "store_extension_sync"), "orphaned sync state pruned")
	assert.Equal(t, 0, countRows(t, pool, "store_extension_compatibility"), "compatibility rows for unused versions pruned")
}
