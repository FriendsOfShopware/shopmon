package jobs

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"math"
	"sync"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/crypto"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/shopware/checker"
	"github.com/jackc/pgx/v5/pgxpool"
	goqueue "github.com/shyim/go-queue"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

var tracer = otel.Tracer("shopmon/jobs")

// statusWeight maps status strings to numeric values for comparison.
var statusWeight = map[string]int{
	"green":  1,
	"yellow": 2,
	"red":    3,
}

// EnvironmentScrapeHandler handles scraping of environments.
type EnvironmentScrapeHandler struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	cfg     *config.Config
	bus     *goqueue.Bus
	mail    *mail.Service
}

// NewEnvironmentScrapeHandler creates a new EnvironmentScrapeHandler.
func NewEnvironmentScrapeHandler(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, bus *goqueue.Bus, mail *mail.Service) *EnvironmentScrapeHandler {
	return &EnvironmentScrapeHandler{pool: pool, queries: q, cfg: cfg, bus: bus, mail: mail}
}

// HandleScrapeAll scrapes all environments.
func (h *EnvironmentScrapeHandler) HandleScrapeAll(ctx context.Context, _ EnvironmentScrapeAll) error {
	ctx, span := tracer.Start(ctx, "environment.scrape_all")
	defer span.End()

	environments, err := h.queries.GetAllEnvironments(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("get all environments: %w", err)
	}

	span.SetAttributes(attribute.Int("environment.count", len(environments)))

	for _, env := range environments {
		if env.ConnectionIssueCount >= 3 {
			slog.Warn("skipping environment due to connection issues", "environmentId", env.ID, "count", env.ConnectionIssueCount)
			continue
		}

		if err := h.scrapeEnvironment(ctx, env); err != nil {
			slog.Error("failed to scrape environment", "environmentId", env.ID, "error", err)
			if err := h.queries.UpdateEnvironmentScrapeError(ctx, queries.UpdateEnvironmentScrapeErrorParams{
				LastScrapedError: strPtr(err.Error()),
				ID:               env.ID,
			}); err != nil {
				slog.Error("failed to update environment scrape error", "environmentId", env.ID, "error", err)
			}
		}
	}

	return nil
}

// HandleScrape scrapes a single environment by ID from the message payload.
func (h *EnvironmentScrapeHandler) HandleScrape(ctx context.Context, msg EnvironmentScrape) error {
	ctx, span := tracer.Start(ctx, "environment.scrape")
	defer span.End()

	span.SetAttributes(attribute.Int("environment.id", int(msg.EnvironmentID)))
	slog.Info("scraping environment", "environmentId", msg.EnvironmentID)

	row, err := h.queries.GetEnvironmentForScrape(ctx, msg.EnvironmentID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("environment %d not found: %w", msg.EnvironmentID, err)
	}

	env := queries.GetAllEnvironmentsRow(row)

	if err := h.scrapeEnvironment(ctx, env); err != nil {
		slog.Error("failed to scrape environment", "environmentId", msg.EnvironmentID, "error", err)
		return err
	}

	slog.Info("finished scraping environment", "environmentId", msg.EnvironmentID)
	return nil
}

