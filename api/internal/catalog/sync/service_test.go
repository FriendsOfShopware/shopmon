package sync

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/maintenance"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/shopwarepackages"
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

// mockFeedVersion is one release in the mock packages.shopware.com feed.
type mockFeedVersion struct {
	version    string
	constraint string
}

// mockStoreServer serves the pluginsByName endpoint for a configurable set of
// extensions whose availability and reported latest version depend on the
// requested Shopware version, plus the packages.shopware.com package feeds,
// and counts requests.
type mockStoreServer struct {
	*httptest.Server
	requests     atomic.Int64
	feedRequests atomic.Int64
	extensions   map[string]*mockExtension
	// feeds maps a lowercased technical name to its package feed versions; a
	// name absent from the map gets a 404 like a package unknown to
	// packages.shopware.com.
	feeds map[string][]mockFeedVersion

	mu sync.Mutex
	// rateLimitedVersions, when set, causes pluginsByName for those Shopware
	// versions to respond with HTTP 429.
	rateLimitedVersions map[string]bool
	// serverErrorVersions, when set, causes pluginsByName for those Shopware
	// versions to respond with HTTP 500 (non-retryable, not a rate limit).
	serverErrorVersions map[string]bool
	// feedServerErrorNames, when set, causes the package feed for those
	// lowercased names to respond with HTTP 500.
	feedServerErrorNames map[string]bool
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
		feeds: map[string][]mockFeedVersion{
			"froshtools": {
				{version: "1.2.0", constraint: "~6.6.0"},
				{version: "1.1.0", constraint: "~6.5.0"},
				{version: "1.0.0", constraint: "~6.5.0"},
			},
		},
		rateLimitedVersions:  map[string]bool{},
		serverErrorVersions:  map[string]bool{},
		feedServerErrorNames: map[string]bool{},
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

	mux.HandleFunc("/feeds/package/store.shopware.com/", func(w http.ResponseWriter, r *http.Request) {
		m.feedRequests.Add(1)
		name := strings.TrimSuffix(strings.TrimPrefix(r.URL.Path, "/feeds/package/store.shopware.com/"), ".json")

		m.mu.Lock()
		versions, ok := m.feeds[name]
		serverError := m.feedServerErrorNames[name]
		m.mu.Unlock()

		if serverError {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		feedVersions := make([]map[string]any, 0, len(versions))
		for _, v := range versions {
			feedVersions = append(feedVersions, map[string]any{
				"version":   v.version,
				"date":      "2023-01-01T00:00:00+00:00",
				"changelog": "changelog " + v.version,
				"require":   map[string]string{"shopware/core": v.constraint},
			})
		}
		w.Header().Set("Content-Type", "application/json")
		assert.NoError(t, json.NewEncoder(w).Encode(map[string]any{
			"name":     "store.shopware.com/" + name,
			"versions": feedVersions,
		}), "encode mock feed response")
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

func (m *mockStoreServer) resetSeen() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.seen = nil
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
	h := NewService(pool, q, &config.Config{ShopwareAPIURL: store.URL, ShopwarePackagesURL: store.URL})

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

// ageSyncBookkeeping backdates the sync bookkeeping of every name, so tests
// can simulate the feed tier (last_synced_at) and the store-probe tier
// (last_store_probe_at) expiring independently.
func ageSyncBookkeeping(t *testing.T, pool *pgxpool.Pool, feedAge, probeAge string) {
	t.Helper()
	_, err := pool.Exec(context.Background(), `
		UPDATE store_extension_sync
		SET last_synced_at = NOW() - $1::interval,
		    last_store_probe_at = NOW() - $2::interval
	`, feedAge, probeAge)
	require.NoError(t, err, "age sync bookkeeping")
}

// storeProbeAge returns how long ago the store was probed for a name, NULL
// when the probe claim was released or the name was never probed.
func storeProbeAge(t *testing.T, pool *pgxpool.Pool, name string) *time.Duration {
	t.Helper()
	var age *time.Duration
	require.NoError(t, pool.QueryRow(context.Background(),
		`SELECT NOW() - last_store_probe_at FROM store_extension_sync WHERE extension_name = $1`, name).Scan(&age))
	return age
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

	// FroshTools is feed-resolved, so the store is probed once at the in-use
	// version with the newest compatible release (6.6.0.0); the feed-missing
	// CustomPlugin is confirmed with a single discovery probe at the
	// requesting version instead of walking every in-use version.
	assert.Equal(t, []string{
		"en_GB@6.6.0.0", "de_DE@6.6.0.0",
		"en_GB@6.5.0.0", "de_DE@6.5.0.0",
	}, store.seenLocales(), "one batched probe per name, en+de each")
	assert.Equal(t, int64(4), store.requests.Load(), "store requests")
	assert.Equal(t, int64(2), store.feedRequests.Load(), "feed requests (FroshTools hit, CustomPlugin 404)")

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
	require.NotNil(t, storeProbeAge(t, pool, "CustomPlugin"), "store miss records the probe time")
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
	// 6.6.0.0 environment that also exists. It has no packages feed, so the
	// store probe is the only source.
	store.extensions["LegacyExt"] = &mockExtension{
		latestByShopwareVersion: map[string]string{"6.5.0.0": "3.4.0"},
		changelogVersions:       []string{"3.4.0", "3.3.0"},
	}

	require.NoError(t, h.SyncNames(ctx, []string{"LegacyExt"}, "6.5.0.0", false), "sync")

	// The catalog and its changelog history are persisted from the single
	// discovery probe at the requesting version.
	assert.Equal(t, []string{"en_GB@6.5.0.0", "de_DE@6.5.0.0"}, store.seenLocales(), "discovery probe at the requesting version only")
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "catalog row")
	assert.Equal(t, 2, countRows(t, pool, "store_extension_version"), "version rows")

	var globalLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension WHERE name = 'LegacyExt'`).Scan(&globalLatest), "read catalog row")
	require.NotNil(t, globalLatest, "global latest")
	assert.Equal(t, "3.4.0", *globalLatest)

	// 6.5.0.0 gets a compatible latest from the probe. 6.6.0.0 is not probed
	// eagerly: without a feed, compatibility for other versions is filled
	// lazily when a scrape of an environment running them reports a gap.
	var latest65 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'LegacyExt' AND shopware_version = '6.5.0.0'`).Scan(&latest65), "read 6.5 compat")
	require.NotNil(t, latest65, "6.5.0.0 compatible latest")
	assert.Equal(t, "3.4.0", *latest65)
	assert.Equal(t, 1, countRows(t, pool, "store_extension_compatibility"), "only the probed version has a row")

	// It is marked synced, so the classification is stable rather than retried.
	assert.Equal(t, 1, countRows(t, pool, "store_extension_sync"))

	// A scrape of the 6.6.0.0 environment reports the same extension: the
	// compatibility gap re-triggers a sync, which fills the gap with a single
	// probe at 6.6.0.0 once the event claim window has passed.
	ageSyncBookkeeping(t, pool, "61 minutes", "20 minutes")
	store.resetSeen()
	require.NoError(t, h.SyncNames(ctx, []string{"LegacyExt"}, "6.6.0.0", false), "gap sync")

	assert.Equal(t, []string{"en_GB@6.6.0.0", "de_DE@6.6.0.0"}, store.seenLocales(), "gap probe at the requesting version")
	var latest66 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'LegacyExt' AND shopware_version = '6.6.0.0'`).Scan(&latest66), "read 6.6 compat")
	assert.Nil(t, latest66, "6.6.0.0 has no compatible release")
}

// TestStoreExtensionSyncCompatibilityFromFeed: compatibility for every in-use
// Shopware version is computed from the packages feed, so the store is probed
// only once for the extension's metadata (at the version with the newest
// compatible release) instead of once per in-use version.
func TestStoreExtensionSyncCompatibilityFromFeed(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false))

	assert.Equal(t, []string{"en_GB@6.6.0.0", "de_DE@6.6.0.0"}, store.seenLocales(),
		"feed-resolved extension is probed once, at the version with the newest compatible release")
	assert.Equal(t, int64(1), store.feedRequests.Load(), "one feed request")

	// Both compatibility rows come from the feed — including 6.5.0.0, which was
	// never probed against the store.
	for swv, want := range map[string]string{"6.5.0.0": "1.1.0", "6.6.0.0": "1.2.0"} {
		var latest *string
		require.NoErrorf(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = $1`, swv).Scan(&latest), "read compatibility %s", swv)
		require.NotNilf(t, latest, "compatible latest for %s", swv)
		assert.Equalf(t, want, *latest, "compatible latest for %s", swv)
	}

	// The catalog subtree is complete and nothing stays eligible for re-sync.
	assert.Equal(t, 1, countRows(t, pool, "store_extension"))
	assert.Equal(t, 3, countRows(t, pool, "store_extension_version"))
	needing, _, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.5.0.0")
	require.NoError(t, err)
	assert.Empty(t, needing, "feed-resolved sync satisfies the compatibility check")
}

