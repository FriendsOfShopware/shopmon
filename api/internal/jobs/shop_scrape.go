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

// ShopScrapeHandler handles scraping of shops.
type ShopScrapeHandler struct {
	pool    *pgxpool.Pool
	queries *queries.Queries
	cfg     *config.Config
	bus     *goqueue.Bus
	mail    *mail.Service
}

// NewShopScrapeHandler creates a new ShopScrapeHandler.
func NewShopScrapeHandler(pool *pgxpool.Pool, q *queries.Queries, cfg *config.Config, bus *goqueue.Bus, mail *mail.Service) *ShopScrapeHandler {
	return &ShopScrapeHandler{pool: pool, queries: q, cfg: cfg, bus: bus, mail: mail}
}

// HandleScrapeAll scrapes all shops.
func (h *ShopScrapeHandler) HandleScrapeAll(ctx context.Context, _ ShopScrapeAll) error {
	ctx, span := tracer.Start(ctx, "shop.scrape_all")
	defer span.End()

	shops, err := h.queries.GetAllShops(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("get all shops: %w", err)
	}

	span.SetAttributes(attribute.Int("shop.count", len(shops)))

	for _, shop := range shops {
		if shop.ConnectionIssueCount >= 3 {
			slog.Warn("skipping shop due to connection issues", "shopId", shop.ID, "count", shop.ConnectionIssueCount)
			continue
		}

		if err := h.scrapeShop(ctx, shop); err != nil {
			slog.Error("failed to scrape shop", "shopId", shop.ID, "error", err)
			h.queries.UpdateShopScrapeError(ctx, queries.UpdateShopScrapeErrorParams{
				LastScrapedError: strPtr(err.Error()),
				ID:               shop.ID,
			})
		}
	}

	return nil
}

// HandleScrape scrapes a single shop by ID from the message payload.
func (h *ShopScrapeHandler) HandleScrape(ctx context.Context, msg ShopScrape) error {
	ctx, span := tracer.Start(ctx, "shop.scrape")
	defer span.End()

	span.SetAttributes(attribute.Int("shop.id", int(msg.ShopID)))
	slog.Info("scraping shop", "shopId", msg.ShopID)

	row, err := h.queries.GetShopForScrape(ctx, msg.ShopID)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("shop %d not found: %w", msg.ShopID, err)
	}

	shop := queries.GetAllShopsRow{
		ID:                   row.ID,
		Name:                 row.Name,
		Url:                  row.Url,
		ClientID:             row.ClientID,
		ClientSecret:         row.ClientSecret,
		ShopwareVersion:      row.ShopwareVersion,
		OrganizationID:       row.OrganizationID,
		Ignores:              row.Ignores,
		LastScrapedAt:        row.LastScrapedAt,
		LastScrapedError:     row.LastScrapedError,
		ConnectionIssueCount: row.ConnectionIssueCount,
		ShopToken:            row.ShopToken,
		SitespeedEnabled:     row.SitespeedEnabled,
		SitespeedUrls:        row.SitespeedUrls,
	}

	if err := h.scrapeShop(ctx, shop); err != nil {
		slog.Error("failed to scrape shop", "shopId", msg.ShopID, "error", err)
		return err
	}

	slog.Info("finished scraping shop", "shopId", msg.ShopID)
	return nil
}

