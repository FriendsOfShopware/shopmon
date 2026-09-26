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

const (
	// storeProbeRefreshWindowSeconds gates routine re-probes of an extension
	// name against the rate-limited store API (hit or miss): pluginsByName is
	// asked at most once per day per name. Store metadata (descriptions,
	// ratings, pictures) changes rarely, and new releases are detected through
	// the packages feed and probed via the event window instead. The window
	// matches the store-probe tier in GetFreshStoreExtensionSyncNames.
	storeProbeRefreshWindowSeconds = 24 * 60 * 60
	// storeProbeEventWindowSeconds gates event-driven probes (forced syncs,
	// releases the feed knows but the catalog does not, compatibility gaps of
	// feed-less store extensions): they may re-probe a name probed more than
	// 15 minutes ago, which deduplicates bursts (many environments updating
	// the same extension at once) without delaying event data by a day.
	storeProbeEventWindowSeconds = 15 * 60
)

// Service owns the shared store extension catalog. It is the
// only writer of the store_extension* tables: environment scrapes read the
// catalog from the database and dispatch a sync when data is missing or stale,
// so an extension installed on many environments is fetched at most once,
// regardless of how many environments have it.
//
// The sync is feed-first: compatibility and new-release detection come from
// the unauthenticated packages.shopware.com feed (no rate limits), while the
// rate-limited store API (pluginsByName) is only probed for what the feed
// does not have — store membership, metadata, translations, pictures, and
// bilingual changelogs. Store probes are claimed in the sync bookkeeping
// before they run, so overlapping sync jobs never probe the same name twice,
// and a confirmed store miss is not re-probed for a day.
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
// stale: never looked up, feed-checked over an hour ago, store-probed over a
// day ago, or store-known but lacking a compatibility entry for the given
// Shopware version. It returns the membership and compatibility state it used,
// so the caller can plan store probes without re-querying it.
func (h *Service) namesNeedingSync(ctx context.Context, names []string, shopwareVersion string) ([]string, catalogState, error) {
	if len(names) == 0 {
		return nil, catalogState{}, nil
	}

	memberNames, err := h.queries.GetStoreExtensionNamesIn(ctx, names)
	if err != nil {
		return nil, catalogState{}, fmt.Errorf("get store extension names: %w", err)
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
			return nil, catalogState{}, fmt.Errorf("get store extension compatibility: %w", err)
		}
		for _, row := range compatRows {
			compatKnown[row.ExtensionName] = row.LatestVersion
		}
	}

	catalog := catalogState{member: member, compatKnown: compatKnown}
	needing, err := h.namesNeedingSyncWith(ctx, names, shopwareVersion, catalog)
	if err != nil {
		return nil, catalogState{}, err
	}
	return needing, catalog, nil
}

// namesNeedingSyncWith applies the freshness filter given pre-resolved
// membership and compatibility state. Only the freshness bookkeeping is queried
// here, so the scrape's dispatch path reuses state it already read.
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