// TestStoreExtensionSyncFeedFallback: when the feed cannot be used (server
// error, or no version with an evaluable constraint), the store probe is the
// compatibility source. The first sync discovers the extension at the
// requesting version; other versions are filled lazily when a scrape reports
// a compatibility gap for them.
func TestStoreExtensionSyncFeedFallback(t *testing.T) {
	for name, mutate := range map[string]func(*mockStoreServer){
		"feed server error": func(store *mockStoreServer) {
			store.feedServerErrorNames["froshtools"] = true
		},
		"no evaluable constraint": func(store *mockStoreServer) {
			store.feeds["froshtools"] = []mockFeedVersion{{version: "1.2.0", constraint: ""}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			h, pool, store := setupSyncTest(t)
			mutate(store)
			ctx := context.Background()

			require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false))

			assert.Equal(t, []string{"en_GB@6.5.0.0", "de_DE@6.5.0.0"}, store.seenLocales(),
				"discovery probe at the requesting version only, no version walk")

			var latest65 *string
			require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.5.0.0'`).Scan(&latest65), "read 6.5 compat")
			require.NotNil(t, latest65, "6.5.0.0 compatible latest")
			assert.Equal(t, "1.1.0", *latest65, "6.5.0.0 compatible latest from the probe")

			// A scrape of the 6.6.0.0 environment reports a compatibility gap;
			// the follow-up sync fills it with a single probe at 6.6.0.0.
			ageSyncBookkeeping(t, pool, "61 minutes", "20 minutes")
			store.resetSeen()
			require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false), "gap sync")

			assert.Equal(t, []string{"en_GB@6.6.0.0", "de_DE@6.6.0.0"}, store.seenLocales(), "gap probe at the gap version")
			var latest66 *string
			require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.6.0.0'`).Scan(&latest66), "read 6.6 compat")
			require.NotNil(t, latest66, "6.6.0.0 compatible latest")
			assert.Equal(t, "1.2.0", *latest66, "6.6.0.0 compatible latest from the gap probe")
		})
	}
}

