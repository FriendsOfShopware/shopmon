package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/maintenance"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/metric/metricdata"
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

	mu sync.Mutex
	// rateLimitedVersions, when set, causes pluginsByName for those Shopware
	// versions to respond with HTTP 429.
	rateLimitedVersions map[string]bool
	// serverErrorVersions, when set, causes pluginsByName for those Shopware
	// versions to respond with HTTP 500 (non-retryable, not a rate limit).
	serverErrorVersions map[string]bool
	// seen records locale@shopwareVersion hits in order.
	seen []string
	// inFlight tracks concurrent handlers; maxInFlight is the high-water mark.
	inFlight    int
	maxInFlight int
	// handlerDelay, when > 0, holds each handler briefly so overlapping
	// concurrent calls raise maxInFlight (used by the serialization test).
	handlerDelay time.Duration
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
		rateLimitedVersions: map[string]bool{},
		serverErrorVersions: map[string]bool{},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/pluginStore/pluginsByName", func(w http.ResponseWriter, r *http.Request) {
		m.requests.Add(1)
		locale := r.URL.Query().Get("locale")
		swv := r.URL.Query().Get("shopwareVersion")

		m.mu.Lock()
		m.seen = append(m.seen, locale+"@"+swv)
		m.inFlight++
		if m.inFlight > m.maxInFlight {
			m.maxInFlight = m.inFlight
		}
		rateLimited := m.rateLimitedVersions[swv]
		serverError := m.serverErrorVersions[swv]
		delay := m.handlerDelay
		m.mu.Unlock()

		defer func() {
			m.mu.Lock()
			m.inFlight--
			m.mu.Unlock()
		}()

		if delay > 0 {
			time.Sleep(delay)
		}

		if rateLimited {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}
		if serverError {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

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

func (m *mockStoreServer) seenLocales() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]string, len(m.seen))
	copy(out, m.seen)
	return out
}

func (m *mockStoreServer) maxConcurrent() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.maxInFlight
}

// fastStoreClient returns a client that fails 429s on the first attempt (no
// real backoff sleep) so sync tests stay fast.
func fastStoreClient(baseURL string) *shopwareaccount.Client {
	client := shopwareaccount.NewClient(baseURL, nil)
	client.ConfigureRetry(1, func(context.Context, time.Duration) error { return nil })
	return client
}

// setupSyncTest returns a sync handler wired to a mock store server, with two
// environments on different Shopware versions (6.5.0.0 and 6.6.0.0).
func setupSyncTest(t *testing.T) (*Service, *pgxpool.Pool, *mockStoreServer) {
	t.Helper()

	pool := testdb.Setup(t)
	q := queries.New(pool)
	store := newMockStoreServer(t)
	h := NewService(pool, q, &config.Config{ShopwareAPIURL: store.URL})

	seedEnvironment(t, pool, "Production", "6.5.0.0")
	seedEnvironment(t, pool, "Staging", "6.6.0.0")

	return h, pool, store
}

