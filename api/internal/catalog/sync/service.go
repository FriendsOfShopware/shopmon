package sync

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"

	"github.com/friendsofshopware/shopmon/api/internal/catalog/storemodel"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
	"github.com/friendsofshopware/shopmon/api/internal/otelx"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/version"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("shopmon/catalog/sync")

// Service owns the shared store extension catalog. It is the
// only writer of the store_extension* tables: environment scrapes read the
// catalog from the database and dispatch a sync when data is missing or stale,
// so an extension installed on many environments is fetched from the store API
// at most once per hour instead of once per environment per scrape.
type Service struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	cfg     *config.Config
	// account is an optional store client override (tests). When nil, SyncNames
	// constructs one from cfg.ShopwareAPIURL.
	account *shopwareaccount.Client
}

// NewService creates a new Service.
func NewService(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config) *Service {
	return &Service{pool: pool, queries: q, cfg: cfg}
}

func (h *Service) accountClient() *shopwareaccount.Client {
	if h.account != nil {
		return h.account
	}
	return shopwareaccount.NewClient(h.cfg.ShopwareAPIURL, nil)
}

// Sync refreshes the requested names using the normal freshness checks.
func (h *Service) Sync(ctx context.Context, names []string, shopwareVersion string) error {
	return h.SyncNames(ctx, names, shopwareVersion, false)
}

// catalogState holds the per-name catalog lookups a scrape already performed:
// store membership and, keyed the same way, the compatibility entries known for
// the environment's Shopware version (value may be nil for "checked, no
// compatible release"). Passing it into namesNeedingSyncWith lets the scrape's
// dispatch reuse those reads instead of re-querying them.
type catalogState struct {
	member      map[string]bool
	compatKnown map[string]*string
}

// namesNeedingSync filters names down to those whose catalog data is missing or
// stale: never looked up (or last looked up over an hour ago), or store-known but
// lacking a compatibility entry for the given Shopware version. It fetches the
// membership and compatibility state itself; the scrape's dispatch path uses
// namesNeedingSyncWith to reuse state it already read.
func (h *Service) namesNeedingSync(ctx context.Context, names []string, shopwareVersion string) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	memberNames, err := h.queries.GetStoreExtensionNamesIn(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("get store extension names: %w", err)
	}
	member := make(map[string]bool, len(memberNames))
	for _, n := range memberNames {
		member[n] = true
	}

	compatKnown := make(map[string]*string)
	if shopwareVersion != "" {
		compatRows, err := h.queries.GetStoreExtensionCompatibility(ctx, queries.GetStoreExtensionCompatibilityParams{
			Column1:         names,
			ShopwareVersion: shopwareVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("get store extension compatibility: %w", err)
		}
		for _, row := range compatRows {
			compatKnown[row.ExtensionName] = row.LatestVersion
		}
	}

	return h.namesNeedingSyncWith(ctx, names, shopwareVersion, catalogState{member: member, compatKnown: compatKnown})
}

// namesNeedingSyncWith applies the freshness filter given pre-resolved
// membership and compatibility state. Only the freshness bookkeeping is queried
// here, so the scrape avoids re-reading membership and compatibility rows it
// just fetched in resolveExtensionsFromCatalog.
func (h *Service) namesNeedingSyncWith(ctx context.Context, names []string, shopwareVersion string, catalog catalogState) ([]string, error) {
	if len(names) == 0 {
		return nil, nil
	}

	freshNames, err := h.queries.GetFreshStoreExtensionSyncNames(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("get fresh sync names: %w", err)
	}
	fresh := make(map[string]bool, len(freshNames))
	for _, n := range freshNames {
		fresh[n] = true
	}

	var needing []string
	for _, n := range names {
		_, compatKnown := catalog.compatKnown[n]
		if !fresh[n] || (catalog.member[n] && shopwareVersion != "" && !compatKnown) {
			needing = append(needing, n)
		}
	}
	return needing, nil
}