func TestLatestCompatibleVersion(t *testing.T) {
	versions := []shopwarepackages.PackageVersion{
		{Version: "1.2.0", Require: map[string]string{"shopware/core": "~6.6.0"}},
		{Version: "1.1.0", Require: map[string]string{"shopware/core": "~6.5.0"}},
		{Version: "1.0.0", Require: map[string]string{"shopware/core": "~6.5.0"}},
	}

	latest, ok := latestCompatibleVersion(versions, "6.5.0.0")
	require.True(t, ok, "evaluable")
	require.NotNil(t, latest)
	assert.Equal(t, "1.1.0", *latest, "newest version matching 6.5")

	latest, ok = latestCompatibleVersion(versions, "6.6.0.0")
	require.True(t, ok, "evaluable")
	require.NotNil(t, latest)
	assert.Equal(t, "1.2.0", *latest, "newest version matching 6.6")

	// Pre-release versions and releases without a constraint are skipped.
	withNoise := append([]shopwarepackages.PackageVersion{
		{Version: "2.0.0-rc1", Require: map[string]string{"shopware/core": "~6.6.0"}},
		{Version: "1.3.0"},
	}, versions...)
	latest, ok = latestCompatibleVersion(withNoise, "6.6.0.0")
	require.True(t, ok, "evaluable")
	require.NotNil(t, latest)
	assert.Equal(t, "1.2.0", *latest, "pre-release and constraint-less versions skipped")

	// The require key match is case-insensitive and range constraints work.
	ranged := []shopwarepackages.PackageVersion{
		{Version: "1.5.0", Require: map[string]string{"Shopware/Core": ">=6.5.0 <6.7.0"}},
	}
	latest, ok = latestCompatibleVersion(ranged, "6.6.0.0")
	require.True(t, ok, "evaluable")
	require.NotNil(t, latest)
	assert.Equal(t, "1.5.0", *latest)

	// A constrained feed that matches nothing is evaluable but yields no version.
	latest, ok = latestCompatibleVersion(versions, "6.4.0.0")
	require.True(t, ok, "evaluable")
	assert.Nil(t, latest, "no compatible release")

	// A feed whose constraints cannot be parsed is not evaluable at all.
	broken := []shopwarepackages.PackageVersion{
		{Version: "1.0.0", Require: map[string]string{"shopware/core": "not-a-constraint"}},
	}
	latest, ok = latestCompatibleVersion(broken, "6.6.0.0")
	assert.False(t, ok, "not evaluable")
	assert.Nil(t, latest)
}

