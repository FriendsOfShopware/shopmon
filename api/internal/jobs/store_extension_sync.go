package jobs

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"sync"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/version"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

// StoreExtensionSyncHandler owns the shared store extension catalog. It is the
// only writer of the store_extension* tables: environment scrapes read the
// catalog from the database and dispatch a sync when data is missing or stale,
// so an extension installed on many environments is fetched from the store API
// at most once per hour instead of once per environment per scrape.
type StoreExtensionSyncHandler struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	cfg     *config.Config
}

// NewStoreExtensionSyncHandler creates a new StoreExtensionSyncHandler.
func NewStoreExtensionSyncHandler(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config) *StoreExtensionSyncHandler {
	return &StoreExtensionSyncHandler{pool: pool, queries: q, cfg: cfg}
}

// HandleSync processes a StoreExtensionSync queue message.
func (h *StoreExtensionSyncHandler) HandleSync(ctx context.Context, msg StoreExtensionSync) error {
	return h.SyncNames(ctx, msg.Names, msg.ShopwareVersion, false)
}

// namesNeedingSync filters names down to those whose catalog data is missing or
// stale: never looked up (or last looked up over an hour ago), or store-known but
// lacking a compatibility entry for the given Shopware version. The scrape uses
// the same filter to decide what to dispatch, and the job re-applies it so
// duplicate dispatches from many environments collapse into no-ops.
func (h *StoreExtensionSyncHandler) namesNeedingSync(ctx context.Context, names []string, shopwareVersion string) ([]string, error) {
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

	memberNames, err := h.queries.GetStoreExtensionNamesIn(ctx, names)
	if err != nil {
		return nil, fmt.Errorf("get store extension names: %w", err)
	}
	member := make(map[string]bool, len(memberNames))
	for _, n := range memberNames {
		member[n] = true
	}

	compatKnown := make(map[string]bool)
	if shopwareVersion != "" {
		compatRows, err := h.queries.GetStoreExtensionCompatibility(ctx, queries.GetStoreExtensionCompatibilityParams{
			Column1:         names,
			ShopwareVersion: shopwareVersion,
		})
		if err != nil {
			return nil, fmt.Errorf("get store extension compatibility: %w", err)
		}
		for _, row := range compatRows {
			compatKnown[row.ExtensionName] = true
		}
	}

	var needing []string
	for _, n := range names {
		if !fresh[n] || (member[n] && shopwareVersion != "" && !compatKnown[n]) {
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
func (h *StoreExtensionSyncHandler) SyncNames(ctx context.Context, names []string, shopwareVersion string, force bool) (err error) {
	ctx, span := tracer.Start(ctx, "store_extension.sync",
		trace.WithAttributes(
			attribute.Int("extension.count", len(names)),
			attribute.Bool("sync.force", force),
		),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	needed := names
	if !force {
		needed, err = h.namesNeedingSync(ctx, names, shopwareVersion)
		if err != nil {
			return err
		}
	}
	if len(needed) == 0 {
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
	// The newest version doubles as the scope for the full catalog fetch: store
	// metadata and the changelog list are version-independent, only the reported
	// compatible latest version differs.
	catalogSwv := swvs[0]

	client := shopwareaccount.NewClient(h.cfg.ShopwareAPIURL, nil)

	var (
		enPlugins, dePlugins []shopwareaccount.StorePlugin
		enErr, deErr         error
		wg                   sync.WaitGroup
	)
	wg.Add(2)
	go func() {
		defer wg.Done()
		enPlugins, enErr = client.PluginsByName(ctx, "en_GB", catalogSwv, needed)
	}()
	go func() {
		defer wg.Done()
		dePlugins, deErr = client.PluginsByName(ctx, "de_DE", catalogSwv, needed)
	}()
	wg.Wait()

	if enErr != nil && deErr != nil {
		// Bookkeeping is deliberately not updated so the queue retry (or the next
		// scrape dispatch) tries again.
		return fmt.Errorf("fetch store plugins: en: %w, de: %v", enErr, deErr)
	}
	if enErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "en_GB", "error", enErr)
	}
	if deErr != nil {
		slog.Warn("failed to fetch store plugins", "locale", "de_DE", "error", deErr)
	}

	enMap := indexStorePlugins(enPlugins)
	deMap := indexStorePlugins(dePlugins)

	// Compatible latest version per probed Shopware version. The catalog fetch
	// already answered catalogSwv; every other in-use version costs one extra
	// call for the whole batch.
	compat := map[string]map[string]string{catalogSwv: {}}
	for name, p := range enMap {
		compat[catalogSwv][name] = p.Version
	}
	for name, p := range deMap {
		if _, ok := compat[catalogSwv][name]; !ok {
			compat[catalogSwv][name] = p.Version
		}
	}
	for _, swv := range swvs[1:] {
		plugins, err := client.PluginsByName(ctx, "en_GB", swv, needed)
		if err != nil {
			// Skip this version; its missing compatibility rows re-trigger a sync
			// from the next scrape of an environment running it.
			slog.Warn("failed to probe store compatibility", "shopwareVersion", swv, "error", err)
			continue
		}
		versions := make(map[string]string, len(plugins))
		for i := range plugins {
			versions[plugins[i].Name] = plugins[i].Version
		}
		compat[swv] = versions
	}

	synced := 0
	for _, name := range needed {
		sd := &storeExtensionData{en: enMap[name], de: deMap[name]}
		if sd.primary() == nil {
			continue // not a store extension; still recorded in the bookkeeping below
		}
		if err := h.persistExtension(ctx, name, sd, compat); err != nil {
			return err
		}
		synced++
	}
	span.SetAttributes(attribute.Int("extension.synced", synced))

	if err := h.queries.UpsertStoreExtensionSyncStates(ctx, needed); err != nil {
		return fmt.Errorf("upsert sync states: %w", err)
	}

	slog.Info("synced store extensions", "requested", len(names), "fetched", len(needed), "found", synced)
	return nil
}

// shopwareVersionsToProbe returns every Shopware version in use by any
// environment plus the requesting one, newest first.
func (h *StoreExtensionSyncHandler) shopwareVersionsToProbe(ctx context.Context, shopwareVersion string) ([]string, error) {
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
func persistStoreExtensionCatalog(ctx context.Context, q *queries.Queries, name string, sd *storeExtensionData) error {
	p := sd.primary()

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
		LatestVersion: nilIfEmpty(sd.globalLatestVersion()),
	}); err != nil {
		return fmt.Errorf("upsert store extension %s: %w", name, err)
	}

	// Persist one translation row per locale the store returned.
	for lang, resp := range map[string]*shopwareaccount.StorePlugin{"en": sd.en, "de": sd.de} {
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

	for _, mc := range sd.mergedChangelogs() {
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
		for lang, text := range map[string]string{"en": mc.En, "de": mc.De} {
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
func (h *StoreExtensionSyncHandler) persistExtension(ctx context.Context, name string, sd *storeExtensionData, compat map[string]map[string]string) error {
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
