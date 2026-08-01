package scrape

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/catalog/storemodel"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/environment"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/notify"
	"github.com/friendsofshopware/shopmon/api/internal/ptr"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("shopmon/monitoring/scrape")

// statusWeight maps status strings to numeric values for comparison.
var statusWeight = map[string]int{
	"green":  1,
	"yellow": 2,
	"red":    3,
}

// Service handles scraping of environments.
type Service struct {
	pool       *pgxpool.Pool
	queries    *queries.Queries
	cfg        *config.Config
	mail       mail.Sender
	catalog    CatalogSynchronizer
	dispatcher Dispatcher
	notifier   *notify.Dispatcher
}

type CatalogSynchronizer interface {
	SyncNames(ctx context.Context, names []string, shopwareVersion string, force bool) error
}

type Dispatcher interface {
	DispatchCatalogSync(ctx context.Context, names []string, shopwareVersion string) error
	DispatchSitespeedScrape(ctx context.Context, environmentID int32, delay time.Duration) error
}

// NewService creates an environment scrape service.
func NewService(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, mail mail.Sender, catalog CatalogSynchronizer, dispatcher Dispatcher) *Service {
	return &Service{
		pool:       pool,
		queries:    q,
		cfg:        cfg,
		mail:       mail,
		catalog:    catalog,
		dispatcher: dispatcher,
		notifier:   notify.NewDispatcher(q, mail),
	}
}

// Scrape refreshes a single environment by ID.
func (h *Service) Scrape(ctx context.Context, environmentID int32) error {
	ctx, span := tracer.Start(ctx, "environment.scrape")
	defer span.End()

	span.SetAttributes(attribute.Int("environment.id", int(environmentID)))
	slog.Info("scraping environment", "environmentId", environmentID)

	row, err := h.queries.GetEnvironmentForScrape(ctx, environmentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("environment %d not found: %w", environmentID, err)
	}

	env := queries.GetAllEnvironmentsRow(row)

	if err := h.scrapeEnvironment(ctx, env); err != nil {
		slog.Error("failed to scrape environment", "environmentId", environmentID, "error", err)
		return err
	}

	slog.Info("finished scraping environment", "environmentId", environmentID)
	return nil
}