// TestStoreExtensionSyncFreshIsNoop: a second sync right after the first must
// be answered from the bookkeeping without any HTTP request or row write. This
// is what collapses concurrent dispatches from many environments sharing the
// same extensions into a single fetch.
func TestStoreExtensionSyncFreshIsNoop(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()
	feedRequestsAfterFirst := store.feedRequests.Load()
	before := snapshotRowVersions(t, pool)

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0", false), "second sync")

	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "fresh re-sync must make no extra store requests")
	assert.Equal(t, feedRequestsAfterFirst, store.feedRequests.Load(), "fresh re-sync must make no extra feed requests")
	after := snapshotRowVersions(t, pool)
	for key, xmin := range before {
		assert.Equalf(t, xmin, after[key], "row %s was rewritten by a fresh re-sync", key)
	}
}

// TestStoreExtensionSyncStoreTierFreshSkipsStoreProbe is the core reuse
// guarantee: once the store was probed, the hourly feed-driven re-syncs
// refresh compatibility from the packages feed alone and do not touch the
// rate-limited store API again until the daily probe tier expires.
func TestStoreExtensionSyncStoreTierFreshSkipsStoreProbe(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()
	before := snapshotRowVersions(t, pool)

	// The feed tier expires (hourly), the store-probe tier does not (daily).
	ageSyncBookkeeping(t, pool, "2 hours", "2 hours")
	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "hourly re-sync")

	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "store-fresh re-sync must not probe the store")
	assert.Equal(t, int64(2), store.feedRequests.Load(), "feed is re-fetched on the hourly cadence")

	// Compatibility rows are refreshed in place (change-guarded), nothing else
	// is rewritten.
	after := snapshotRowVersions(t, pool)
	for key, xmin := range before {
		assert.Equalf(t, xmin, after[key], "row %s was rewritten by a store-fresh re-sync", key)
	}

	// Once the daily probe tier expires, the store is probed again.
	ageSyncBookkeeping(t, pool, "2 hours", "25 hours")
	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "daily refresh")
	assert.Greater(t, store.requests.Load(), requestsAfterFirst, "expired probe tier re-probes the store")
}

// TestStoreExtensionSyncStoreMissCached24h: a name the store does not know (a
// custom plugin) is confirmed with one discovery probe and then cached for a
// day, instead of being re-probed on every hourly sync.
func TestStoreExtensionSyncStoreMissCached24h(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"CustomPlugin"}, "6.5.0.0", false), "first sync")
	assert.Equal(t, int64(2), store.requests.Load(), "discovery probe, en+de")
	require.NotNil(t, storeProbeAge(t, pool, "CustomPlugin"), "miss records the probe time")

	// Hourly re-syncs re-check the feed but not the store.
	ageSyncBookkeeping(t, pool, "2 hours", "2 hours")
	require.NoError(t, h.SyncNames(ctx, []string{"CustomPlugin"}, "6.5.0.0", false), "hourly re-sync")
	assert.Equal(t, int64(2), store.requests.Load(), "store miss is cached, no re-probe")

	// After a day the miss is re-confirmed with a single probe.
	ageSyncBookkeeping(t, pool, "2 hours", "25 hours")
	require.NoError(t, h.SyncNames(ctx, []string{"CustomPlugin"}, "6.5.0.0", false), "daily re-check")
	assert.Equal(t, int64(4), store.requests.Load(), "expired miss is re-confirmed once")
	assert.Equal(t, 0, countRows(t, pool, "store_extension"), "still no catalog row")
}