// scrapeShop performs the full scrape flow for a single shop.
func (h *ShopScrapeHandler) scrapeShop(ctx context.Context, shop queries.GetAllShopsRow) error {
	ctx, span := tracer.Start(ctx, "shop.scrape.shop",
		trace.WithAttributes(
			attribute.Int("shop.id", int(shop.ID)),
			attribute.String("shop.name", shop.Name),
			attribute.String("shop.url", shop.Url),
		),
	)
	defer span.End()

	log := slog.With("shopId", shop.ID, "name", shop.Name)

	log.Info("decrypting client secret")
	clientSecret, err := crypto.Decrypt(shop.ClientSecret, h.cfg.AppSecret)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return fmt.Errorf("decrypt client secret: %w", err)
	}

	client := shopware.NewClient(shop.Url, shop.ClientID, clientSecret, shop.ShopToken)

	{
		_, authSpan := tracer.Start(ctx, "shop.scrape.authenticate")
		err := client.Authenticate()
		if err != nil {
			authSpan.RecordError(err)
			authSpan.SetStatus(codes.Error, err.Error())
		}
		authSpan.End()

		if err != nil {
			log.Warn("authentication failed", "error", err)
			h.notifyAuthError(ctx, shop, err)
			h.queries.UpdateShopScrapeError(ctx, queries.UpdateShopScrapeErrorParams{
				LastScrapedError: strPtr(fmt.Sprintf("Authentication failed: %v", err)),
				ID:               shop.ID,
			})
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

	_, fetchSpan := tracer.Start(ctx, "shop.scrape.fetch_data")
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
	log.Info("fetched shop data")
	if fr.configErr != nil {
		errMsg := fmt.Sprintf("failed to fetch shop config: %v", fr.configErr)
		h.notifyDataFetchError(ctx, shop, errMsg)
		h.queries.UpdateShopScrapeError(ctx, queries.UpdateShopScrapeErrorParams{
			LastScrapedError: strPtr(errMsg),
			ID:               shop.ID,
		})
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
			slog.Warn("failed to parse queue data", "shopId", shop.ID, "error", err)
		}
	}

	var cacheInfo shopwareCacheInfo
	if fr.cacheInfoErr == nil {
		if err := json.Unmarshal(fr.cacheInfoData, &cacheInfo); err != nil {
			slog.Warn("failed to parse cache info", "shopId", shop.ID, "error", err)
		}
	}

	if len(extensions) > 0 {
		_, enrichSpan := tracer.Start(ctx, "shop.scrape.enrich_extensions",
			trace.WithAttributes(attribute.Int("extension.count", len(extensions))),
		)
		h.enrichExtensionsFromStore(extensions, shopConfig.Version)
		enrichSpan.End()
	}

	oldExtensions, err := h.queries.GetShopExtensions(ctx, shop.ID)
	if err != nil {
		slog.Warn("failed to get old extensions", "shopId", shop.ID, "error", err)
		oldExtensions = nil
	}

	extensionsDiff := calculateExtensionDiff(oldExtensions, extensions)

	var ignores []string
	if shop.Ignores != nil {
		json.Unmarshal(shop.Ignores, &ignores)
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

	_, checkSpan := tracer.Start(ctx, "shop.scrape.run_checks")
	checkerResult := checker.RunAll(ctx, checkerInput)
	checkSpan.SetAttributes(attribute.String("check.status", string(checkerResult.Status)))
	checkSpan.End()

	newStatus := string(checkerResult.Status)
	h.handleStatusChange(ctx, shop, newStatus)

	ctx, persistSpan := tracer.Start(ctx, "shop.scrape.persist",
		trace.WithAttributes(
			attribute.String("shop.status", newStatus),
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
	defer tx.Rollback(ctx)

	txQueries := h.queries.WithTx(tx)

	if err := txQueries.DeleteShopChecks(ctx, shop.ID); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("delete shop checks: %w", err)
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

		if err := txQueries.UpsertShopExtension(ctx, queries.UpsertShopExtensionParams{
			ShopID:        shop.ID,
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
			return fmt.Errorf("upsert shop extension %s: %w", ext.Name, err)
		}
	}
	if len(extNames) > 0 {
		if err := txQueries.DeleteShopExtensionsNotIn(ctx, queries.DeleteShopExtensionsNotInParams{
			ShopID:  shop.ID,
			Column2: extNames,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("delete stale shop extensions: %w", err)
		}
	}

	taskIDs := make([]string, 0, len(scheduledTasks))
	for _, st := range scheduledTasks {
		taskIDs = append(taskIDs, st.ID)
		overdue := isScheduledTaskOverdue(st)
		if err := txQueries.UpsertShopScheduledTask(ctx, queries.UpsertShopScheduledTaskParams{
			ShopID:            shop.ID,
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
		if err := txQueries.DeleteShopScheduledTasksNotIn(ctx, queries.DeleteShopScheduledTasksNotInParams{
			ShopID:  shop.ID,
			Column2: taskIDs,
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
		if err := txQueries.UpsertShopQueue(ctx, queries.UpsertShopQueueParams{
			ShopID: shop.ID,
			Name:   q.Name,
			Size:   q.Size,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("upsert shop queue %s: %w", q.Name, err)
		}
	}
	if len(queueNames) > 0 {
		if err := txQueries.DeleteShopQueuesNotIn(ctx, queries.DeleteShopQueuesNotInParams{
			ShopID:  shop.ID,
			Column2: queueNames,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("delete stale shop queues: %w", err)
		}
	}

	cacheAdapter := cacheInfo.CacheAdapter
	if cacheAdapter == "" {
		cacheAdapter = "filesystem"
	}
	if err := txQueries.UpsertShopCache(ctx, queries.UpsertShopCacheParams{
		ShopID:       shop.ID,
		Environment:  cacheInfo.Environment,
		HttpCache:    cacheInfo.HttpCache,
		CacheAdapter: cacheAdapter,
	}); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("upsert shop cache: %w", err)
	}

	for _, c := range checkerResult.Checks {
		var link *string
		if c.Link != "" {
			link = &c.Link
		}
		if err := txQueries.InsertShopCheck(ctx, queries.InsertShopCheckParams{
			ShopID:  shop.ID,
			CheckID: c.ID,
			Level:   string(c.Level),
			Message: c.Message,
			Source:  c.Source,
			Link:    link,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("insert shop check %s: %w", c.ID, err)
		}
	}

	hasShopwareUpdate := shop.ShopwareVersion != shopConfig.Version
	hasExtensionChanges := len(extensionsDiff) > 0

	var lastChangelogJSON []byte
	if hasShopwareUpdate {
		lastChangelogJSON, _ = json.Marshal(map[string]interface{}{
			"date": time.Now().Format(time.RFC3339),
			"from": shop.ShopwareVersion,
			"to":   shopConfig.Version,
		})
	}

	if hasExtensionChanges || hasShopwareUpdate {
		extensionsDiffJSON, _ := json.Marshal(extensionsDiff)

		var oldVersion, newVersion *string
		if hasShopwareUpdate {
			oldVersion = &shop.ShopwareVersion
			newVersion = &shopConfig.Version
		}

		shopID := shop.ID
		if err := txQueries.InsertShopChangelog(ctx, queries.InsertShopChangelogParams{
			ShopID:             &shopID,
			Extensions:         extensionsDiffJSON,
			OldShopwareVersion: oldVersion,
			NewShopwareVersion: newVersion,
		}); err != nil {
			persistSpan.RecordError(err)
			persistSpan.SetStatus(codes.Error, err.Error())
			persistSpan.End()
			return fmt.Errorf("insert shop changelog: %w", err)
		}
	}

	favicon := getFavicon(shop.Url)

	if err := txQueries.UpdateShopAfterScrape(ctx, queries.UpdateShopAfterScrapeParams{
		Status:           newStatus,
		ShopwareVersion:  shopConfig.Version,
		LastScrapedError: nil,
		Favicon:          favicon,
		ShopImage:        nil,
		LastChangelog:    lastChangelogJSON,
		ID:               shop.ID,
	}); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("update shop after scrape: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		persistSpan.RecordError(err)
		persistSpan.SetStatus(codes.Error, err.Error())
		persistSpan.End()
		return fmt.Errorf("commit transaction: %w", err)
	}
	persistSpan.End()

	span.SetAttributes(attribute.String("shop.version", shopConfig.Version))
	log.Info("scrape completed", "version", shopConfig.Version)

	if h.bus != nil && shop.SitespeedEnabled && (hasExtensionChanges || hasShopwareUpdate) {
		err := goqueue.Dispatch(ctx, h.bus, SitespeedScrape{ShopID: shop.ID}, goqueue.WithDelay(15*time.Minute))
		if err != nil {
			log.Warn("failed to enqueue sitespeed job", "error", err)
		} else {
			log.Info("enqueued sitespeed job")
		}
	}

	slog.Info("scraped shop", "shopId", shop.ID, "name", shop.Name, "version", shopConfig.Version)
	return nil
}