func seedEnvironment(t *testing.T, pool *pgxpool.Pool, name, shopwareVersion string) int32 {
	t.Helper()
	ctx := context.Background()
	_, err := pool.Exec(ctx, `INSERT INTO organization (id, name, slug, created_at) VALUES ('org-1', 'Test Org', 'test-org', NOW()) ON CONFLICT (id) DO NOTHING`)
	require.NoError(t, err, "seed organization")

	var shopID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO shop (organization_id, name, created_at, updated_at)
		VALUES ('org-1', 'Test Shop ' || $1::text, NOW(), NOW())
		RETURNING id
	`, name).Scan(&shopID)
	require.NoError(t, err, "seed shop")

	var environmentID int32
	err = pool.QueryRow(ctx, `
		INSERT INTO environment (organization_id, shop_id, name, url, client_id, client_secret, shopware_version, environment_token, created_at)
		VALUES ('org-1', $1, $2, 'https://shop.example.com', 'client', 'secret', $3, 'token', NOW())
		RETURNING id
	`, shopID, name, shopwareVersion).Scan(&environmentID)
	require.NoError(t, err, "seed environment")
	return environmentID
}

func snapshotRowVersions(t *testing.T, pool *pgxpool.Pool) map[string]string {
	t.Helper()
	queriesByTable := map[string]string{
		"store_extension":                     `SELECT name, xmin::text FROM store_extension`,
		"store_extension_translation":         `SELECT extension_name || ':' || language, xmin::text FROM store_extension_translation`,
		"store_extension_version":             `SELECT extension_name || ':' || version, xmin::text FROM store_extension_version`,
		"store_extension_version_translation": `SELECT extension_version_id || ':' || language, xmin::text FROM store_extension_version_translation`,
		"store_extension_image":               `SELECT extension_name || ':' || url, xmin::text FROM store_extension_image`,
		"store_extension_compatibility":       `SELECT extension_name || ':' || shopware_version, xmin::text FROM store_extension_compatibility`,
	}
	result := make(map[string]string)
	for table, query := range queriesByTable {
		rows, err := pool.Query(context.Background(), query)
		require.NoError(t, err, "snapshot %s", table)
		for rows.Next() {
			var key, version string
			require.NoError(t, rows.Scan(&key, &version), "scan %s", table)
			result[table+"/"+key] = version
		}
		rows.Close()
		require.NoError(t, rows.Err(), "iterate %s", table)
	}
	return result
}

func withStoreSyncMetrics(t *testing.T) func() map[string]int64 {
	t.Helper()
	reader := sdkmetric.NewManualReader()
	mp := sdkmetric.NewMeterProvider(sdkmetric.WithReader(reader))
	prev := otel.GetMeterProvider()
	otel.SetMeterProvider(mp)
	metrics.Register()
	t.Cleanup(func() {
		otel.SetMeterProvider(prev)
		_ = mp.Shutdown(context.Background())
	})
	return func() map[string]int64 {
		t.Helper()
		var rm metricdata.ResourceMetrics
		require.NoError(t, reader.Collect(context.Background(), &rm))
		out := make(map[string]int64)
		for _, sm := range rm.ScopeMetrics {
			for _, m := range sm.Metrics {
				sum, ok := m.Data.(metricdata.Sum[int64])
				if !ok {
					continue
				}
				for _, dp := range sum.DataPoints {
					key := m.Name
					for _, attr := range dp.Attributes.ToSlice() {
						key += "|" + string(attr.Key) + "=" + attr.Value.AsString()
					}
					out[key] = dp.Value
				}
			}
		}
		return out
	}
}

func countRows(t *testing.T, pool *pgxpool.Pool, table string) int {
	t.Helper()
	var count int
	require.NoError(t, pool.QueryRow(context.Background(), "SELECT COUNT(*) FROM "+table).Scan(&count), "count %s", table)
	return count
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

// TestStoreExtensionSyncSerializesLocaleProbes: en and de for a version must
// not run concurrently, otherwise catalog sync doubles store API pressure.
func TestStoreExtensionSyncSerializesLocaleProbes(t *testing.T) {
	h, _, store := setupSyncTest(t)
	store.handlerDelay = 30 * time.Millisecond

	require.NoError(t, h.SyncNames(context.Background(), []string{"FroshTools"}, "6.5.0.0", false))

	assert.Equal(t, 1, store.maxConcurrent(), "locale/version probes must be serial")
	assert.Equal(t, []string{
		"en_GB@6.6.0.0", "de_DE@6.6.0.0",
		"en_GB@6.5.0.0", "de_DE@6.5.0.0",
	}, store.seenLocales(), "versions newest-first, locales en then de")
}

// TestStoreExtensionSyncAbortsRemainingVersionsOn429: once the store rate-limits
// a version probe, SyncNames must stop probing further versions instead of
// stomping through the rest of the list, leave sync bookkeeping untouched so a
// later scheduled sync can continue, persist any versions already probed, and
// return nil so the queue acks instead of nacking into a still-limited API.
func TestStoreExtensionSyncAbortsRemainingVersionsOn429(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()
	collect := withStoreSyncMetrics(t)

	// Newest version (6.6.0.0) succeeds; older 6.5.0.0 is rate-limited. Probes
	// run newest-first, so we get partial progress then abort.
	store.rateLimitedVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false)
	require.NoError(t, err, "rate-limit abort must not fail the job")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=rate_limited"], "abort must record rate_limited")

	// 6.6 en+de succeeded; 6.5 en hit 429 and aborted before de (and before any
	// further versions, of which there are none).
	assert.Equal(t, []string{
		"en_GB@6.6.0.0", "de_DE@6.6.0.0",
		"en_GB@6.5.0.0",
	}, store.seenLocales())
	assert.Equal(t, int64(3), store.requests.Load())

	// Partial catalog for the successful version is kept.
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "partial catalog persisted")
	var latest66 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.6.0.0'`).Scan(&latest66))
	require.NotNil(t, latest66)
	assert.Equal(t, "1.2.0", *latest66)

	// No compatibility row for the aborted version, and sync bookkeeping must
	// not be advanced (otherwise the hourly freshness gate would starve retries).
	assert.Equal(t, 1, countRows(t, pool, "store_extension_compatibility"))
	assert.Equal(t, 0, countRows(t, pool, "store_extension_sync"), "bookkeeping not marked fresh after rate limit")

	// Incomplete work stays eligible for the next scheduled pass — both the
	// aborted Shopware version (compat gap) and a fresh re-sync of the same
	// names (bookkeeping never written).
	needing, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.5.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"FroshTools"}, needing, "aborted version must remain eligible")
	needing, err = h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"FroshTools"}, needing, "partial sync must not mark the name-set fresh")
}