// TestStoreExtensionSyncNewReleaseEventProbe: when the feed learns about a
// release the catalog has not fetched yet, the store is probed through the
// short event claim window even though the daily probe tier is still fresh —
// the bilingual changelog of a new release must not wait for the daily
// refresh.
func TestStoreExtensionSyncNewReleaseEventProbe(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()

	// A new release appears (in the store and in the packages feed).
	frosh := store.extensions["FroshTools"]
	frosh.latestByShopwareVersion["6.6.0.0"] = "1.3.0"
	frosh.changelogVersions = append([]string{"1.3.0"}, frosh.changelogVersions...)
	store.feeds["froshtools"] = append([]mockFeedVersion{{version: "1.3.0", constraint: "~6.6.0"}}, store.feeds["froshtools"]...)

	// The feed tier expires (hourly); the last store probe is older than the
	// event window but far from the daily refresh window.
	ageSyncBookkeeping(t, pool, "2 hours", "20 minutes")
	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "release sync")

	assert.Greater(t, store.requests.Load(), requestsAfterFirst, "new release triggers an event probe")
	var globalLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension WHERE name = 'FroshTools'`).Scan(&globalLatest), "read catalog row")
	require.NotNil(t, globalLatest, "global latest after release")
	assert.Equal(t, "1.3.0", *globalLatest, "global latest after release")
	assert.Equal(t, 4, countRows(t, pool, "store_extension_version"), "version rows")
}

// TestStoreExtensionSyncNewReleaseWithinEventWindow: a release detected
// minutes after the last store probe does not immediately re-probe (the event
// window deduplicates bursts), but the feed still updates the compatibility
// rows, so update hints appear without waiting for the store probe.
func TestStoreExtensionSyncNewReleaseWithinEventWindow(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()

	frosh := store.extensions["FroshTools"]
	frosh.latestByShopwareVersion["6.6.0.0"] = "1.3.0"
	frosh.changelogVersions = append([]string{"1.3.0"}, frosh.changelogVersions...)
	store.feeds["froshtools"] = append([]mockFeedVersion{{version: "1.3.0", constraint: "~6.6.0"}}, store.feeds["froshtools"]...)

	// Last probe 5 minutes ago: inside the event window, so no store probe.
	ageSyncBookkeeping(t, pool, "2 hours", "5 minutes")
	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "release sync within event window")

	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "event window deduplicates the probe")

	// The update hint still propagates through the feed-derived compatibility
	// row; only the changelog text waits for the next probe.
	var latest66 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.6.0.0'`).Scan(&latest66), "read 6.6 compat")
	require.NotNil(t, latest66, "6.6.0.0 compatible latest")
	assert.Equal(t, "1.3.0", *latest66, "compat row updated from the feed without a store probe")

	var globalLatest *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension WHERE name = 'FroshTools'`).Scan(&globalLatest), "read catalog row")
	require.NotNil(t, globalLatest, "global latest")
	assert.Equal(t, "1.2.0", *globalLatest, "catalog latest unchanged until the probe runs")
}

// TestStoreExtensionSyncForceDedupedByEventWindow: a forced sync right after a
// probe must not re-fetch (a burst of scrapes updating the same extension
// shares one probe); once the event window has passed, a forced sync
// re-fetches but identical data must not produce a single new row version.
func TestStoreExtensionSyncForceDedupedByEventWindow(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "first sync")
	requestsAfterFirst := store.requests.Load()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", true), "forced sync inside the event window")
	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "forced sync inside the event window must not re-probe")

	ageSyncBookkeeping(t, pool, "0 minutes", "20 minutes")
	before := snapshotRowVersions(t, pool)

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", true), "forced sync after the event window")
	assert.Greater(t, store.requests.Load(), requestsAfterFirst, "forced sync past the event window must hit the store")
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

	// A new release appears (in the store and in the packages feed) and the
	// bookkeeping ages out.
	frosh := store.extensions["FroshTools"]
	frosh.latestByShopwareVersion["6.6.0.0"] = "1.3.0"
	frosh.changelogVersions = append([]string{"1.3.0"}, frosh.changelogVersions...)
	store.feeds["froshtools"] = append([]mockFeedVersion{{version: "1.3.0", constraint: "~6.6.0"}}, store.feeds["froshtools"]...)
	ageSyncBookkeeping(t, pool, "2 days", "2 days")

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
	needing, _, err := h.namesNeedingSync(ctx, []string{"FroshTools", "CustomPlugin"}, "6.5.0.0")
	require.NoError(t, err, "namesNeedingSync")
	assert.Empty(t, needing, "needing sync for known version")

	// New version: only the store-known extension needs a compatibility probe;
	// the store miss stays quiet until its bookkeeping ages out.
	needing, _, err = h.namesNeedingSync(ctx, []string{"FroshTools", "CustomPlugin"}, "6.7.0.0")
	require.NoError(t, err, "namesNeedingSync")
	assert.Equal(t, []string{"FroshTools"}, needing, "needing sync for new version")
}

// TestNamesNeedingSyncCompatibilityGapFilledFromFeed: a compatibility gap for
// a feed-resolved extension is filled from the packages feed alone — no store
// probe is needed to answer "what is the latest version compatible with this
// Shopware version".
func TestNamesNeedingSyncCompatibilityGapFilledFromFeed(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.5.0.0", false), "sync")
	requestsAfterFirst := store.requests.Load()

	// An environment upgrades to 6.7.0.0: the gap makes the name eligible, the
	// feed answers it, and the store is not probed again.
	seedEnvironment(t, pool, "Upgraded", "6.7.0.0")
	needing, _, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.7.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"FroshTools"}, needing, "gap makes the name eligible")

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools"}, "6.7.0.0", false), "gap sync")
	assert.Equal(t, requestsAfterFirst, store.requests.Load(), "gap filled from the feed without a store probe")

	var latest67 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.7.0.0'`).Scan(&latest67), "read 6.7 compat")
	assert.Nil(t, latest67, "6.7.0.0 has no compatible release, recorded from the feed")
}

