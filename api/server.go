package main

import (
	"context"
	_ "embed"
	"log/slog"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/apirouter"
	"github.com/friendsofshopware/shopmon/api/internal/audit"
	auditpostgres "github.com/friendsofshopware/shopmon/api/internal/audit/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	catalogpostgres "github.com/friendsofshopware/shopmon/api/internal/catalog/postgres"
	catalogshopware "github.com/friendsofshopware/shopmon/api/internal/catalog/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	deploymentpostgres "github.com/friendsofshopware/shopmon/api/internal/deployment/postgres"
	deploymentqueue "github.com/friendsofshopware/shopmon/api/internal/deployment/queue"
	deployments3 "github.com/friendsofshopware/shopmon/api/internal/deployment/s3output"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	credentialmail "github.com/friendsofshopware/shopmon/api/internal/identity/credentialmail"
	identitypasskey "github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
	passkeypostgres "github.com/friendsofshopware/shopmon/api/internal/identity/passkey/postgres"
	identitypostgres "github.com/friendsofshopware/shopmon/api/internal/identity/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/identity/redisstate"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	monitoringpostgres "github.com/friendsofshopware/shopmon/api/internal/monitoring/postgres"
	monitoringqueue "github.com/friendsofshopware/shopmon/api/internal/monitoring/queue"
	monitoringshopware "github.com/friendsofshopware/shopmon/api/internal/monitoring/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	notificationpostgres "github.com/friendsofshopware/shopmon/api/internal/notification/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationmail "github.com/friendsofshopware/shopmon/api/internal/organization/invitationmail"
	organizationpostgres "github.com/friendsofshopware/shopmon/api/internal/organization/postgres"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
	ssooidc "github.com/friendsofshopware/shopmon/api/internal/organization/sso/oidc"
	ssopostgres "github.com/friendsofshopware/shopmon/api/internal/organization/sso/postgres"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
	packagesapi "github.com/friendsofshopware/shopmon/api/internal/packagesmirror/packagesapi"
	packagespostgres "github.com/friendsofshopware/shopmon/api/internal/packagesmirror/postgres"
	accountread "github.com/friendsofshopware/shopmon/api/internal/readmodel/account"
	adminread "github.com/friendsofshopware/shopmon/api/internal/readmodel/admin"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/storage"
	"github.com/friendsofshopware/shopmon/api/internal/telemetry"
	"github.com/friendsofshopware/shopmon/api/internal/webui"
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
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
	otelShutdown := telemetry.Setup(ctx, cfg.OtelServiceName, cfg.OtelServiceVersion, cfg.OtelDeploymentEnv, cfg.OtelTraceEndpoint, cfg.OtelLogEndpoint)
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

	// S3 Storage
	var s3 *storage.S3Storage
	if cfg.S3Endpoint != "" {
		s3, err = storage.NewS3Storage(cfg.S3Endpoint, cfg.S3AccessKey, cfg.S3SecretKey, cfg.S3Bucket, cfg.S3Region)
		if err != nil {
			return err
		}
	}

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

	// Dispatch-only queue bus. Executable handlers are registered exclusively by
	// the worker process.
	bus, err := jobs.NewBus(ctx, pool, jobs.BusConfig{OTelEnabled: cfg.OtelEnabled})
	if err != nil {
		return err
	}

	organizationRepository := organizationpostgres.NewAuthorizationRepository(q)
	organizationAuthorizer := organization.NewAuthorizer(organizationRepository)
	organizationStore := organizationpostgres.NewRepository(pool, q)
	organizationMailer := organizationmail.New(mailSvc, cfg.FrontendURL)
	organizationService := organization.NewService(organizationStore, organizationAuthorizer, organizationMailer)
	monitoringRepository := monitoringpostgres.NewRepository(pool, q)
	monitoringGateway := monitoringshopware.NewGateway(cfg.AppSecret)
	monitoringDispatcher := monitoringqueue.NewDispatcher(bus)
	var monitoringOutputs monitoring.DeploymentOutputStore
	if s3 != nil {
		monitoringOutputs = s3
	}
	monitoringService := monitoring.NewService(monitoringRepository, organizationAuthorizer, monitoringGateway, monitoringDispatcher, monitoringOutputs)
	notificationRepository := notificationpostgres.NewRepository(q)
	notificationService := notification.NewService(notificationRepository)
	packagesRepository := packagespostgres.NewShopRepository(q)
	packagesClient := packagesapi.NewClient(cfg.PackagesAPIURL, cfg.PackagesAPIToken, httputil.NewHTTPClient(httputil.WithTimeout(30*time.Second)))
	packagesService := packagesmirror.NewService(packagesRepository, organizationAuthorizer, packagesClient, packagesmirror.Config{
		BaseURL: cfg.PackagesAPIURL,
		Token:   cfg.PackagesAPIToken,
	})
	deploymentRepository := deploymentpostgres.NewRepository(pool, q)
	deploymentDispatcher := deploymentqueue.NewDispatcher(bus)
	var deploymentOutputs deployment.OutputStore
	if s3 != nil {
		deploymentOutputs = deployments3.NewStore(s3)
	}
	deploymentService := deployment.NewService(deploymentRepository, organizationAuthorizer, deploymentOutputs, deploymentDispatcher, cfg.FrontendURL, cfg.DeploymentScrapeDelay)
	catalogRepository := catalogpostgres.NewVersionRepository(q)
	catalogGateway := catalogshopware.NewGateway(shopwareaccount.NewClient(cfg.ShopwareAPIURL, httputil.NewHTTPClient(httputil.WithTimeout(30*time.Second))))
	catalogService := catalog.NewService(catalogRepository, catalogGateway)
	accountReadModel := accountread.NewService(q)
	adminReadModel := adminread.NewService(q)
	environmentReadModel := environmentread.NewService(q, organizationAuthorizer, cfg)
	ssoRepository := ssopostgres.NewRepository(q)
	ssoGateway := ssooidc.NewGateway(15 * time.Second)
	ssoService := organizationsso.NewService(ssoRepository, organizationAuthorizer, ssoGateway)
	identityAccountRepository := identitypostgres.NewAccountRepository(q)
	identityAccountService := identity.NewAccountService(identityAccountRepository)
	identityAdminRepository := identitypostgres.NewAdminRepository(pool, q)
	identityAdminService := identity.NewAdminService(identityAdminRepository)
	sessionLifecycleRepository := identitypostgres.NewSessionLifecycleRepository(q)
	sessionService := identity.NewSessionService(sessionLifecycleRepository)
	credentialRepository := identitypostgres.NewCredentialRepository(pool, q)
	credentialMailer := credentialmail.New(mailSvc, cfg.FrontendURL)
	credentialService := identity.NewCredentialService(credentialRepository, sessionService, credentialMailer, identity.CredentialConfig{
		RegistrationEnabled: !cfg.DisableRegistration,
	})
	federatedRepository := identitypostgres.NewFederatedRepository(pool, q)
	federatedService := identity.NewFederatedService(federatedRepository)
	challengeStore, err := redisstate.New(cfg.RedisURL, "shopmon:challenge:")
	if err != nil {
		return err
	}
	defer func() {
		if err := challengeStore.Close(); err != nil {
			slog.Error("failed to close identity state store", "error", err)
		}
	}()
	passkeyRepository := passkeypostgres.NewRepository(q)
	passkeyService, err := identitypasskey.NewService(identitypasskey.Config{
		RPID: cfg.WebAuthnRPID, RPName: cfg.WebAuthnRPName, RPOrigins: cfg.WebAuthnRPOrigins,
	}, passkeyRepository, challengeStore, sessionService)
	if err != nil {
		return err
	}
	auditRepository := auditpostgres.NewRepository(q)
	auditService := audit.NewService(auditRepository)

	// Handler
	h := handler.New(handler.Dependencies{
		Database: pool,
		Features: handler.InstanceFeatures{
			RegistrationEnabled:  !cfg.DisableRegistration,
			GithubAuthEnabled:    cfg.GithubClientID != "" && cfg.GithubClientSecret != "",
			SitespeedEnabled:     cfg.SitespeedEndpoint != "",
			PackageMirrorEnabled: cfg.PackagesAPIURL != "" && cfg.PackagesAPIToken != "",
		},
		Monitoring:    monitoringService,
		Notifications: notificationService,
		Packages:      packagesService,
		Deployments:   deploymentService,
		Catalog:       catalogService,
		SSO:           ssoService,
		Account:       accountReadModel,
		Admin:         adminReadModel,
		Environments:  environmentReadModel,
	})

	// Router
	r := chi.NewRouter()
	if cfg.OtelEnabled {
		r.Use(func(next http.Handler) http.Handler {
			return otelhttp.NewMiddleware("",
				// Only trace API/auth traffic. Everything else is served by the
				// embedded frontend handler (static assets + SPA fallback), which
				// would otherwise create a high-cardinality span per asset URL.
				otelhttp.WithFilter(func(r *http.Request) bool {
					return r.URL.Path == "/api" || strings.HasPrefix(r.URL.Path, "/api/")
				}),
			)(next)
		})
		// Must run after otelhttp so it can decorate the span, but the route
		// pattern is only known once chi has matched (i.e. after next returns).
		r.Use(middleware.RouteTag)
	}
	r.Use(chimiddleware.Logger)
	// Use our own recoverer so panics on API routes return the standard JSON
	// error shape instead of chi's plain-text 500.
	r.Use(middleware.Recoverer)
	if len(cfg.TrustedProxies) > 0 {
		r.Use(chimiddleware.ClientIPFromXFF(cfg.TrustedProxies...))
	} else {
		r.Use(chimiddleware.ClientIPFromRemoteAddr)
	}
	r.Use(middleware.TraceIDHeader)

	// Auth handler
	authHandler := auth.NewAuthHandler(auth.Dependencies{
		Organizations: organizationService,
		SSO:           ssoService,
		Accounts:      identityAccountService,
		AdminUsers:    identityAdminService,
		Credentials:   credentialService,
		Sessions:      sessionService,
		Federated:     federatedService,
		Passkeys:      passkeyService,
		Audit:         auditService,
		State:         challengeStore,
		Config: auth.HandlerConfig{
			FrontendURL:        cfg.FrontendURL,
			GitHubClientID:     cfg.GithubClientID,
			GitHubClientSecret: cfg.GithubClientSecret,
		},
	})
	sessionRepository := identitypostgres.NewSessionRepository(q)
	sessionAuthenticator := identity.NewSessionAuthenticator(sessionRepository)

	// API routes
	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.OptionalAuthMiddleware(sessionAuthenticator))

		// OpenAPI spec & docs
		apiRouter.Get("/openapi.yaml", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/x-yaml")
			_, _ = w.Write(openapiSpec)
		})
		apiRouter.Get("/docs", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(`<!doctype html>
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

		// Register the generated auth + API handlers with consistent JSON error
		// handling (param-binding 400s, 404/405 for unmatched routes). The wiring
		// lives in internal/apirouter so the production server and the test
		// harness share one source of truth.
		//
		// RequireAuth is intentionally NOT applied to the API routes: this
		// sub-router also serves intentionally public endpoints
		// (/api/openapi.yaml, /api/docs, the auth routes and health), so a
		// blanket guard would break them. OptionalAuthMiddleware above populates
		// the user context, and each handler enforces auth via h.requireUser
		// (see internal/handler).
		//
		// To enforce authentication declaratively for an authenticated-only set
		// of endpoints, register them on a dedicated sub-router and pass
		// middleware.RequireAuth through the generated Middlewares option. Note:
		// oapi-codegen's ChiServerOptions.Middlewares applies the same chain to
		// *every* route in that generated handler set (no per-route support), so
		// mixing public and protected endpoints requires splitting into separate
		// generated handlers / sub-routers rather than guarding individual ops.
		apirouter.Mount(apiRouter, h, authHandler, apirouter.Options{
			AuthMiddlewares: []authapi.MiddlewareFunc{
				auth.RateLimitMiddleware(auth.NewRateLimiter(ctx, 60*time.Second, cfg.AuthRateLimitMax)),
			},
		})
	})

	if frontendHandler, ok := webui.Handler(); ok {
		slog.Info("serving embedded frontend")
		r.NotFound(frontendHandler.ServeHTTP)
	} else {
		slog.Warn("no embedded frontend assets were found")
	}

	// Wrap router so /healthz bypasses all middleware (tracing, logging, etc.)
	handler := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path == "/healthz" {
			w.WriteHeader(http.StatusOK)
			return
		}
		r.ServeHTTP(w, req)
	})

	srv := &http.Server{
		Addr:    cfg.ListenAddr,
		Handler: handler,
	}

	serverErr := make(chan error, 1)
	go func() {
		slog.Info("starting server", "addr", cfg.ListenAddr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return srv.Shutdown(shutdownCtx)
}