// SyncNames refreshes the shared catalog for the given extension names: store
// metadata, translations, versions with changelogs, images, and the compatible
// latest version for every Shopware version currently in use by any
// environment. Names unknown to the store are recorded in the sync bookkeeping
// so they are not asked again for an hour. With force the freshness filter is
// skipped (used inline by scrapes when a just-updated extension version is not
// in the catalog yet).
func (h *Service) SyncNames(ctx context.Context, names []string, shopwareVersion string, force bool) (err error) {
	ctx, span := tracer.Start(ctx, "store_extension.sync",
		trace.WithAttributes(
			attribute.Int("extension.count", len(names)),
			attribute.Bool("sync.force", force),
		),
	)
	// recordOutcome stays false for the "nothing to sync" early return so
	// no-op freshness checks do not inflate shopmon.store_sync.outcome ok.
	recordOutcome := true
	defer func() {
		if err != nil {
			// 429 aborts are expected (job retries; shopmon.store_sync.outcome
			// =rate_limited). Other failures remain hard span errors.
			otelx.RecordDependency(span, err)
		}
		span.End()
		if !recordOutcome {
			return
		}
		outcome := metrics.OutcomeOK
		if err != nil {
			if shopwareaccount.IsRateLimited(err) {
				outcome = metrics.OutcomeRateLimited
			} else {
				outcome = metrics.OutcomeError
			}
		}
		metrics.RecordStoreSyncOutcome(ctx, outcome)
	}()

	needed := names
	if !force {
		needed, err = h.namesNeedingSync(ctx, names, shopwareVersion)
		if err != nil {
			return err
		}
	}
	if len(needed) == 0 {
		recordOutcome = false
		return nil
	}
	span.SetAttributes(attribute.Int("extension.needed", len(needed)))

	swvs, err := h.shopwareVersionsToProbe(ctx, shopwareVersion)
	if err != nil {
		return err
	}
	if len(swvs) == 0 {
		return nil
	}

	client := h.accountClient()

	// The store's pluginsByName endpoint is scoped to the requested Shopware
	// version: it omits any plugin whose releases are all incompatible with that
	// version. So an extension must be persisted from a probe that actually
	// returned it, not from a fixed "newest" version — otherwise an extension
	// installed only on an older environment (whose latest release predates the
	// newest environment's Shopware version) would be dropped from the catalog
	// while still being marked synced. Every in-use version is therefore probed
	// (en + de), and each extension's catalog data is taken from the richest
	// probe that returned it.
	//
	// Versions are probed sequentially (and locales within a version serially)
	// so a rate-limit from the store is not amplified by fan-out. On 429 after
	// client retries, remaining versions are aborted and the job returns an
	// error for a later retry instead of burning through the rest of the list.
	//
	// compat[swv][name] is the compatible latest version the store reports for
	// that Shopware version; best[name] is the plugin data used to build the
	// shared catalog subtree, preferring the probe with the most changelog
	// history (which yields the correct uncapped global latest version).
	compat := make(map[string]map[string]string, len(swvs))
	best := make(map[string]*storemodel.Data, len(needed))
	anyProbeSucceeded := false
	var rateLimitErr error

	for _, swv := range swvs {
		enPlugins, dePlugins, err := h.probeVersion(ctx, client, swv, needed)
		if err != nil {
			if shopwareaccount.IsRateLimited(err) {
				// Stop probing further versions; continuing would only deepen the
				// 429 burst. Partial progress (if any) is persisted below without
				// advancing sync bookkeeping so the queue / next scrape retries.
				slog.Warn("store rate limited, aborting remaining version probes",
					"shopwareVersion", swv, "error", err)
				rateLimitErr = err
				break
			}
			// Skip this version; its missing compatibility rows re-trigger a sync
			// from the next scrape of an environment running it.
			slog.Warn("failed to probe store plugins", "shopwareVersion", swv, "error", err)
			continue
		}
		anyProbeSucceeded = true

		enMap := storemodel.IndexPlugins(enPlugins)
		deMap := storemodel.IndexPlugins(dePlugins)

		versions := make(map[string]string)
		for _, name := range needed {
			sd := &storemodel.Data{English: enMap[name], German: deMap[name]}
			if p := sd.Primary(); p != nil {
				versions[name] = p.Version
				if cur, ok := best[name]; !ok || len(sd.MergedChangelogs()) > len(cur.MergedChangelogs()) {
					best[name] = sd
				}
			}
		}
		compat[swv] = versions
	}

	if !anyProbeSucceeded {
		// Bookkeeping is deliberately not updated so the queue retry (or the next
		// scrape dispatch) tries again.
		if rateLimitErr != nil {
			return fmt.Errorf("store rate limited; all probes failed for %d version(s): %w", len(swvs), rateLimitErr)
		}
		return fmt.Errorf("all store probes failed for %d version(s)", len(swvs))
	}

	synced := 0
	for _, name := range needed {
		sd, ok := best[name]
		if !ok {
			continue // not a store extension; still recorded in the bookkeeping below
		}
		if err := h.persistExtension(ctx, name, sd, compat); err != nil {
			return err
		}
		synced++
	}
	span.SetAttributes(attribute.Int("extension.synced", synced))

	if rateLimitErr != nil {
		// Persist what we have but do not mark names fresh — missing compatibility
		// rows and stale bookkeeping will re-dispatch, and the returned error lets
		// the queue retry with its own backoff.
		slog.Warn("synced store extensions partially before rate limit",
			"requested", len(names), "fetched", len(needed), "found", synced)
		return fmt.Errorf("store rate limited after probing %d/%d version(s): %w", len(compat), len(swvs), rateLimitErr)
	}

	if err := h.queries.UpsertStoreExtensionSyncStates(ctx, needed); err != nil {
		return fmt.Errorf("upsert sync states: %w", err)
	}

	slog.Info("synced store extensions", "requested", len(names), "fetched", len(needed), "found", synced)
	return nil
}

