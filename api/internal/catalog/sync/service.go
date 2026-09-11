package sync

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"math"
	"slices"
	"sort"
	"strings"

	"github.com/friendsofshopware/shopmon/api/internal/catalog/storemodel"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/metrics"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/shopwarepackages"
	"github.com/friendsofshopware/shopmon/api/internal/version"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
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
	// packages is an optional packages.shopware.com client override (tests).
	// When nil, SyncNames constructs one from cfg.ShopwarePackagesURL.
	packages *shopwarepackages.Client
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

func (h *Service) packagesClient() *shopwarepackages.Client {
	if h.packages != nil {
		return h.packages
	}
	return shopwarepackages.NewClient(h.cfg.ShopwarePackagesURL, nil)
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
	// rateLimited is tracked separately from err: a 429 abort is a successful
	// partial sync (job acked) but must still increment the rate_limited outcome.
	rateLimited := false
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
		if !recordOutcome {
			return
		}
		outcome := metrics.OutcomeOK
		switch {
		case rateLimited:
			outcome = metrics.OutcomeRateLimited
		case err != nil:
			outcome = metrics.OutcomeError
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

	// Compatibility is resolved from the packages.shopware.com feed wherever
	// possible: the feed lists every release with its shopware/core constraint,
	// so the compatible latest version for every in-use Shopware version can be
	// computed locally instead of probing the rate-limited store API once per
	// version. The store is still probed for what the feed does not have —
	// store membership, metadata, translations, pictures, and the changelog
	// history the global latest version is derived from.
	//
	// compat[name][swv] is the compatible latest version for that Shopware
	// version (nil for "checked, no compatible release"); feedResolved marks
	// names whose compat map came from the feed. best[name] is the plugin data
	// used to build the shared catalog subtree, preferring the probe with the
	// most changelog history (which yields the correct uncapped global latest
	// version).
	compat := make(map[string]map[string]*string, len(needed))
	feedResolved := make(map[string]bool, len(needed))
	for _, name := range needed {
		byShopwareVersion, err := h.compatibleVersionsFromFeed(ctx, name, swvs)
		if err != nil {
			return err
		}
		if byShopwareVersion != nil {
			compat[name] = byShopwareVersion
			feedResolved[name] = true
		}
	}

	// The store's pluginsByName endpoint is scoped to the requested Shopware
	// version: it omits any plugin whose releases are all incompatible with that
	// version. So an extension must be persisted from a probe that actually
	// returned it, not from a fixed "newest" version — otherwise an extension
	// installed only on an older environment (whose latest release predates the
	// newest environment's Shopware version) would be dropped from the catalog
	// while still being marked synced. Versions are therefore walked newest
	// first, probing only names that still need the store: not yet returned by
	// any probe, or without feed-resolved compatibility (the feed does not
	// cover custom plugins, so those keep the per-version probing).
	//
	// Versions are probed sequentially (and locales within a version serially)
	// so a rate-limit from the store is not amplified by fan-out. On 429 after
	// client retries, remaining versions are aborted. Partial progress is
	// persisted without advancing sync bookkeeping so the next scheduled scrape
	// can finish the work. The job returns success so the queue does not
	// nack/retry into a still-limited Store API.
	best := make(map[string]*storemodel.Data, len(needed))
	anyProbeSucceeded := false
	var rateLimitErr error

	for _, swv := range swvs {
		var probeNames []string
		for _, name := range needed {
			if _, found := best[name]; !found || !feedResolved[name] {
				probeNames = append(probeNames, name)
			}
		}
		if len(probeNames) == 0 {
			break
		}

		enPlugins, dePlugins, err := h.probeVersion(ctx, client, swv, probeNames)
		if err != nil {
			if shopwareaccount.IsRateLimited(err) {
				// Stop probing further versions; continuing would only deepen the
				// 429 burst. Partial progress (if any) is persisted below without
				// advancing sync bookkeeping so the next scheduled scrape retries.
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

		for _, name := range probeNames {
			sd := &storemodel.Data{English: enMap[name], German: deMap[name]}
			p := sd.Primary()

			if !feedResolved[name] {
				// Without a feed the compatible latest version comes from the
				// probe itself; an absent plugin records a NULL row ("checked,
				// no compatible release") for this Shopware version.
				if compat[name] == nil {
					compat[name] = make(map[string]*string, len(swvs))
				}
				var latest *string
				if p != nil {
					latest = nilIfEmpty(p.Version)
				}
				compat[name][swv] = latest
			}

			if p == nil {
				continue
			}
			if cur, ok := best[name]; !ok || len(sd.MergedChangelogs()) > len(cur.MergedChangelogs()) {
				best[name] = sd
			}
		}
	}

	if !anyProbeSucceeded {
		// Bookkeeping is deliberately not updated so the next scheduled scrape
		// dispatch tries again. A rate-limit abort is not a job failure: nacking
		// would immediately re-hit a still-limited Store API and burn retries.
		if rateLimitErr != nil {
			rateLimited = true
			return nil
		}
		return fmt.Errorf("all store probes failed for %d version(s)", len(swvs))
	}

	synced := 0
	for _, name := range needed {
		sd, ok := best[name]
		if !ok {
			continue // not a store extension; still recorded in the bookkeeping below
		}
		if err := h.persistExtension(ctx, name, sd, compat[name]); err != nil {
			return err
		}
		synced++
	}
	span.SetAttributes(attribute.Int("extension.synced", synced))

	if rateLimitErr != nil {
		// Persist what we have but do not mark names fresh — missing compatibility
		// rows and stale bookkeeping keep the name-set eligible for the next
		// scheduled sync. Returning nil acks the job instead of burning the
		// queue retry budget against a still-limited API.
		slog.Warn("synced store extensions partially before rate limit",
			"requested", len(names), "fetched", len(needed), "found", synced)
		rateLimited = true
		return nil
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

// compatibleVersionsFromFeed resolves the latest compatible extension version
// for each Shopware version from the packages.shopware.com feed. It returns
// nil when the feed cannot be used — unknown package (custom plugins are not
// on packages.shopware.com), request failure, or no version with an evaluable
// shopware/core constraint — in which case the caller falls back to probing
// the store API per Shopware version.
func (h *Service) compatibleVersionsFromFeed(ctx context.Context, name string, shopwareVersions []string) (map[string]*string, error) {
	feed, err := h.packagesClient().PackageFeed(ctx, storePackageName(name))
	if err != nil {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		if errors.Is(err, shopwarepackages.ErrNotFound) {
			slog.DebugContext(ctx, "no packages feed for extension, falling back to store probes", "extension", name)
		} else {
			slog.WarnContext(ctx, "failed to fetch packages feed, falling back to store probes", "extension", name, "error", err)
		}
		return nil, nil
	}

	byShopwareVersion := make(map[string]*string, len(shopwareVersions))
	evaluable := false
	for _, swv := range shopwareVersions {
		latest, ok := latestCompatibleVersion(feed.Versions, swv)
		if ok {
			evaluable = true
		}
		byShopwareVersion[swv] = latest
	}
	if !evaluable {
		slog.WarnContext(ctx, "packages feed has no evaluable versions, falling back to store probes", "extension", name)
		return nil, nil
	}
	return byShopwareVersion, nil
}

// latestCompatibleVersion returns the newest feed version whose shopware/core
// constraint matches the given Shopware version. Versions with a pre-release
// marker or without a constraint are skipped. The second return value reports
// whether any version had an evaluable constraint, so the caller can tell a
// genuinely incompatible extension apart from a feed it cannot make sense of.
func latestCompatibleVersion(versions []shopwarepackages.PackageVersion, shopwareVersion string) (*string, bool) {
	var latest *string
	evaluable := false
	for _, feedVersion := range versions {
		constraint := shopwareCoreConstraint(feedVersion.Require)
		if constraint == "" || strings.Contains(feedVersion.Version, "-") {
			continue
		}
		ok, err := version.Satisfies(shopwareVersion, constraint)
		if err != nil {
			continue
		}
		evaluable = true
		if !ok {
			continue
		}
		if latest == nil || version.Compare(*latest, feedVersion.Version) < 0 {
			v := feedVersion.Version
			latest = &v
		}
	}
	return latest, evaluable
}

func shopwareCoreConstraint(require map[string]string) string {
	for packageName, constraint := range require {
		if strings.EqualFold(packageName, "shopware/core") {
			return constraint
		}
	}
	return ""
}

// storePackageName maps a store technical name to its Composer package name on
// packages.shopware.com, which serves every store extension under a lowercased
// store.shopware.com/ vendor prefix.
func storePackageName(technicalName string) string {
	return "store.shopware.com/" + strings.ToLower(technicalName)
}

// shopwareVersionsToProbe returns every Shopware version in use by any
// environment plus the requesting one, newest first.
func (h *Service) shopwareVersionsToProbe(ctx context.Context, shopwareVersion string) ([]string, error) {
	swvs, err := h.queries.GetDistinctEnvironmentShopwareVersions(ctx)
	if err != nil {
		return nil, fmt.Errorf("get distinct shopware versions: %w", err)
	}
	found := slices.Contains(swvs, shopwareVersion)
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
// rows in a single transaction. compat maps each in-use Shopware version to
// its compatible latest version; a nil value records "checked, no compatible
// release", and a Shopware version absent from the map gets no row, so the
// compatibility gap re-triggers a sync on the next scrape.
func (h *Service) persistExtension(ctx context.Context, name string, sd *storemodel.Data, compat map[string]*string) error {
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	qtx := h.queries.WithTx(tx)

	if err := persistStoreExtensionCatalog(ctx, qtx, name, sd); err != nil {
		return err
	}

	for swv, latest := range compat {
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
