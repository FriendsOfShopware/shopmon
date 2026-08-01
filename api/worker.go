package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	catalogadvisory "github.com/friendsofshopware/shopmon/api/internal/catalog/advisory"
	catalogchangelog "github.com/friendsofshopware/shopmon/api/internal/catalog/changelog"
	"github.com/friendsofshopware/shopmon/api/internal/catalog/securityplugin"
	catalogsync "github.com/friendsofshopware/shopmon/api/internal/catalog/sync"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	jobspostgres "github.com/friendsofshopware/shopmon/api/internal/jobs/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/maintenance"
	monitoringscrape "github.com/friendsofshopware/shopmon/api/internal/monitoring/scrape"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring/sitespeed"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
	goqueue "github.com/shyim/go-queue"
	queueotel "github.com/shyim/go-queue/middleware/otel"
	"github.com/spf13/cobra"
)

func workerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "worker",
		Short: "Start the background worker",
		RunE:  runWorker,
	}
}

func runWorker(cmd *cobra.Command, args []string) error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	otelShutdown := telemetry.Setup(ctx, cfg.OtelServiceName+"-worker", cfg.OtelServiceVersion, cfg.OtelDeploymentEnv, cfg.OtelTraceEndpoint, cfg.OtelLogEndpoint)
	defer func() {
		if err := otelShutdown(context.Background()); err != nil {
			slog.Error("otel shutdown error", "error", err)
		}
	}()

	// Database
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := queries.New(pool)

	// Mail
	mailSvc, err := mail.NewService(mail.Config{
		DSN:         cfg.MailDSN,
		From:        cfg.MailFrom,
		ReplyTo:     cfg.SMTPReplyTo,
		FrontendURL: cfg.FrontendURL,
	})
	if err != nil {
		return err
	}
	defer func() { _ = mailSvc.Close() }()

	// The worker owns executable handlers; API and fixture processes use the
	// same bus strictly for dispatch.
	bus, err := jobs.NewBus(ctx, pool, jobs.BusConfig{OTelEnabled: cfg.OtelEnabled})
	if err != nil {
		return err
	}

	storeSync := catalogsync.NewService(pool, q, cfg)
	jobDispatcher := jobs.NewDispatcher(bus)
	environmentScrape := monitoringscrape.NewService(pool, q, cfg, mailSvc, storeSync, jobDispatcher)
	sitespeedScrape := sitespeed.NewService(q, cfg)
	cleanup := maintenance.NewService(q)
	changelog := catalogchangelog.NewService(q, cfg)
	advisorySync := catalogadvisory.NewService(pool, q, catalogadvisory.Config{
		GithubToken: cfg.GithubToken,
		NVDAPIKey:   cfg.NVDAPIKey,
		Mailer:      mailSvc,
	})
	securityPluginSync := securityplugin.NewService(q, shopwareaccount.NewClient(
		cfg.ShopwareAPIURL, httputil.NewHTTPClient(httputil.WithTimeout(30*time.Second))))
	if err := jobs.RegisterHandlers(bus, jobs.Handlers{
		EnvironmentScraper:         environmentScrape,
		StoreExtensionSynchronizer: storeSync,
		SitespeedScraper:           sitespeedScrape,
		Cleanup:                    cleanup,
		ChangelogSynchronizer:      changelog,
		AdvisorySynchronizer:       advisorySync,
		SecurityPluginSynchronizer: securityPluginSync,
	}); err != nil {
		return err
	}

	scheduleRepository := jobspostgres.NewScheduleRepository(q)
	scheduler, err := jobs.NewScheduler(scheduleRepository, jobDispatcher, jobs.SchedulerConfig{
		SitespeedEnabled: cfg.SitespeedEndpoint != "" && cfg.SitespeedAPIKey != "",
	})
	if err != nil {
		return err
	}
	scheduler.Start()

	// Populate the Shopware version cache immediately on startup so the data is
	// available without waiting for the first hourly tick. Only versions not yet
	// stored are fetched, so this is cheap once the cache is warm.
	if err := jobDispatcher.EnqueueShopwareChangelogSync(ctx); err != nil {
		slog.Error("failed to dispatch initial shopware changelog sync", "error", err)
	}
	// The backport map is dispatched before the advisory sync so a cold instance
	// resolves advisories against a populated map rather than reporting every one
	// as unknown on first boot.
	if err := jobDispatcher.EnqueueSecurityPluginSync(ctx); err != nil {
		slog.Error("failed to dispatch initial security plugin sync", "error", err)
	}
	// Same for Packagist advisories — first boot should not wait for :17.
	if err := jobDispatcher.EnqueueComposerAdvisorySync(ctx); err != nil {
		slog.Error("failed to dispatch initial composer advisory sync", "error", err)
	}

	// Worker
	worker := goqueue.NewWorker(bus, goqueue.WorkerConfig{
		Concurrency:     10,
		HandlerTimeout:  10 * time.Minute,
		ShutdownTimeout: 1 * time.Minute,
		RetryStrategy: goqueue.RetryStrategy{
			MaxRetries: 3,
			Delay:      5 * time.Second,
			Multiplier: 2.0,
			MaxDelay:   1 * time.Minute,
		},
		Middleware: []goqueue.Middleware{
			queueotel.Middleware(
				queueotel.WithSpanNameNormalizer(queueotel.DefaultSpanNameNormalizer),
			), // tracing
			queueotel.MetricsMiddleware(), // metrics
		},
		ErrorHandler: func(ctx context.Context, env *goqueue.Envelope, err error) {
			slog.Error("job failed", "type", env.Type, "error", err)
		},
	})

	workerDone := make(chan struct{})
	go func() {
		defer close(workerDone)
		slog.Info("starting worker")
		if err := worker.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("worker error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down worker, draining in-flight jobs...")

	// Stop scheduling new tasks and wait for any active schedule to finish.
	schedulerCtx := scheduler.Stop()

	// worker.Run drains in-flight jobs (bounded by ShutdownTimeout) once ctx is
	// cancelled. Wait for it to finish before deferred resources (pool, otel) are
	// torn down, with an outer bound so shutdown cannot hang forever.
	drainCtx, drainCancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer drainCancel()

	select {
	case <-workerDone:
		slog.Info("worker drained")
	case <-drainCtx.Done():
		slog.Warn("worker drain timed out")
	}

	select {
	case <-schedulerCtx.Done():
	case <-drainCtx.Done():
		slog.Warn("scheduler drain timed out")
	}

	slog.Info("worker stopped")

	return nil
}