// probeVersion fetches the given names from the store in both locales scoped to
// one Shopware version, sequentially (en then de) to avoid doubling concurrent
// pressure on the store API. It returns a rate-limit error immediately so the
// caller can abort remaining versions. Otherwise it returns an error only when
// both locale calls fail; a single-locale failure is logged and its (nil)
// result is used.
func (h *Service) probeVersion(ctx context.Context, client *shopwareaccount.Client, shopwareVersion string, names []string) ([]shopwareaccount.StorePlugin, []shopwareaccount.StorePlugin, error) {
	enPlugins, enErr := client.PluginsByName(ctx, "en_GB", shopwareVersion, names)
	if shopwareaccount.IsRateLimited(enErr) {
		return nil, nil, enErr
	}
	dePlugins, deErr := client.PluginsByName(ctx, "de_DE", shopwareVersion, names)
	if shopwareaccount.IsRateLimited(deErr) {
		return nil, nil, deErr
	}

	if enErr != nil && deErr != nil {
		return nil, nil, fmt.Errorf("en: %w, de: %v", enErr, deErr)
	}
	if enErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "en_GB", "shopwareVersion", shopwareVersion, "error", enErr)
	}
	if deErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "de_DE", "shopwareVersion", shopwareVersion, "error", deErr)
	}
	return enPlugins, dePlugins, nil
}

// shopwareVersionsToProbe returns every Shopware version in use by any
// environment plus the requesting one, newest first.
func (h *Service) shopwareVersionsToProbe(ctx context.Context, shopwareVersion string) ([]string, error) {
	swvs, err := h.queries.GetDistinctEnvironmentShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("get distinct shopware versions: %w", err)
	}
	found := false
	for _, v := range swvs {
		if v == shopwareVersion {
			found = true
			break
		}
	}
	if !found && shopwareVersion != "" {
		swvs = append(swvs, shopwareVersion)
	}
	sort.Slice(swvs, func(i, j int) bool {
		return version.Compare(swvs[i], swvs[j]) > 0
	})
	return swvs, nil
}