// scrapeEnvironment performs the full scrape flow for a single environment.
func (h *EnvironmentScrapeHandler) scrapeEnvironment(ctx context.Context, env queries.GetAllEnvironmentsRow) error {
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
	clientSecret, err := crypto.Decrypt(env.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("decrypt client secret: %w", err)
	}

	client := shopware.NewClient(env.Url, env.ClientID, clientSecret, env.EnvironmentToken)

	{
		_, authSpan := tracer.Start(ctx, "environment.scrape.authenticate")
		err := client.Authenticate()
		if err != nil {
			authSpan.RecordError(err)
			authSpan.SetStatus(codes.Error, err.Error())
		}
		authSpan.End()

		if err != nil {
			log.Warn("authentication failed", "error", err)
			h.notifyAuthError(ctx, env, err)
			if err := h.queries.UpdateEnvironmentScrapeError(ctx, queries.UpdateEnvironmentScrapeErrorParams{
				LastScrapedError: strPtr(fmt.Sprintf("Authentication failed: %v", err)),
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
			LastScrapedError: strPtr(errMsg),
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
		if err := json.Unmarshal(fr.pluginData, &pluginsResp); err == nil {
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
	}

	if fr.appErr == nil {
		var appsResp shopwareSearchResponse[shopwareApp]
		if err := json.Unmarshal(fr.appData, &appsResp); err == nil {
			for _, a := range appsResp.Data {
				extensions = append(extensions, extensionEntry{
					Name:        a.Name,
					Label:       a.Label,
					Active:      a.Active,
					Version:     a.Version,
					Installed:   true,
					InstalledAt: strPtr(a.CreatedAt),
				})
			}
		}
	}

	var scheduledTasks []shopwareScheduledTask
	if fr.taskErr == nil {
		var tasksResp shopwareSearchResponse[shopwareScheduledTask]
		if err := json.Unmarshal(fr.taskData, &tasksResp); err == nil {
			scheduledTasks = tasksResp.Data
		}
	}

	var queueEntries []shopwareQueueEntry
	if fr.queueErr == nil {
		if err := json.Unmarshal(fr.queueData, &queueEntries); err != nil {
			slog.Warn("failed to parse queue data", "environmentId", env.ID, "error", err)
		}
	}

	var cacheInfo shopwareCacheInfo
	if fr.cacheInfoErr == nil {
		if err := json.Unmarshal(fr.cacheInfoData, &cacheInfo); err != nil {
			slog.Warn("failed to parse cache info", "environmentId", env.ID, "error", err)
		}
	}

	if len(extensions) > 0 {
		_, enrichSpan := tracer.Start(ctx, "environment.scrape.enrich_extensions",
			trace.WithAttributes(attribute.Int("extension.count", len(extensions))),
		)
		h.enrichExtensionsFromStore(extensions, shopConfig.Version)
		enrichSpan.End()
	}

	oldExtensions, err := h.queries.GetEnvironmentExtensions(ctx, env.ID)
	if err != nil {
		slog.Warn("failed to get old extensions", "environmentId", env.ID, "error", err)
		oldExtensions = nil
	}

	extensionsDiff := calculateExtensionDiff(oldExtensions, extensions)

	var ignores []string
	if env.Ignores != nil {
		if err := json.Unmarshal(env.Ignores, &ignores); err != nil {
			slog.Error("failed to unmarshal environment ignores", "environmentId", env.ID, "error", err)
		}
	}

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
		checkerQueues[i] = checker.QueueInfo{
			Name: q.Name,
			Size: int(q.Size),
		}
	}

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
		Client:  client,
		Ignores: ignores,
	}

	_, checkSpan := tracer.Start(ctx, "environment.scrape.run_checks")
	checkerResult := checker.RunAll(ctx, checkerInput)
	checkSpan.SetAttributes(attribute.String("check.status", string(checkerResult.Status)))
	checkSpan.End()

	newStatus := string(checkerResult.Status)
	h.handleStatusChange(ctx, env, newStatus)

	ctx, persistSpan := tracer.Start(ctx, "environment.scrape.persist",
		trace.WithAttributes(
			attribute.String("environment.status", newStatus),
			attribute.Int("extension.count", len(extensions)),
			attribute.Int("check.count", len(checkerResult.Checks)),
		),
	)
	tx, err := h.pool.Begin(ctx)
	if err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	txQueries := h.queries.WithTx(tx)

	if err := txQueries.DeleteEnvironmentChecks(ctx, env.ID); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("delete environment checks: %w", err)
	}

	extNames := make([]string, 0, len(extensions))
	for _, ext := range extensions {
		extNames = append(extNames, ext.Name)

		var ratingAvg *int32
		if ext.RatingAverage != nil {
			v := int32(math.Round(*ext.RatingAverage))
			ratingAvg = &v
		}

		var changelogJSON []byte
		if ext.Changelog != nil {
			changelogJSON, _ = json.Marshal(ext.Changelog)
		}

		if err := txQueries.UpsertEnvironmentExtension(ctx, queries.UpsertEnvironmentExtensionParams{
			EnvironmentID: env.ID,
			Name:          ext.Name,
			Label:         ext.Label,
			Active:        ext.Active,
			Version:       ext.Version,
			LatestVersion: ext.LatestVersion,
			Installed:     ext.Installed,
			RatingAverage: ratingAvg,
			StoreLink:     ext.StoreLink,
			Changelog:     changelogJSON,
			InstalledAt:   ext.InstalledAt,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("upsert environment extension %s: %w", ext.Name, err)
		}
	}
	if len(extNames) > 0 {
		if err := txQueries.DeleteEnvironmentExtensionsNotIn(ctx, queries.DeleteEnvironmentExtensionsNotInParams{
			EnvironmentID: env.ID,
			Column2:       extNames,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("delete stale environment extensions: %w", err)
		}
	}

	taskIDs := make([]string, 0, len(scheduledTasks))
	for _, st := range scheduledTasks {
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
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("upsert scheduled task %s: %w", st.ID, err)
		}
	}
	if len(taskIDs) > 0 {
		if err := txQueries.DeleteEnvironmentScheduledTasksNotIn(ctx, queries.DeleteEnvironmentScheduledTasksNotInParams{
			EnvironmentID: env.ID,
			Column2:       taskIDs,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("delete stale scheduled tasks: %w", err)
		}
	}

	queueNames := make([]string, 0, len(queueEntries))
	for _, q := range queueEntries {
		queueNames = append(queueNames, q.Name)
		if err := txQueries.UpsertEnvironmentQueue(ctx, queries.UpsertEnvironmentQueueParams{
			EnvironmentID: env.ID,
			Name:          q.Name,
			Size:          q.Size,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("upsert environment queue %s: %w", q.Name, err)
		}
	}
	if len(queueNames) > 0 {
		if err := txQueries.DeleteEnvironmentQueuesNotIn(ctx, queries.DeleteEnvironmentQueuesNotInParams{
			EnvironmentID: env.ID,
			Column2:       queueNames,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("delete stale environment queues: %w", err)
		}
	}

	cacheAdapter := cacheInfo.CacheAdapter
	if cacheAdapter == "" {
		cacheAdapter = "filesystem"
	}
	if err := txQueries.UpsertEnvironmentCache(ctx, queries.UpsertEnvironmentCacheParams{
		EnvironmentID: env.ID,
		Environment:   cacheInfo.Environment,
		HttpCache:     cacheInfo.HttpCache,
		CacheAdapter:  cacheAdapter,
	}); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("upsert environment cache: %w", err)
	}

	for _, c := range checkerResult.Checks {
		var link *string
		if c.Link != "" {
			link = &c.Link
		}
		if err := txQueries.InsertEnvironmentCheck(ctx, queries.InsertEnvironmentCheckParams{
			EnvironmentID: env.ID,
			CheckID:       c.ID,
			Level:         string(c.Level),
			Message:       c.Message,
			Source:        c.Source,
			Link:          link,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("insert environment check %s: %w", c.ID, err)
		}
	}

	hasShopwareUpdate := env.ShopwareVersion != shopConfig.Version
	hasExtensionChanges := len(extensionsDiff) > 0

	var lastChangelogJSON []byte
	if hasShopwareUpdate {
		lastChangelogJSON, _ = json.Marshal(map[string]interface{}{
			"date": time.Now().Format(time.RFC3339),
			"from": env.ShopwareVersion,
			"to":   shopConfig.Version,
		})
	}

	if hasExtensionChanges || hasShopwareUpdate {
		extensionsDiffJSON, _ := json.Marshal(extensionsDiff)

		var oldVersion, newVersion *string
		if hasShopwareUpdate {
			oldVersion = &env.ShopwareVersion
			newVersion = &shopConfig.Version
		}

		envID := env.ID
		if err := txQueries.InsertEnvironmentChangelog(ctx, queries.InsertEnvironmentChangelogParams{
			EnvironmentID:      &envID,
			Extensions:         extensionsDiffJSON,
			OldShopwareVersion: oldVersion,
			NewShopwareVersion: newVersion,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("insert environment changelog: %w", err)
		}
	}

	favicon := getFavicon(env.Url)

	if err := txQueries.UpdateEnvironmentAfterScrape(ctx, queries.UpdateEnvironmentAfterScrapeParams{
		Status:           newStatus,
		ShopwareVersion:  shopConfig.Version,
		LastScrapedError: nil,
		Favicon:          favicon,
		EnvironmentImage: nil,
		LastChangelog:     lastChangelogJSON,
		ID:               env.ID,
	}); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("update environment after scrape: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("commit transaction: %w", err)
	}
	persistSpan.End()

	span.SetAttributes(attribute.String("environment.version", shopConfig.Version))
	log.Info("scrape completed", "version", shopConfig.Version)

	if h.bus != nil && env.SitespeedEnabled && (hasExtensionChanges || hasShopwareUpdate) {
		err := goqueue.Dispatch(ctx, h.bus, SitespeedScrape{EnvironmentID: env.ID}, goqueue.WithDelay(15*time.Minute))
		if err != nil {
			log.Warn("failed to enqueue sitespeed job", "error", err)
		} else {
			log.Info("enqueued sitespeed job")
		}
	}

	slog.Info("scraped environment", "environmentId", env.ID, "name", env.Name, "version", shopConfig.Version)
	return nil
}