// TestStoreExtensionSyncAllProbesRateLimited: when the first version is already
// rate-limited, nothing is persisted, bookkeeping stays stale, and the job
// still succeeds so the queue does not immediately retry.
func TestStoreExtensionSyncAllProbesRateLimited(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()
	collect := withStoreSyncMetrics(t)
	store.rateLimitedVersions["6.6.0.0"] = true
	store.rateLimitedVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false)
	require.NoError(t, err, "rate-limit abort must not fail the job")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=rate_limited"], "abort must record rate_limited")

	assert.Equal(t, []string{"en_GB@6.6.0.0"}, store.seenLocales(), "must not continue after first 429")
	assert.Equal(t, 0, countRows(t, pool, "store_extension"))
	assert.Equal(t, 0, countRows(t, pool, "store_extension_sync"))

	needing, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"FroshTools"}, needing, "unsynced names must remain eligible")
}

// TestStoreExtensionSyncNon429StillErrors: a non-429 probe failure on every
// version is a real job error, not a graceful backoff.
func TestStoreExtensionSyncNon429StillErrors(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	collect := withStoreSyncMetrics(t)
	store.serverErrorVersions["6.6.0.0"] = true
	store.serverErrorVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(context.Background(), []string{"FroshTools"}, "6.6.0.0", false)
	require.Error(t, err)
	assert.False(t, shopwareaccount.IsRateLimited(err), "500 must not be classified as rate-limited")
	assert.Contains(t, err.Error(), "all store probes failed")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=error"], "non-429 must record error")

	assert.Equal(t, 0, countRows(t, pool, "store_extension"))
	assert.Equal(t, 0, countRows(t, pool, "store_extension_sync"))
}

// TestStoreExtensionSyncRateLimitThenScheduledPass: after a 429 abort, a later
// SyncNames (the next scheduled scrape dispatch) finishes the remaining
// versions and only then marks bookkeeping fresh.
func TestStoreExtensionSyncRateLimitThenScheduledPass(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()
	store.rateLimitedVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false), "partial abort")
	assert.Equal(t, 0, countRows(t, pool, "store_extension_sync"))

	store.mu.Lock()
	store.rateLimitedVersions["6.5.0.0"] = false
	store.seen = nil
	store.mu.Unlock()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false), "scheduled follow-up")
	assert.Equal(t, 1, countRows(t, pool, "store_extension_sync"), "bookkeeping marked fresh only after a complete sync")
	assert.Equal(t, 2, countRows(t, pool, "store_extension_compatibility"), "both versions probed")

	needing, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Empty(t, needing, "complete sync must satisfy freshness")
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

	cleanup := maintenance.NewService(queries.New(pool))
	require.NoError(t, cleanup.CleanupOldData(ctx), "cleanup")

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
