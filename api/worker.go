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

	otelShutdown := telemetry.Setup(ctx, cfg.OtelServiceName+"-worker", cfg.OtelTraceEndpoint, cfg.OtelLogEndpoint)
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
	mailSvc := mail.NewService(mail.SMTPConfig{
		Host:    cfg.SMTPHost,
		Port:    cfg.SMTPPort,
		Secure:  cfg.SMTPSecure,
		User:    cfg.SMTPUser,
		Pass:    cfg.SMTPPass,
		From:    cfg.MailFrom,
		ReplyTo: cfg.SMTPReplyTo,
	})

	// Queue bus with all handlers registered
	bus, err := jobs.NewBus(ctx, pool, q, cfg, mailSvc)
	if err != nil {
		return err
	}

	// Cron scheduler for recurring tasks
	c := cron.New()
	if _, err := c.AddFunc("0 * * * *", func() {
		if err := goqueue.Dispatch(context.Background(), bus, jobs.EnvironmentScrapeAll{}); err != nil {
			slog.Error("failed to dispatch environment scrape all", "error", err)
		}
	}); err != nil {
		slog.Error("failed to add environment scrape cron", "error", err)
	}
	if _, err := c.AddFunc("0 3 * * *", func() {
		if err := goqueue.Dispatch(context.Background(), bus, jobs.SitespeedScrapeAll{}); err != nil {
			slog.Error("failed to dispatch sitespeed scrape all", "error", err)
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
	c.Start()

	// Worker
	worker := goqueue.NewWorker(bus, goqueue.WorkerConfig{
		Concurrency:    10,
		HandlerTimeout: 10 * time.Minute,
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

	go func() {
		slog.Info("starting worker")
		if err := worker.Run(ctx); err != nil && err != context.Canceled {
			slog.Error("worker error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down worker...")
	c.Stop()
	slog.Info("worker stopped")

	return nil
}
