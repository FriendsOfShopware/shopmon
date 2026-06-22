package main

import (
	"context"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
	cron "github.com/robfig/cron/v3"
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

	// Queue bus with all handlers registered
	bus, err := jobs.NewBus(ctx, pool, q, cfg, mailSvc)
	if err != nil {
		return err
	}

	// Cron scheduler for recurring tasks.
	// Aggregate scrapes fan out into one per-environment queue task each so the
	// queue can retry and parallelise individual environments independently.
	c := cron.New()
	if _, err := c.AddFunc("0 * * * *", func() {
		ctx := context.Background()
		environments, err := q.GetAllEnvironments(ctx)
		if err != nil {
			slog.Error("failed to list environments for scrape", "error", err)
			return
		}
		for _, env := range environments {
			// Skip environments that have failed to connect repeatedly so we
			// back off instead of hammering an unreachable shop every hour.
			if env.ConnectionIssueCount >= 3 {
				continue
			}
			if err := goqueue.Dispatch(ctx, bus, jobs.EnvironmentScrape{EnvironmentID: env.ID}); err != nil {
				slog.Error("failed to dispatch environment scrape", "environmentId", env.ID, "error", err)
			}
		}
	}); err != nil {
		slog.Error("failed to add environment scrape cron", "error", err)
	}
	if _, err := c.AddFunc("0 3 * * *", func() {
		if cfg.SitespeedEndpoint == "" || cfg.SitespeedAPIKey == "" {
			return
		}
		ctx := context.Background()
		environments, err := q.GetEnvironmentsWithSitespeedEnabled(ctx)
		if err != nil {
			slog.Error("failed to list environments for sitespeed scrape", "error", err)
			return
		}
		for _, env := range environments {
			if err := goqueue.Dispatch(ctx, bus, jobs.SitespeedScrape{EnvironmentID: env.ID}); err != nil {
				slog.Error("failed to dispatch sitespeed scrape", "environmentId", env.ID, "error", err)
			}
		}
	}); err != nil {
		slog.Error("failed to add sitespeed scrape cron", "error", err)
	}
	if _, err := c.AddFunc("0 4 * * *", func() {
		if err := goqueue.Dispatch(context.Background(), bus, jobs.LockCleanup{}); err != nil {
			slog.Error("failed to dispatch lock cleanup", "error", err)
		}
	}); err != nil {
		slog.Error("failed to add lock cleanup cron", "error", err)
	}
	if _, err := c.AddFunc("0 5 * * *", func() {
		if err := goqueue.Dispatch(context.Background(), bus, jobs.InvitationCleanup{}); err != nil {
			slog.Error("failed to dispatch invitation cleanup", "error", err)
		}
	}); err != nil {
		slog.Error("failed to add invitation cleanup cron", "error", err)
	}
	if _, err := c.AddFunc("30 4 * * *", func() {
		if err := goqueue.Dispatch(context.Background(), bus, jobs.OldDataCleanup{}); err != nil {
			slog.Error("failed to dispatch old data cleanup", "error", err)
		}
	}); err != nil {
		slog.Error("failed to add old data cleanup cron", "error", err)
	}
	c.Start()

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

	// Stop scheduling new cron tasks and wait for any running cron job to finish.
	cronCtx := c.Stop()

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
	case <-cronCtx.Done():
	case <-drainCtx.Done():
		slog.Warn("cron drain timed out")
	}

	slog.Info("worker stopped")

	return nil
}
