package main

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	apiserver "github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/friendsofshopware/shopmon/api/internal/storage"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
	"github.com/spf13/cobra"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

//go:embed openapi/spec.yaml
var openapiSpec []byte

func serverCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "server",
		Short: "Start the HTTP API server",
		RunE:  runServer,
	}
}

func runServer(cmd *cobra.Command, args []string) error {
	cfg := config.Load()

	ctx, stop := signal.NotifyContext(cmd.Context(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// OpenTelemetry
	otelShutdown := telemetry.Setup(ctx, cfg.OtelServiceName, cfg.OtelTraceEndpoint, cfg.OtelLogEndpoint)
	defer otelShutdown(context.Background())

	// Database
	pool, err := database.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	q := queries.New(pool)

	// S3 Storage
	var s3 *storage.S3Storage
	if cfg.S3Endpoint != "" {
		s3, err = storage.NewS3Storage(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region)
		if err != nil {
			return err
		}
	}

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

	// Handler
	h := handler.New(pool, q, s3, cfg, mailSvc, bus)

	// Router
	r := chi.NewRouter()
	if cfg.OtelEnabled {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewMiddleware("",
				otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
					rctx := chi.RouteContext(r.Context())
					if rctx != nil && rctx.RoutePattern() != "" {
						return r.Method + " " + rctx.RoutePattern()
					}
					return r.Method + " " + r.URL.Path
				}),
			)(next)
		})
	}
	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)
	r.Use(chimiddleware.RealIP)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Auth handler
	authHandler := auth.NewAuthHandler(pool, q, cfg, mailSvc)

	// API routes
	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.OptionalAuthMiddleware(pool))

		// OpenAPI spec & docs
		apiRouter.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			w.Write(openapiSpec)
		})
		apiRouter.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(`<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Shopmon API Docs</title>
</head>
<body style="margin:0;padding:0;height:100vh">
  <elements-api
    apiDescriptionUrl="/api/openapi.yaml"
    router="hash"
    layout="sidebar"
    tryItCredentialsPolicy="include"
  />
  <script src="https://unpkg.com/@stoplight/elements/web-components.min.js"></script>
  <link rel="stylesheet" href="https://unpkg.com/@stoplight/elements/styles.min.css">
</body>
</html>`))
		})

		// Auth routes (generated from OpenAPI spec)
		authapi.HandlerWithOptions(authHandler, authapi.ChiServerOptions{
			BaseRouter: apiRouter,
			Middlewares: []authapi.MiddlewareFunc{
				auth.RateLimitMiddleware(auth.NewRateLimiter(ctx, 60*time.Second, 20)),
			},
		})
		// Callback routes not in generated spec
		apiRouter.Get("/auth/callback/github", authHandler.GithubCallback)
		apiRouter.Get("/auth/sso/callback/{providerId}", authHandler.SSOCallback)

		// Generated API routes
		apiserver.HandlerWithOptions(h, apiserver.ChiServerOptions{
			BaseRouter: apiRouter,
		})
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: r,
	}

	go func() {
		slog.Info("starting server", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "error", err)
		}
	}()

	<-ctx.Done()
	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