// SyncNames refreshes the shared catalog for the given extension names.
//
// Every needed name goes through the packages.shopware.com feed pass, which
// yields the compatible latest version for every Shopware version currently in
// use and detects releases the catalog has not fetched yet. Only names that
// additionally need store API data are probed against pluginsByName — one
// batched request per Shopware version and locale, each name at a single
// version instead of once per in-use version. Names unknown to the store are
// recorded in the sync bookkeeping so they are not asked again for a day.
// With force the freshness filter is skipped (used inline by scrapes when a
// just-updated extension version is not in the catalog yet); forced probes
// still go through the short event claim window so a burst of scrapes updating
// the same extension does not multiply store requests.
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

	// Duplicate names would violate the upsert statements' cardinality, and
	// syncing the same name twice is wasted work anyway.
	names = dedup(names)

	needed := names
	catalog := catalogState{}
	if !force {
		needed, catalog, err = h.namesNeedingSync(ctx, names, shopwareVersion)
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

	// The feed pass resolves compatibility for every in-use Shopware version
	// locally: the feed lists every release with its shopware/core constraint,
	// so the rate-limited store API is not probed once per version. The store
	// is still probed for what the feed does not have — store membership,
	// metadata, translations, pictures, and the bilingual changelog history
	// the global latest version is derived from.
	//
	// compat[name][swv] is the compatible latest version for that Shopware
	// version (nil for "checked, no compatible release"); feedResolved marks
	// names whose compat map came from the feed. best[name] is the plugin data
	// used to build the shared catalog subtree.
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

	// Membership and the catalog's global latest version, for probe planning
	// (new-release detection) and for the compat-only persist path.
	latestRows, err := h.queries.GetStoreExtensionLatestVersions(ctx, needed)
	if err != nil {
		return fmt.Errorf("get store extension latest versions: %w", err)
	}
	latest := make(map[string]*string, len(latestRows))
	for _, row := range latestRows {
		latest[row.Name] = row.LatestVersion
	}

	plan := planStoreProbes(needed, shopwareVersion, swvs, force, latest, compat, catalog)

	// Claim the planned probes in the shared bookkeeping before hitting the
	// store: a name claimed by another worker (or probed within the claim
	// window) is skipped, so overlapping sync jobs — dispatched by concurrent
	// scrapes of environments sharing extensions — never duplicate a
	// rate-limited probe. Claims of names that end up not probed are released
	// again below, so failures stay eligible for the next scheduled pass.
	claimed, claimedNames, err := h.claimStoreProbes(ctx, plan)
	if err != nil {
		return err
	}
	probedOK := make(map[string]bool, len(claimedNames))
	defer func() {
		release := make([]string, 0, len(claimedNames))
		for _, name := range claimedNames {
			// On success (including a graceful rate-limit abort) the claims of
			// successfully probed names stand; everything else — and all claims
			// when the sync fails — is released so the work is retried.
			if err == nil && probedOK[name] {
				continue
			}
			release = append(release, name)
		}
		if err := h.releaseStoreProbes(context.WithoutCancel(ctx), release); err != nil {
			slog.WarnContext(ctx, "failed to release store probe claims", "error", err)
		}
	}()

	// Batch the claimed probes by Shopware version: one request per version
	// and locale. Batches run sequentially (and locales within a batch
	// serially) so a rate-limit from the store is not amplified by fan-out. On
	// 429 after client retries, remaining batches are aborted. Partial
	// progress is persisted without advancing sync bookkeeping so the next
	// scheduled scrape can finish the work. The job returns success so the
	// queue does not nack/retry into a still-limited Store API.
	best := make(map[string]*storemodel.Data, len(claimed))
	anyProbeSucceeded := false
	var rateLimitErr error

	if len(claimed) > 0 {
		client := h.accountClient()
		for _, swv := range groupProbeVersions(claimed) {
			probeNames := claimedNamesAtVersion(claimed, swv)
			enPlugins, dePlugins, err := h.probeVersion(ctx, client, swv, probeNames)
			if err != nil {
				if shopwareaccount.IsRateLimited(err) {
					// Stop probing further batches; continuing would only deepen
					// the 429 burst. Partial progress (if any) is persisted
					// below without advancing sync bookkeeping so the next
					// scheduled scrape retries.
					slog.Warn("store rate limited, aborting remaining version probes",
						"shopwareVersion", swv, "error", err)
					rateLimitErr = err
					break
				}
				// Skip this batch; its claims are released above so the names
				// are retried by the next scheduled sync.
				slog.Warn("failed to probe store plugins", "shopwareVersion", swv, "error", err)
				continue
			}
			anyProbeSucceeded = true

			enMap := storemodel.IndexPlugins(enPlugins)
			deMap := storemodel.IndexPlugins(dePlugins)

			for _, name := range probeNames {
				probedOK[name] = true
				sd := &storemodel.Data{English: enMap[name], German: deMap[name]}
				p := sd.Primary()

				if feedResolved[name] {
					if p != nil {
						best[name] = sd
					}
					continue
				}

				// Without a feed this probe is the compatibility source for the
				// probed Shopware version; an absent plugin records a NULL row
				// ("checked, no compatible release") for catalog members.
				if _, isMember := latest[name]; isMember || p != nil {
					if compat[name] == nil {
						compat[name] = make(map[string]*string, 1)
					}
					compat[name][swv] = nil
					if p != nil {
						compat[name][swv] = nilIfEmpty(p.Version)
						best[name] = sd
					}
				}
			}
		}
	}
	span.SetAttributes(attribute.Int("extension.probed", len(probedOK)))

	if len(claimed) > 0 && !anyProbeSucceeded {
		if rateLimitErr != nil {
			// Bookkeeping is deliberately not advanced so the next scheduled
			// scrape dispatch tries again. A rate-limit abort is not a job
			// failure: nacking would immediately re-hit a still-limited Store
			// API and burn retries.
			rateLimited = true
			return nil
		}
		return fmt.Errorf("all store probes failed for %d claimed name(s)", len(claimed))
	}

	synced := 0
	for _, name := range needed {
		sd, ok := best[name]
		if !ok {
			continue // not returned by the store; still recorded in the bookkeeping below
		}
		if err := h.persistExtension(ctx, name, sd, compat[name]); err != nil {
			return err
		}
		synced++
	}

	// Catalog members that were not returned by a store probe still get their
	// compatibility rows from the feed pass (or from a feed-less gap probe), so
	// a compatibility gap is filled without waiting for store metadata.
	for _, name := range needed {
		if _, ok := best[name]; ok {
			continue
		}
		if _, isMember := latest[name]; !isMember || compat[name] == nil {
			continue
		}
		if err := h.persistCompatibility(ctx, name, compat[name]); err != nil {
			return err
		}
	}
	span.SetAttributes(attribute.Int("extension.synced", synced))

	if rateLimitErr != nil {
		// Persist what we have but do not mark names fresh — released claims
		// and stale bookkeeping keep the name-set eligible for the next
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

// probePlanEntry is the store-probe plan for one extension name: the single
// Shopware version to probe at, and whether the probe is event-driven (forced
// sync, new release, compatibility gap) or a routine refresh. Event probes use
// a short claim window so they are not delayed by a recent routine probe;
// refresh probes use the daily window that caches store hits and misses.
type probePlanEntry struct {
	shopwareVersion string
	event           bool
}

// planStoreProbes decides at which Shopware version each needed name would be
// probed against the store API, and whether the probe is event-driven. Every
// needed name gets an entry; the claim step afterwards decides which probes
// actually run (a name probed within the claim window is skipped).
func planStoreProbes(needed []string, requested string, swvs []string, force bool, latest map[string]*string, feedCompat map[string]map[string]*string, catalog catalogState) map[string]probePlanEntry {
	plan := make(map[string]probePlanEntry, len(needed))
	for _, name := range needed {
		latestVersion, isMember := latest[name]
		byShopwareVersion, resolved := feedCompat[name]

		// Feed-resolved names are probed at the in-use version whose compatible
		// release is the newest: the store scopes plugin visibility and
		// changelog history to the requested version, so this yields the
		// richest response from a single probe. Feed-less names are probed at
		// the requesting environment's version — the feed covers every store
		// extension, so a name without a feed is almost certainly a custom
		// plugin, and one probe confirms or refutes store membership instead of
		// walking every in-use version.
		probeVersion := requestedOrNewest(requested, swvs)
		if resolved {
			probeVersion = bestProbeVersion(byShopwareVersion, swvs, probeVersion)
		}

		entry := probePlanEntry{shopwareVersion: probeVersion}
		switch {
		case force:
			entry.event = true
		case !isMember:
			// New or previously unknown name; the refresh window caches the
			// outcome, so a confirmed custom plugin is re-checked once a day.
		case resolved && newerReleaseKnown(latestVersion, byShopwareVersion):
			// The feed knows a release the catalog has not fetched yet: probe
			// for its bilingual changelog without waiting for the daily
			// refresh.
			entry.event = true
		case !resolved && requested != "":
			if _, known := catalog.compatKnown[name]; !known {
				// Store extension whose feed is unusable, lacking a
				// compatibility entry for the requesting version: fill the gap
				// lazily with a single probe instead of a full version walk.
				entry.event = true
			}
		}
		plan[name] = entry
	}
	return plan
}

// bestProbeVersion picks the in-use Shopware version whose feed-compatible
// release is the newest; ties go to the newest Shopware version (swvs is
// sorted newest first). Without any compatible release it returns fallback.
func bestProbeVersion(compat map[string]*string, swvs []string, fallback string) string {
	best := ""
	var bestRelease *string
	for _, swv := range swvs {
		v, ok := compat[swv]
		if !ok || v == nil {
			continue
		}
		if bestRelease == nil || version.Compare(*v, *bestRelease) > 0 {
			best, bestRelease = swv, v
		}
	}
	if best == "" {
		return fallback
	}
	return best
}

func requestedOrNewest(requested string, swvs []string) string {
	if requested != "" {
		return requested
	}
	return swvs[0]
}

// newerReleaseKnown reports whether the feed's newest compatible release for
// the in-use Shopware versions is ahead of the catalog's global latest
// version. Releases compatible only with Shopware versions no environment runs
// are ignored: a store probe could not fetch their changelog either.
func newerReleaseKnown(latest *string, compat map[string]*string) bool {
	var maxCompat *string
	for _, v := range compat {
		if v != nil && (maxCompat == nil || version.Compare(*v, *maxCompat) > 0) {
			maxCompat = v
		}
	}
	if maxCompat == nil {
		return false
	}
	if latest == nil || *latest == "" {
		return true
	}
	return version.Compare(*maxCompat, *latest) > 0
}

// claimStoreProbes atomically marks the planned names as probed-now and
// returns the ones this sync actually claimed: names with no bookkeeping row
// or whose last probe is older than the plan entry's claim window. Names
// already claimed by a concurrent worker are excluded, which deduplicates
// in-flight probes across worker processes.
func (h *Service) claimStoreProbes(ctx context.Context, plan map[string]probePlanEntry) (map[string]probePlanEntry, []string, error) {
	claimed := make(map[string]probePlanEntry, len(plan))
	if len(plan) == 0 {
		return claimed, nil, nil
	}

	var eventNames, refreshNames []string
	for name, entry := range plan {
		if entry.event {
			eventNames = append(eventNames, name)
		} else {
			refreshNames = append(refreshNames, name)
		}
	}

	claim := func(names []string, windowSeconds int32) error {
		if len(names) == 0 {
			return nil
		}
		got, err := h.queries.ClaimStoreExtensionProbes(ctx, queries.ClaimStoreExtensionProbesParams{
			Column1: names,
			Column2: windowSeconds,
		})
		if err != nil {
			return fmt.Errorf("claim store probes: %w", err)
		}
		for _, name := range got {
			claimed[name] = plan[name]
		}
		return nil
	}

	if err := claim(eventNames, storeProbeEventWindowSeconds); err != nil {
		return nil, nil, err
	}
	if err := claim(refreshNames, storeProbeRefreshWindowSeconds); err != nil {
		return nil, nil, err
	}

	claimedNames := make([]string, 0, len(claimed))
	for name := range claimed {
		claimedNames = append(claimedNames, name)
	}
	sort.Strings(claimedNames)
	return claimed, claimedNames, nil
}

// releaseStoreProbes clears the probe timestamp of previously claimed names,
// making them eligible for the next scheduled sync again.
func (h *Service) releaseStoreProbes(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}
	if err := h.queries.ReleaseStoreExtensionProbes(ctx, names); err != nil {
		return fmt.Errorf("release store probes: %w", err)
	}
	return nil
}

// groupProbeVersions returns the distinct probe versions of the claimed plan,
// newest first, so probe batches run in a deterministic order.
func groupProbeVersions(claimed map[string]probePlanEntry) []string {
	seen := make(map[string]bool, len(claimed))
	versions := make([]string, 0, len(claimed))
	for _, entry := range claimed {
		if !seen[entry.shopwareVersion] {
			seen[entry.shopwareVersion] = true
			versions = append(versions, entry.shopwareVersion)
		}
	}
	sort.Slice(versions, func(i, j int) bool {
		return version.Compare(versions[i], versions[j]) > 0
	})
	return versions
}

// claimedNamesAtVersion returns the claimed names planned for one Shopware
// version, sorted for deterministic batches.
func claimedNamesAtVersion(claimed map[string]probePlanEntry, shopwareVersion string) []string {
	names := make([]string, 0, len(claimed))
	for name, entry := range claimed {
		if entry.shopwareVersion == shopwareVersion {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	return names
}

func dedup(names []string) []string {
	if len(names) < 2 {
		return names
	}
	sorted := slices.Clone(names)
	slices.Sort(sorted)
	return slices.Compact(sorted)
}

// probeVersion fetches the given names from the store in both locales scoped to
// one Shopware version, sequentially (en then de) to avoid doubling concurrent
// pressure on the store API. It returns a rate-limit error immediately so the
// caller can abort remaining batches. Otherwise it returns an error only when
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

// persistCompatibility upserts compatibility rows for a catalog member whose
// catalog subtree did not change: the compatible latest version per in-use
// Shopware version from the feed pass, or the probed version's entry from a
// feed-less gap probe. A nil value records "checked, no compatible release".
func (h *Service) persistCompatibility(ctx context.Context, name string, compat map[string]*string) error {
	for swv, latest := range compat {
		if err := h.queries.UpsertStoreExtensionCompatibility(ctx, queries.UpsertStoreExtensionCompatibilityParams{
			ExtensionName:   name,
			ShopwareVersion: swv,
			LatestVersion:   latest,
		}); err != nil {
			return fmt.Errorf("upsert compatibility %s@%s: %w", name, swv, err)
		}
	}
	return nil
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