// persistStoreExtensionCatalog refreshes the environment-independent store
// catalog data of one extension: metadata, translations, versions with their
// changelogs, and listing images.
func persistStoreExtensionCatalog(ctx context.Context, q *queries.Queries, name string, sd *storemodel.Data) error {
	p := sd.Primary()

	var storeID *int32
	if p.ID != 0 {
		v := int32(p.ID)
		storeID = &v
	}
	var rating *int32
	if p.RatingAverage != 0 {
		v := int32(math.Round(p.RatingAverage))
		rating = &v
	}

	if err := q.UpsertStoreExtension(ctx, queries.UpsertStoreExtensionParams{
		Name:            name,
		StoreID:         storeID,
		IconUrl:         nilIfEmpty(p.IconPath),
		ProducerName:    nilIfEmpty(p.ProducerName),
		ProducerWebsite: nilIfEmpty(p.ProducerWebsite),
		RatingAverage:   rating,
		StoreLink:       nilIfEmpty(p.StoreLink),
		ReleaseDate:     nilIfEmpty(p.ReleaseDate),
		// The catalog row is shared across environments, so it stores the
		// environment-independent global latest version rather than the
		// Shopware-version-capped value (which lives on the compatibility table).
		LatestVersion: nilIfEmpty(sd.GlobalLatestVersion()),
	}); err != nil {
		return fmt.Errorf("upsert store extension %s: %w", name, err)
	}

	// Persist one translation row per locale the store returned.
	for lang, resp := range map[string]*shopwareaccount.StorePlugin{"en": sd.English, "de": sd.German} {
		if resp == nil {
			continue
		}
		if err := q.UpsertStoreExtensionTranslation(ctx, queries.UpsertStoreExtensionTranslationParams{
			ExtensionName:      name,
			Language:           lang,
			Label:              nilIfEmpty(resp.Label),
			ShortDescription:   nilIfEmpty(resp.ShortDescription),
			Description:        nilIfEmpty(resp.Description),
			InstallationManual: nilIfEmpty(resp.InstallationManual),
		}); err != nil {
			return fmt.Errorf("upsert store extension translation %s/%s: %w", name, lang, err)
		}
	}

	// Released versions are immutable, so upsert only versions that are new or
	// whose release date changed instead of rewriting the full version history.
	existingVersions, err := q.GetStoreExtensionVersionsByName(ctx, name)
	if err != nil {
		return fmt.Errorf("get store extension versions %s: %w", name, err)
	}
	knownVersions := make(map[string]queries.GetStoreExtensionVersionsByNameRow, len(existingVersions))
	for _, v := range existingVersions {
		knownVersions[v.Version] = v
	}

	for _, mc := range sd.MergedChangelogs() {
		var versionID int32
		releasedAt := nilIfEmpty(mc.ReleasedAt)
		if known, ok := knownVersions[mc.Version]; ok && (releasedAt == nil || (known.ReleasedAt != nil && *known.ReleasedAt == *releasedAt)) {
			versionID = known.ID
		} else {
			versionID, err = q.UpsertStoreExtensionVersion(ctx, queries.UpsertStoreExtensionVersionParams{
				ExtensionName: name,
				Version:       mc.Version,
				ReleasedAt:    releasedAt,
			})
			if err != nil {
				return fmt.Errorf("upsert store extension version %s@%s: %w", name, mc.Version, err)
			}
		}
		for lang, text := range map[string]string{"en": mc.English, "de": mc.German} {
			if text == "" {
				continue
			}
			if err := q.UpsertStoreExtensionVersionTranslation(ctx, queries.UpsertStoreExtensionVersionTranslationParams{
				ExtensionVersionID: versionID,
				Language:           lang,
				Changelog:          nilIfEmpty(text),
			}); err != nil {
				return fmt.Errorf("upsert store extension version translation %s@%s/%s: %w", name, mc.Version, lang, err)
			}
		}
	}

	imageURLs := make([]string, 0, len(p.Pictures))
	for _, pic := range p.Pictures {
		imageURLs = append(imageURLs, pic.URL)
		if err := q.UpsertStoreExtensionImage(ctx, queries.UpsertStoreExtensionImageParams{
			ExtensionName: name,
			Url:           pic.URL,
			Preview:       pic.Preview,
			Priority:      int32(pic.Priority),
		}); err != nil {
			return fmt.Errorf("upsert store extension image %s: %w", name, err)
		}
	}
	// Only prune when the store returned pictures: an empty list would make
	// "url != ALL('{}')" match every row and wipe all stored images on a sync
	// that momentarily returned no pictures.
	if len(imageURLs) > 0 {
		if err := q.DeleteStoreExtensionImagesNotIn(ctx, queries.DeleteStoreExtensionImagesNotInParams{
			ExtensionName: name,
			Column2:       imageURLs,
		}); err != nil {
			return fmt.Errorf("delete stale store extension images %s: %w", name, err)
		}
	}

	return nil
}

// persistExtension writes one extension's catalog subtree and compatibility
// rows in a single transaction.
func (h *Service) persistExtension(ctx context.Context, name string, sd *storemodel.Data, compat map[string]map[string]string) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := h.queries.WithTx(tx)

	if err := persistStoreExtensionCatalog(ctx, qtx, name, sd); err != nil {
		return err
	}

	for swv, versions := range compat {
		// An extension absent from a version's probe has no compatible release
		// there; the NULL row records that the combination was checked.
		var latest *string
		if v, ok := versions[name]; ok && v != "" {
			latest = &v
		}
		if err := qtx.UpsertStoreExtensionCompatibility(ctx, queries.UpsertStoreExtensionCompatibilityParams{
			ExtensionName:   name,
			ShopwareVersion: swv,
			LatestVersion:   latest,
		}); err != nil {
			return fmt.Errorf("upsert compatibility %s@%s: %w", name, swv, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

func nilIfEmpty(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}