// TestStoreExtensionSyncSerializesLocaleProbes: en and de for a version must
// not run concurrently, otherwise catalog sync doubles store API pressure.
func TestStoreExtensionSyncSerializesLocaleProbes(t *testing.T) {
	h, _, store := setupSyncTest(t)
	store.handlerDelay = 30 * time.Millisecond

	require.NoError(t, h.SyncNames(context.Background(), []string{"FroshTools"}, "6.5.0.0", false))

	assert.Equal(t, 1, store.maxConcurrent(), "locale probes must be serial")
	assert.Equal(t, []string{"en_GB@6.6.0.0", "de_DE@6.6.0.0"}, store.seenLocales(), "one probe batch, en then de")
}

// addOldExt registers a second feed-resolved extension whose newest compatible
// release is on 6.5.0.0, so a sync of both extensions probes two version
// batches: FroshTools at 6.6.0.0, OldExt at 6.5.0.0.
func addOldExt(store *mockStoreServer) {
	store.extensions["OldExt"] = &mockExtension{
		latestByShopwareVersion: map[string]string{"6.5.0.0": "2.0.0"},
		changelogVersions:       []string{"2.0.0", "1.9.0"},
	}
	store.feeds["oldext"] = []mockFeedVersion{{version: "2.0.0", constraint: "~6.5.0"}}
}

// TestStoreExtensionSyncAbortsRemainingVersionsOn429: once the store
// rate-limits a probe batch, SyncNames must stop probing further batches
// instead of stomping through the rest of the list, release the claims of the
// names it did not probe so a later scheduled sync can continue, persist any
// names already probed, and return nil so the queue acks instead of nacking
// into a still-limited API.
func TestStoreExtensionSyncAbortsRemainingVersionsOn429(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	addOldExt(store)
	ctx := context.Background()
	collect := withStoreSyncMetrics(t)

	// Newest batch (6.6.0.0: FroshTools) succeeds; the 6.5.0.0 batch (OldExt)
	// is rate-limited. Batches run newest-first, so we get partial progress
	// then abort.
	store.rateLimitedVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(ctx, []string{"FroshTools", "OldExt"}, "6.6.0.0", false)
	require.NoError(t, err, "rate-limit abort must not fail the job")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=rate_limited"], "abort must record rate_limited")

	// 6.6 en+de succeeded; 6.5 en hit 429 and aborted before de.
	assert.Equal(t, []string{
		"en_GB@6.6.0.0", "de_DE@6.6.0.0",
		"en_GB@6.5.0.0",
	}, store.seenLocales())
	assert.Equal(t, int64(3), store.requests.Load())

	// The successfully probed extension is persisted, including its
	// feed-derived compatibility rows.
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "partial catalog persisted")
	var latest66 *string
	require.NoError(t, pool.QueryRow(ctx, `SELECT latest_version FROM store_extension_compatibility WHERE extension_name = 'FroshTools' AND shopware_version = '6.6.0.0'`).Scan(&latest66))
	require.NotNil(t, latest66)
	assert.Equal(t, "1.2.0", *latest66)

	// The aborted name's claim is released, so it stays eligible for the next
	// scheduled pass instead of waiting out the claim window.
	assert.Nil(t, storeProbeAge(t, pool, "OldExt"), "aborted name's claim must be released")
	require.NotNil(t, storeProbeAge(t, pool, "FroshTools"), "probed name keeps its claim")

	needing, _, err := h.namesNeedingSync(ctx, []string{"OldExt"}, "6.5.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"OldExt"}, needing, "aborted name must remain eligible")
	needing, _, err = h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Empty(t, needing, "fully synced name must not stay eligible")
}