// scrapeEnvironment performs the full scrape flow for a single environment.
func (h *Service) scrapeEnvironment(ctx context.Context, env queries.GetAllEnvironmentsRow) error {
	ctx, span := tracer.Start(ctx, "environment.scrape.environment",
		trace.WithAttributes(
			attribute.Int("environment.id", int(env.ID)),
			attribute.String("environment.name", env.Name),
			attribute.String("environment.url", env.Url),
		),
	)
	defer span.End()

	log := slog.With("environmentId", env.ID, "name", env.Name)

	log.Info("decrypting client secret")
	client, err := environment.NewClientFromEncrypted(env.Url, env.ClientID, env.ClientSecret, env.EnvironmentToken, h.cfg.AppSecret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}
	{
		authCtx, authSpan := tracer.Start(ctx, "environment.scrape.authenticate")
		err := client.Authenticate(authCtx)
		if err != nil {
			authSpan.RecordError(err)
			authSpan.SetStatus(codes.Error, err.Error())
		}
		authSpan.End()

		if err != nil {
			log.Warn("authentication failed", "error", err)
			h.notifyAuthError(ctx, env, err)
			if err := h.queries.UpdateEnvironmentScrapeError(ctx, queries.UpdateEnvironmentScrapeErrorParams{
				LastScrapedError: ptr.Of(fmt.Sprintf("Authentication failed: %v", err)),
				ID:               env.ID,
			}); err != nil {
				slog.Error("failed to update environment scrape error", "environmentId", env.ID, "error", err)
			}
			return nil
		}
	}

	type fetchResults struct {
		configData    []byte
		pluginData    []byte
		appData       []byte
		taskData      []byte
		queueData     []byte
		cacheInfoData []byte
		configErr     error
		pluginErr     error
		appErr        error
		taskErr       error
		queueErr      error
		cacheInfoErr  error
	}

	var fr fetchResults

	_, fetchSpan := tracer.Start(ctx, "environment.scrape.fetch_data")
	var wg sync.WaitGroup
	wg.Add(6)

	go func() {
		defer wg.Done()
		fr.configData, fr.configErr = client.Get(ctx, "/_info/config")
	}()
	go func() {
		defer wg.Done()
		fr.pluginData, fr.pluginErr = client.Post(ctx, "/search/plugin", map[string]interface{}{"limit": 500})
	}()
	go func() {
		defer wg.Done()
		fr.appData, fr.appErr = client.Post(ctx, "/search/app", map[string]interface{}{"limit": 500})
	}()
	go func() {
		defer wg.Done()
		fr.taskData, fr.taskErr = client.Post(ctx, "/search/scheduled-task", map[string]interface{}{"limit": 500})
	}()
	go func() {
		defer wg.Done()
		fr.queueData, fr.queueErr = client.Get(ctx, "/_info/queue.json")
	}()
	go func() {
		defer wg.Done()
		fr.cacheInfoData, fr.cacheInfoErr = client.Get(ctx, "/_action/cache_info")
	}()

	wg.Wait()
	fetchSpan.End()
	log.Info("fetched environment data")
	if fr.configErr != nil {
		errMsg := fmt.Sprintf("failed to fetch environment config: %v", fr.configErr)
		h.notifyDataFetchError(ctx, env, errMsg)
		if err := h.queries.UpdateEnvironmentScrapeError(ctx, queries.UpdateEnvironmentScrapeErrorParams{
			LastScrapedError: ptr.Of(errMsg),
			ID:               env.ID,
		}); err != nil {
			slog.Error("failed to update environment scrape error", "environmentId", env.ID, "error", err)
		}
		return nil
	}

	var shopConfig shopwareConfig
	if err := json.Unmarshal(fr.configData, &shopConfig); err != nil {
		return fmt.Errorf("decode config: %w", err)
	}

	var extensions []extensionEntry

	if fr.pluginErr == nil {
		var pluginsResp shopwareSearchResponse[shopwarePlugin]
		if err := json.Unmarshal(fr.pluginData, &pluginsResp); err != nil {
			slog.Warn("failed to parse Shopware plugins", "environmentId", env.ID, "error", err)
		} else {
			for _, p := range pluginsResp.Data {
				ext := extensionEntry{
					Name:        p.Name,
					Label:       p.Label,
					Active:      p.Active,
					Version:     p.Version,
					Installed:   p.InstalledAt != nil,
					InstalledAt: p.InstalledAt,
				}
				if p.UpgradeVersion != nil {
					ext.LatestVersion = p.UpgradeVersion
				}
				extensions = append(extensions, ext)
			}
		}
	} else {
		slog.Warn("failed to fetch Shopware plugins", "environmentId", env.ID, "error", fr.pluginErr)
	}

	if fr.appErr == nil {
		var appsResp shopwareSearchResponse[shopwareApp]
		if err := json.Unmarshal(fr.appData, &appsResp); err != nil {
			slog.Warn("failed to parse Shopware apps", "environmentId", env.ID, "error", err)
		} else {
			for _, a := range appsResp.Data {
				extensions = append(extensions, extensionEntry{
					Name:        a.Name,
					Label:       a.Label,
					Active:      a.Active,
					Version:     a.Version,
					Installed:   true,
					InstalledAt: ptr.Of(a.CreatedAt),
				})
			}
		}
	} else {
		slog.Warn("failed to fetch Shopware apps", "environmentId", env.ID, "error", fr.appErr)
	}

	var scheduledTasks []shopwareScheduledTask
	if fr.taskErr == nil {
		var tasksResp shopwareSearchResponse[shopwareScheduledTask]
		if err := json.Unmarshal(fr.taskData, &tasksResp); err != nil {
			slog.Warn("failed to parse Shopware scheduled tasks", "environmentId", env.ID, "error", err)
		} else {
			scheduledTasks = tasksResp.Data
		}
	} else {
		slog.Warn("failed to fetch Shopware scheduled tasks", "environmentId", env.ID, "error", fr.taskErr)
	}

	var queueEntries []shopwareQueueEntry
	if fr.queueErr == nil {
		if err := json.Unmarshal(fr.queueData, &queueEntries); err != nil {
			slog.Error("failed to parse queue data", "environmentId", env.ID, "error", err)
		}
	} else {
		slog.Warn("failed to fetch Shopware queue data", "environmentId", env.ID, "error", fr.queueErr)
	}

	var cacheInfo shopwareCacheInfo
	if fr.cacheInfoErr == nil {
		if err := json.Unmarshal(fr.cacheInfoData, &cacheInfo); err != nil {
			slog.Warn("failed to parse cache info", "environmentId", env.ID, "error", err)
		}
	} else {
		slog.Warn("failed to fetch Shopware cache info", "environmentId", env.ID, "error", fr.cacheInfoErr)
	}

	oldExtensions := h.loadExistingExtensions(ctx, env.ID)

	// Force-sync store extensions whose scraped version is missing from the
	// catalog (typically right after an extension update), so the changelog for
	// the update notification below comes out of the catalog in the same scrape.
	if h.catalog != nil {
		h.syncUpdatedExtensions(ctx, oldExtensions, extensions, shopConfig.Version)
	}

	// Classify extensions and resolve their compatible latest version from the
	// local catalog. The scrape never talks to the store API itself; missing or
	// stale catalog data is refreshed by the StoreExtensionSync job dispatched
	// after the persist.
	catalog, err := h.resolveExtensionsFromCatalog(ctx, extensions, oldExtensions, shopConfig.Version)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	extensionsDiff := calculateExtensionDiff(oldExtensions, extensions)
	h.attachStoreChangelogs(ctx, extensionsDiff, extensions)

	var ignores []string
	if env.Ignores != nil {
		if err := json.Unmarshal(env.Ignores, &ignores); err != nil {
			slog.Error("failed to unmarshal environment ignores", "environmentId", env.ID, "error", err)
		}
	}
	// Advisory suppressions live in their own table but suppress the same way as
	// a legacy ignore: the check is still emitted and persisted, it just stops
	// escalating the environment status. Unioning here keeps both mechanisms on
	// one code path, so existing ignores keep working untouched.
	ignores = append(ignores, h.suppressedCheckIDs(ctx, env.ID)...)

	checkerExtensions := make([]checker.Extension, len(extensions))
	for i, ext := range extensions {
		checkerExtensions[i] = checker.Extension{
			Name:          ext.Name,
			Label:         ext.Label,
			Active:        ext.Active,
			Version:       ext.Version,
			LatestVersion: ext.LatestVersion,
			Installed:     ext.Installed,
		}
	}

	checkerTasks := make([]checker.ScheduledTask, len(scheduledTasks))
	for i, st := range scheduledTasks {
		checkerTasks[i] = checker.ScheduledTask{
			ID:                st.ID,
			Name:              st.Name,
			Status:            st.Status,
			RunInterval:       int(st.RunInterval),
			NextExecutionTime: st.NextExecutionTime,
			LastExecutionTime: st.LastExecutionTime,
		}
	}

	checkerQueues := make([]checker.QueueInfo, len(queueEntries))
	for i, q := range queueEntries {
		size, err := q.Size.Int64()
		if err != nil {
			slog.Error("failed to parse queue size", "environmentId", env.ID, "queue", q.Name, "size", q.Size.String(), "error", err)
		}
		checkerQueues[i] = checker.QueueInfo{
			Name: q.Name,
			Size: int(size),
		}
	}

	securityAdvisories := loadSecurityAdvisories(ctx, h.queries)
	securityPluginFixes, securityPluginCoverage := loadSecurityPluginFixes(ctx, h.queries)

	// The SBOM needs the extension list to know whether FroshTools is present,
	// so it is fetched after the parallel block rather than inside it. Failures
	// are absorbed by fetchSbom and never fail the scrape.
	sbomRes := h.fetchSbom(ctx, env, client, extensions)

	checkerInput := checker.Input{
		Extensions: checkerExtensions,
		Config: checker.ShopConfig{
			Version: shopConfig.Version,
			AdminWorker: struct {
				EnableAdminWorker bool `json:"enableAdminWorker"`
			}{
				EnableAdminWorker: shopConfig.AdminWorker.EnableAdminWorker,
			},
		},
		ScheduledTasks: checkerTasks,
		QueueInfo:      checkerQueues,
		CacheInfo: checker.CacheInfo{
			Environment:  cacheInfo.Environment,
			HttpCache:    cacheInfo.HttpCache,
			CacheAdapter: cacheInfo.CacheAdapter,
		},
		Client:             client,
		Ignores:            ignores,
		SecurityAdvisories: securityAdvisories,
		// Composer package inventory from the FroshTools SBOM. Empty when the
		// shop cannot serve one, in which case checkSecurity falls back to
		// matching the Shopware core version only.
		InstalledPackages:      sbomPackageVersions(sbomRes),
		SecurityPluginFixes:    securityPluginFixes,
		SecurityPluginCoverage: securityPluginCoverage,
	}

	_, checkSpan := tracer.Start(ctx, "environment.scrape.run_checks")
	checkerResult := checker.RunAll(ctx, checkerInput)
	checkSpan.SetAttributes(attribute.String("check.status", string(checkerResult.Status)))
	checkSpan.End()

	newStatus := string(checkerResult.Status)

	// Capture the previously persisted checks before they are replaced so the
	// status transition can be explained by the specific checks that changed.
	oldChecks, err := h.queries.GetEnvironmentChecks(ctx, env.ID)
	if err != nil {
		slog.Warn("failed to load previous checks for status diff", "environmentId", env.ID, "error", err)
		oldChecks = nil
	}
	transition := h.handleStatusTransition(ctx, env, oldChecks, checkerResult.Checks, newStatus)

	// Fetch the favicon before opening the persist transaction: it is an HTTP
	// call with its own timeout and must not run while holding row locks.
	favicon := getFavicon(ctx, env.Url)

	if err := h.persistScrapeResult(ctx, env, scrapeResult{
		status:          newStatus,
		shopwareVersion: shopConfig.Version,
		extensions:      extensions,
		scheduledTasks:  scheduledTasks,
		queueEntries:    queueEntries,
		cacheInfo:       cacheInfo,
		checks:          checkerResult.Checks,
		extensionsDiff:  extensionsDiff,
		favicon:         favicon,
		transition:      transition,
		sbom:            sbomRes,
	}); err != nil {
		return err
	}

	h.dispatchStoreExtensionSync(ctx, extensions, shopConfig.Version, catalog)

	hasShopwareUpdate := env.ShopwareVersion != shopConfig.Version
	hasExtensionChanges := len(extensionsDiff) > 0

	span.SetAttributes(attribute.String("environment.version", shopConfig.Version))
	log.Info("scrape completed", "version", shopConfig.Version)

	if h.dispatcher != nil && env.SitespeedEnabled && (hasExtensionChanges || hasShopwareUpdate) {
		err := h.dispatcher.DispatchSitespeedScrape(ctx, env.ID, 15*time.Minute)
		if err != nil {
			log.Warn("failed to enqueue sitespeed job", "error", err)
		} else {
			log.Info("enqueued sitespeed job")
		}
	}

	slog.Info("scraped environment", "environmentId", env.ID, "name", env.Name, "version", shopConfig.Version)
	return nil
}