// TestStoreExtensionSyncAllProbesRateLimited: when the first batch is already
// rate-limited, nothing is persisted, all claims are released, and the job
// still succeeds so the queue does not immediately retry.
func TestStoreExtensionSyncAllProbesRateLimited(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	ctx := context.Background()
	collect := withStoreSyncMetrics(t)
	store.rateLimitedVersions["6.6.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(ctx, []string{"FroshTools"}, "6.6.0.0", false)
	require.NoError(t, err, "rate-limit abort must not fail the job")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=rate_limited"], "abort must record rate_limited")

	assert.Equal(t, []string{"en_GB@6.6.0.0"}, store.seenLocales(), "must not continue after first 429")
	assert.Equal(t, 0, countRows(t, pool, "store_extension"))
	assert.Nil(t, storeProbeAge(t, pool, "FroshTools"), "claim must be released after a 429")

	needing, _, err := h.namesNeedingSync(ctx, []string{"FroshTools"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Equal(t, []string{"FroshTools"}, needing, "unsynced names must remain eligible")
}

// TestStoreExtensionSyncNon429StillErrors: a non-429 probe failure on every
// batch is a real job error, not a graceful backoff — and the claims are
// released so the retried job can actually re-probe.
func TestStoreExtensionSyncNon429StillErrors(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	collect := withStoreSyncMetrics(t)
	store.serverErrorVersions["6.6.0.0"] = true
	h.account = fastStoreClient(store.URL)

	err := h.SyncNames(context.Background(), []string{"FroshTools"}, "6.6.0.0", false)
	require.Error(t, err)
	assert.False(t, shopwareaccount.IsRateLimited(err), "500 must not be classified as rate-limited")
	assert.Contains(t, err.Error(), "all store probes failed")
	assert.Equal(t, int64(1), collect()["shopmon.store_sync.outcome|outcome=error"], "non-429 must record error")

	assert.Equal(t, 0, countRows(t, pool, "store_extension"))
	assert.Nil(t, storeProbeAge(t, pool, "FroshTools"), "claim must be released after a failure")

	// The retry re-probes instead of being swallowed by the claim window.
	store.serverErrorVersions["6.6.0.0"] = false
	require.NoError(t, h.SyncNames(context.Background(), []string{"FroshTools"}, "6.6.0.0", false), "retry")
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "retry persists the catalog")
}

// TestStoreExtensionSyncRateLimitThenScheduledPass: after a 429 abort, a later
// SyncNames (the next scheduled scrape dispatch) finishes the remaining names
// without re-probing the ones that already succeeded, and only then marks
// bookkeeping fresh.
func TestStoreExtensionSyncRateLimitThenScheduledPass(t *testing.T) {
	h, pool, store := setupSyncTest(t)
	addOldExt(store)
	ctx := context.Background()
	store.rateLimitedVersions["6.5.0.0"] = true
	h.account = fastStoreClient(store.URL)

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "OldExt"}, "6.6.0.0", false), "partial abort")
	assert.Equal(t, 1, countRows(t, pool, "store_extension"), "FroshTools persisted before the abort")

	store.mu.Lock()
	store.rateLimitedVersions["6.5.0.0"] = false
	store.seen = nil
	store.mu.Unlock()

	require.NoError(t, h.SyncNames(ctx, []string{"FroshTools", "OldExt"}, "6.6.0.0", false), "scheduled follow-up")

	// FroshTools was fully synced before the abort and keeps its claim, so the
	// follow-up probes only the aborted name.
	assert.Equal(t, []string{"en_GB@6.5.0.0", "de_DE@6.5.0.0"}, store.seenLocales(), "follow-up probes only the aborted name")
	assert.Equal(t, 2, countRows(t, pool, "store_extension"), "both extensions persisted")
	assert.Equal(t, 4, countRows(t, pool, "store_extension_compatibility"), "both extensions have feed-derived compat rows")

	needing, _, err := h.namesNeedingSync(ctx, []string{"FroshTools", "OldExt"}, "6.6.0.0")
	require.NoError(t, err)
	assert.Empty(t, needing, "complete sync must satisfy freshness")
}

// TestClaimStoreExtensionProbes verifies the claim/release bookkeeping
// primitives the sync uses to deduplicate store probes across workers.
func TestClaimStoreExtensionProbes(t *testing.T) {
	_, pool, _ := setupSyncTest(t)
	ctx := context.Background()
	q := queries.New(pool)

	claim := func(names []string, windowSeconds int32) []string {
		got, err := q.ClaimStoreExtensionProbes(ctx, queries.ClaimStoreExtensionProbesParams{Column1: names, Column2: windowSeconds})
		require.NoError(t, err, "claim")
		return got
	}

	// Never-probed names are claimed.
	assert.Equal(t, []string{"FroshTools"}, claim([]string{"FroshTools"}, storeProbeRefreshWindowSeconds))
	// Inside the window the claim is denied — this is what collapses
	// overlapping sync jobs onto one probe.
	assert.Empty(t, claim([]string{"FroshTools"}, storeProbeRefreshWindowSeconds))
	// A shorter (event) window denies as well while the probe is fresher than it.
	assert.Empty(t, claim([]string{"FroshTools"}, storeProbeEventWindowSeconds))

	// A released claim can be taken again immediately (failure retry path).
	require.NoError(t, q.ReleaseStoreExtensionProbes(ctx, []string{"FroshTools"}), "release")
	assert.Equal(t, []string{"FroshTools"}, claim([]string{"FroshTools"}, storeProbeRefreshWindowSeconds), "released claim is claimable")

	// An aged-out probe is claimable again.
	_, err := pool.Exec(ctx, `UPDATE store_extension_sync SET last_store_probe_at = NOW() - INTERVAL '25 hours'`)
	require.NoError(t, err, "age probe")
	assert.Equal(t, []string{"FroshTools"}, claim([]string{"FroshTools"}, storeProbeRefreshWindowSeconds), "expired probe is claimable")
}