// loadSecurityAdvisories loads the visible Packagist advisory catalog for the
// security checker. Failures are logged and return an empty list so a stale or
// empty catalog never blocks the rest of the scrape.
func loadSecurityAdvisories(ctx context.Context, q *queries.Queries) []checker.SecurityAdvisory {
	rows, err := q.ListVisibleComposerAdvisories(ctx)
	if err != nil {
		slog.WarnContext(ctx, "failed to load security advisories for checks", "error", err)
		return nil
	}
	if len(rows) == 0 {
		return nil
	}

	ids := make([]string, 0, len(rows))
	for _, row := range rows {
		ids = append(ids, row.AdvisoryID)
	}
	pkgRows, err := q.ListComposerAdvisoryPackagesForAdvisories(ctx, ids)
	if err != nil {
		slog.WarnContext(ctx, "failed to load advisory packages for checks", "error", err)
		return nil
	}
	pkgsByID := make(map[string][]checker.SecurityAdvisoryPackage, len(ids))
	for _, p := range pkgRows {
		pkgsByID[p.AdvisoryID] = append(pkgsByID[p.AdvisoryID], checker.SecurityAdvisoryPackage{
			PackageName:      p.PackageName,
			AffectedVersions: p.AffectedVersions,
		})
	}

	out := make([]checker.SecurityAdvisory, 0, len(rows))
	for _, row := range rows {
		sev := ""
		if row.SeverityOverride != nil && *row.SeverityOverride != "" {
			sev = *row.SeverityOverride
		} else if row.Severity != nil {
			sev = *row.Severity
		}
		link := ""
		if row.RemediationUrl != nil && *row.RemediationUrl != "" {
			link = *row.RemediationUrl
		} else if row.Link != nil {
			link = *row.Link
		}
		cve := ""
		if row.Cve != nil {
			cve = *row.Cve
		}
		ghsa := ""
		if row.GhsaID != nil {
			ghsa = *row.GhsaID
		}
		out = append(out, checker.SecurityAdvisory{
			ID:       row.AdvisoryID,
			Title:    row.Title,
			CVE:      cve,
			GHSA:     ghsa,
			Link:     link,
			Severity: sev,
			Packages: pkgsByID[row.AdvisoryID],
		})
	}
	return out
}

// scrapeResult bundles everything a scrape produced for one environment, ready
// to be written to the database. Extensions must already be classified and
// resolved against the store catalog (isStore, LatestVersion).
type scrapeResult struct {
	status          string
	shopwareVersion string
	extensions      []extensionEntry
	scheduledTasks  []shopwareScheduledTask
	queueEntries    []shopwareQueueEntry
	cacheInfo       shopwareCacheInfo
	checks          []checker.Check
	extensionsDiff  []extensionDiff
	favicon         *string
	transition      *statusTransition
	sbom            sbomResult
}

// persistScrapeResult writes a scrape result in a single transaction. All
// writes are structured to avoid WAL churn on unchanged data: upserts carry
// change guards. The shared store catalog is not written here; it is owned by
// the StoreExtensionSync job, and scrapes only link against it.
func (h *Service) persistScrapeResult(ctx context.Context, env queries.GetAllEnvironmentsRow, res scrapeResult) (err error) {
	ctx, span := tracer.Start(ctx, "environment.scrape.persist",
		trace.WithAttributes(
			attribute.String("environment.status", res.status),
			attribute.Int("extension.count", len(res.extensions)),
			attribute.Int("check.count", len(res.checks)),
		),
	)
	defer func() {
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		}
		span.End()
	}()

	tx, err := h.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txQueries := h.queries.WithTx(tx)

	// Split extensions into store-known (linked to the shared catalog) and
	// unknown (environment_extension), then prune rows no longer present. Skipped
	// entirely when the shop returned no extensions (likely a fetch error) so a
	// transient failure does not wipe stored extensions. The catalog itself is
	// never written here; that is the StoreExtensionSync job's responsibility.
	if len(res.extensions) > 0 {
		storeNames := make([]string, 0, len(res.extensions))
		unknownNames := make([]string, 0, len(res.extensions))
		for _, ext := range res.extensions {
			if ext.isStore {
				storeNames = append(storeNames, ext.Name)
				if err := txQueries.UpsertEnvironmentStoreExtension(ctx, queries.UpsertEnvironmentStoreExtensionParams{
					EnvironmentID: env.ID,
					ExtensionName: ext.Name,
					Label:         ext.Label,
					Version:       ext.Version,
					LatestVersion: ext.LatestVersion,
					Active:        ext.Active,
					Installed:     ext.Installed,
					InstalledAt:   ext.InstalledAt,
				}); err != nil {
					return fmt.Errorf("upsert environment store extension %s: %w", ext.Name, err)
				}
			} else {
				unknownNames = append(unknownNames, ext.Name)
				if err := txQueries.UpsertEnvironmentExtension(ctx, queries.UpsertEnvironmentExtensionParams{
					EnvironmentID: env.ID,
					Name:          ext.Name,
					Label:         ext.Label,
					Active:        ext.Active,
					Version:       ext.Version,
					LatestVersion: ext.LatestVersion,
					Installed:     ext.Installed,
					InstalledAt:   ext.InstalledAt,
				}); err != nil {
					return fmt.Errorf("upsert environment extension %s: %w", ext.Name, err)
				}
			}
		}

		if err := txQueries.DeleteEnvironmentStoreExtensionsNotIn(ctx, queries.DeleteEnvironmentStoreExtensionsNotInParams{
			EnvironmentID: env.ID,
			Column2:       storeNames,
		}); err != nil {
			return fmt.Errorf("delete stale store extensions: %w", err)
		}
		if err := txQueries.DeleteEnvironmentExtensionsNotIn(ctx, queries.DeleteEnvironmentExtensionsNotInParams{
			EnvironmentID: env.ID,
			Column2:       unknownNames,
		}); err != nil {
			return fmt.Errorf("delete stale environment extensions: %w", err)
		}
	}

	taskIDs := make([]string, 0, len(res.scheduledTasks))
	for _, st := range res.scheduledTasks {
		taskIDs = append(taskIDs, st.ID)
		overdue := isScheduledTaskOverdue(st)
		if err := txQueries.UpsertEnvironmentScheduledTask(ctx, queries.UpsertEnvironmentScheduledTaskParams{
			EnvironmentID:     env.ID,
			TaskID:            st.ID,
			Name:              st.Name,
			Status:            st.Status,
			Interval:          st.RunInterval,
			Overdue:           overdue,
			LastExecutionTime: st.LastExecutionTime,
			NextExecutionTime: st.NextExecutionTime,
		}); err != nil {
			return fmt.Errorf("upsert scheduled task %s: %w", st.ID, err)
		}
	}
	if len(taskIDs) > 0 {
		if err := txQueries.DeleteEnvironmentScheduledTasksNotIn(ctx, queries.DeleteEnvironmentScheduledTasksNotInParams{
			EnvironmentID: env.ID,
			Column2:       taskIDs,
		}); err != nil {
			return fmt.Errorf("delete stale scheduled tasks: %w", err)
		}
	}

	queueNames := make([]string, 0, len(res.queueEntries))
	for _, q := range res.queueEntries {
		queueNames = append(queueNames, q.Name)
		size, err := q.Size.Int64()
		if err != nil {
			return fmt.Errorf("convert queue size for %s: %w", q.Name, err)
		}
		if err := txQueries.UpsertEnvironmentQueue(ctx, queries.UpsertEnvironmentQueueParams{
			EnvironmentID: env.ID,
			Name:          q.Name,
			Size:          int32(size),
		}); err != nil {
			return fmt.Errorf("upsert environment queue %s: %w", q.Name, err)
		}
	}
	if len(queueNames) > 0 {
		if err := txQueries.DeleteEnvironmentQueuesNotIn(ctx, queries.DeleteEnvironmentQueuesNotInParams{
			EnvironmentID: env.ID,
			Column2:       queueNames,
		}); err != nil {
			return fmt.Errorf("delete stale environment queues: %w", err)
		}
	}

	cacheAdapter := res.cacheInfo.CacheAdapter
	if cacheAdapter == "" {
		cacheAdapter = "filesystem"
	}
	if err := txQueries.UpsertEnvironmentCache(ctx, queries.UpsertEnvironmentCacheParams{
		EnvironmentID: env.ID,
		Environment:   res.cacheInfo.Environment,
		HttpCache:     res.cacheInfo.HttpCache,
		CacheAdapter:  cacheAdapter,
	}); err != nil {
		return fmt.Errorf("upsert environment cache: %w", err)
	}

	if err := persistSbom(ctx, txQueries, env.ID, res.sbom); err != nil {
		return fmt.Errorf("persist sbom: %w", err)
	}

	// Upsert checks and prune the ones no longer reported, instead of the previous
	// delete-all + re-insert which rewrote every check row each scrape.
	translator := h.notifier.Translator()
	checkIDs := make([]string, 0, len(res.checks))
	for _, c := range res.checks {
		checkIDs = append(checkIDs, c.ID)
		var link *string
		if c.Link != "" {
			link = &c.Link
		}

		// Persist key + params for per-viewer rendering, plus an English-rendered
		// fallback in the message column for legacy/non-i18n consumers.
		params := []byte("{}")
		if len(c.MessageParams) > 0 {
			if b, err := json.Marshal(c.MessageParams); err == nil {
				params = b
			}
		}
		var messageKey *string
		if c.MessageKey != "" {
			mk := c.MessageKey
			messageKey = &mk
		}
		message := translator.RenderCheck(notify.DefaultLocale, c.MessageKey, c.MessageParams)

		if err := txQueries.InsertEnvironmentCheck(ctx, queries.InsertEnvironmentCheckParams{
			EnvironmentID: env.ID,
			CheckID:       c.ID,
			Level:         string(c.Level),
			Message:       message,
			MessageKey:    messageKey,
			Params:        params,
			Source:        c.Source,
			Link:          link,
		}); err != nil {
			return fmt.Errorf("insert environment check %s: %w", c.ID, err)
		}
	}
	if err := txQueries.DeleteEnvironmentChecksNotIn(ctx, queries.DeleteEnvironmentChecksNotInParams{
		EnvironmentID: env.ID,
		Column2:       checkIDs,
	}); err != nil {
		return fmt.Errorf("delete stale environment checks: %w", err)
	}

	// Record the status transition on the timeline (alongside the new checks).
	if res.transition != nil {
		reasonsJSON, err := json.Marshal(res.transition.reasons)
		if err != nil || len(res.transition.reasons) == 0 {
			reasonsJSON = []byte("[]")
		}
		if err := txQueries.InsertEnvironmentStatusEvent(ctx, queries.InsertEnvironmentStatusEventParams{
			EnvironmentID: env.ID,
			OldStatus:     res.transition.oldStatus,
			NewStatus:     res.transition.newStatus,
			Reasons:       reasonsJSON,
		}); err != nil {
			return fmt.Errorf("insert environment status event: %w", err)
		}
	}

	hasShopwareUpdate := env.ShopwareVersion != res.shopwareVersion
	hasExtensionChanges := len(res.extensionsDiff) > 0

	var lastChangelogJSON []byte
	if hasShopwareUpdate {
		lastChangelogJSON, err = json.Marshal(map[string]interface{}{
			"date": time.Now().Format(time.RFC3339),
			"from": env.ShopwareVersion,
			"to":   res.shopwareVersion,
		})
		if err != nil {
			return fmt.Errorf("marshal Shopware version changelog: %w", err)
		}
	}

	if hasExtensionChanges || hasShopwareUpdate {
		extensionsDiffJSON, err := json.Marshal(res.extensionsDiff)
		if err != nil {
			return fmt.Errorf("marshal extension changes: %w", err)
		}

		var oldVersion, newVersion *string
		if hasShopwareUpdate {
			oldVersion = &env.ShopwareVersion
			newVersion = &res.shopwareVersion
		}

		envID := env.ID
		if err := txQueries.InsertEnvironmentChangelog(ctx, queries.InsertEnvironmentChangelogParams{
			EnvironmentID:      &envID,
			Extensions:         extensionsDiffJSON,
			OldShopwareVersion: oldVersion,
			NewShopwareVersion: newVersion,
		}); err != nil {
			return fmt.Errorf("insert environment changelog: %w", err)
		}
	}

	if err := txQueries.UpdateEnvironmentAfterScrape(ctx, queries.UpdateEnvironmentAfterScrapeParams{
		Status:           res.status,
		ShopwareVersion:  res.shopwareVersion,
		LastScrapedError: nil,
		Favicon:          res.favicon,
		EnvironmentImage: nil,
		LastChangelog:    lastChangelogJSON,
		ID:               env.ID,
	}); err != nil {
		return fmt.Errorf("update environment after scrape: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// resolveExtensionsFromCatalog classifies extensions as store-known (catalog
// membership) and resolves their compatible latest version from the local
// compatibility table. When a compatibility entry is not there yet (sync still
// pending), the shop-reported upgrade version is kept and, failing that, the
// prior link value is carried over so the update hint does not flicker away.
func (h *Service) resolveExtensionsFromCatalog(ctx context.Context, extensions []extensionEntry, oldExtensions []existingExtension, shopwareVersion string) (catalogState, error) {
	if len(extensions) == 0 {
		return catalogState{}, nil
	}

	names := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		names = append(names, ext.Name)
	}

	memberNames, err := h.queries.GetStoreExtensionNamesIn(ctx, names)
	if err != nil {
		return catalogState{}, fmt.Errorf("get store extension names: %w", err)
	}
	member := make(map[string]bool, len(memberNames))
	for _, n := range memberNames {
		member[n] = true
	}

	compatRows, err := h.queries.GetStoreExtensionCompatibility(ctx, queries.GetStoreExtensionCompatibilityParams{
		Column1:         names,
		ShopwareVersion: shopwareVersion,
	})
	if err != nil {
		return catalogState{}, fmt.Errorf("get store extension compatibility: %w", err)
	}
	compat := make(map[string]*string, len(compatRows))
	for _, row := range compatRows {
		compat[row.ExtensionName] = row.LatestVersion
	}

	oldStoreLatest := make(map[string]*string)
	for _, e := range oldExtensions {
		if e.IsStore {
			oldStoreLatest[e.Name] = e.LatestVersion
		}
	}

	for i := range extensions {
		ext := &extensions[i]
		if !member[ext.Name] {
			continue
		}
		ext.isStore = true
		if latest, ok := compat[ext.Name]; ok {
			// A present row with NULL latest means "checked, no compatible
			// release"; keep the shop-reported upgrade version in that case.
			if latest != nil {
				v := *latest
				ext.LatestVersion = &v
			}
		} else if ext.LatestVersion == nil {
			ext.LatestVersion = oldStoreLatest[ext.Name]
		}
	}

	return catalogState{member: member, compatKnown: compat}, nil
}

// syncUpdatedExtensions force-syncs store extensions whose freshly scraped
// version is missing from the catalog — typically right after an extension
// update — so the changelog for the update notification is available in the
// same scrape. Errors only cost the changelog text of one notification, so
// they are logged, not propagated.
func (h *Service) syncUpdatedExtensions(ctx context.Context, oldExtensions []existingExtension, extensions []extensionEntry, shopwareVersion string) {
	oldByName := make(map[string]existingExtension, len(oldExtensions))
	for _, e := range oldExtensions {
		oldByName[e.Name] = e
	}

	var stale []string
	for _, ext := range extensions {
		old, ok := oldByName[ext.Name]
		if !ok || !old.IsStore || old.Version == ext.Version {
			continue
		}
		versions, err := h.queries.GetStoreExtensionVersionsByName(ctx, ext.Name)
		if err != nil {
			slog.Warn("failed to load catalog versions", "extension", ext.Name, "error", err)
			continue
		}
		known := false
		for _, v := range versions {
			if v.Version == ext.Version {
				known = true
				break
			}
		}
		if !known {
			stale = append(stale, ext.Name)
		}
	}
	if len(stale) == 0 {
		return
	}

	if err := h.catalog.SyncNames(ctx, stale, shopwareVersion, true); err != nil {
		slog.Warn("failed to sync updated store extensions", "extensions", stale, "error", err)
	}
}

// attachStoreChangelogs fills in the changelog entries of "updated" diffs of
// store-known extensions from the local catalog.
func (h *Service) attachStoreChangelogs(ctx context.Context, diffs []extensionDiff, extensions []extensionEntry) {
	isStore := make(map[string]bool, len(extensions))
	for _, ext := range extensions {
		if ext.isStore {
			isStore[ext.Name] = true
		}
	}

	for i := range diffs {
		d := &diffs[i]
		if d.State != "updated" || !isStore[d.Name] || d.OldVersion == nil || d.NewVersion == nil {
			continue
		}
		rows, err := h.queries.GetStoreExtensionChangelogs(ctx, d.Name)
		if err != nil {
			slog.Warn("failed to load store changelogs", "extension", d.Name, "error", err)
			continue
		}
		mcs := make([]storemodel.MergedChangelog, 0, len(rows))
		for _, r := range rows {
			mc := storemodel.MergedChangelog{Version: r.Version}
			if r.ChangelogEn != nil {
				mc.English = *r.ChangelogEn
			}
			if r.ChangelogDe != nil {
				mc.German = *r.ChangelogDe
			}
			if r.ReleasedAt != nil {
				mc.ReleasedAt = *r.ReleasedAt
			}
			mcs = append(mcs, mc)
		}
		d.Changelog = storemodel.Between(mcs, *d.OldVersion, *d.NewVersion)
	}
}

// dispatchStoreExtensionSync enqueues a catalog sync for the environment's
// extensions whose catalog data is missing or stale. It reuses the membership
// and compatibility maps already resolved by resolveExtensionsFromCatalog, so
// only the freshness state is queried here. The sync job re-applies the same
// filter, so overlapping dispatches from other environments are cheap.
func (h *Service) dispatchStoreExtensionSync(ctx context.Context, extensions []extensionEntry, shopwareVersion string, catalog catalogState) {
	if h.dispatcher == nil || len(extensions) == 0 {
		return
	}

	names := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		names = append(names, ext.Name)
	}
	freshNames, err := h.queries.GetFreshStoreExtensionSyncNames(ctx, names)
	if err != nil {
		slog.Warn("failed to load fresh store extension sync names", "error", err)
		return
	}
	fresh := make(map[string]bool, len(freshNames))
	for _, name := range freshNames {
		fresh[name] = true
	}
	needing := make([]string, 0, len(names))
	for _, name := range names {
		_, compatibilityKnown := catalog.compatKnown[name]
		if !fresh[name] || (catalog.member[name] && shopwareVersion != "" && !compatibilityKnown) {
			needing = append(needing, name)
		}
	}
	if len(needing) == 0 {
		return
	}

	if err := h.dispatcher.DispatchCatalogSync(ctx, needing, shopwareVersion); err != nil {
		slog.Warn("failed to dispatch store extension sync", "error", err)
	}
}