// TestClaimStoreExtensionProbesConcurrent: concurrent claims for the same name
// — overlapping sync jobs on different workers — yield exactly one winner, so
// a name is never probed twice at the same time.
func TestClaimStoreExtensionProbesConcurrent(t *testing.T) {
	_, pool, _ := setupSyncTest(t)
	ctx := context.Background()
	q := queries.New(pool)

	const workers = 8
	var wg sync.WaitGroup
	wins := atomic.Int64{}
	for range workers {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, err := q.ClaimStoreExtensionProbes(ctx, queries.ClaimStoreExtensionProbesParams{
				Column1: []string{"FroshTools"},
				Column2: storeProbeRefreshWindowSeconds,
			})
			if err != nil {
				t.Errorf("claim: %v", err)
				return
			}
			wins.Add(int64(len(got)))
		}()
	}
	wg.Wait()
	assert.Equal(t, int64(1), wins.Load(), "exactly one concurrent claim wins")
}

// TestPlanStoreProbes unit-tests the probe planning: which store API probe (if
// the claim lets it through) each name gets, and at which Shopware version.
func TestPlanStoreProbes(t *testing.T) {
	swvs := []string{"6.6.0.0", "6.5.0.0"}
	compat := map[string]map[string]*string{
		"FeedExt": {"6.5.0.0": new("1.1.0"), "6.6.0.0": new("1.2.0")},
	}
	memberLatest := map[string]*string{"FeedExt": new("1.2.0"), "FeedlessMember": new("2.0.0")}

	t.Run("feed-resolved probes at the newest compatible release's version", func(t *testing.T) {
		plan := planStoreProbes([]string{"FeedExt"}, "6.5.0.0", swvs, false, memberLatest, compat, catalogState{})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.6.0.0"}, plan["FeedExt"], "routine refresh, probed where the newest release is compatible")
	})

	t.Run("feed-less names probe at the requesting version", func(t *testing.T) {
		plan := planStoreProbes([]string{"CustomPlugin"}, "6.5.0.0", swvs, false, memberLatest, compat, catalogState{})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.5.0.0"}, plan["CustomPlugin"])
	})

	t.Run("force marks probes as events", func(t *testing.T) {
		plan := planStoreProbes([]string{"FeedExt"}, "6.5.0.0", swvs, true, memberLatest, compat, catalogState{})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.6.0.0", event: true}, plan["FeedExt"])
	})

	t.Run("new release marks an event", func(t *testing.T) {
		staleCatalog := map[string]*string{"FeedExt": new("1.1.0")}
		plan := planStoreProbes([]string{"FeedExt"}, "6.5.0.0", swvs, false, staleCatalog, compat, catalogState{})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.6.0.0", event: true}, plan["FeedExt"], "catalog behind the feed is an event probe")
	})

	t.Run("feed-less member with a compatibility gap marks an event", func(t *testing.T) {
		plan := planStoreProbes([]string{"FeedlessMember"}, "6.5.0.0", swvs, false, memberLatest, compat, catalogState{
			member:      map[string]bool{"FeedlessMember": true},
			compatKnown: map[string]*string{},
		})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.5.0.0", event: true}, plan["FeedlessMember"])
	})

	t.Run("feed-less member without a gap is a routine refresh", func(t *testing.T) {
		plan := planStoreProbes([]string{"FeedlessMember"}, "6.5.0.0", swvs, false, memberLatest, compat, catalogState{
			member:      map[string]bool{"FeedlessMember": true},
			compatKnown: map[string]*string{"FeedlessMember": new("2.0.0")},
		})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.5.0.0"}, plan["FeedlessMember"])
	})

	t.Run("release compatible only with unused versions is not an event", func(t *testing.T) {
		// The feed's newest release (1.3.0) targets 6.7, which no environment
		// runs: the store probe could not fetch it either, so it must not
		// re-trigger an event probe on every sync.
		futureCompat := map[string]map[string]*string{
			"FeedExt": {"6.5.0.0": new("1.1.0"), "6.6.0.0": new("1.2.0"), "6.7.0.0": nil},
		}
		plan := planStoreProbes([]string{"FeedExt"}, "6.5.0.0", swvs, false, memberLatest, futureCompat, catalogState{})
		assert.Equal(t, probePlanEntry{shopwareVersion: "6.6.0.0"}, plan["FeedExt"], "no event probe")
	})
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
