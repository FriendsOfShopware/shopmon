package testutil

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/friendsofshopware/shopmon/api/internal/apirouter"
	"github.com/friendsofshopware/shopmon/api/internal/audit"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/catalog"
	catalogshopware "github.com/friendsofshopware/shopmon/api/internal/catalog/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/deployment"
	deploymentqueue "github.com/friendsofshopware/shopmon/api/internal/deployment/queue"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/friendsofshopware/shopmon/api/internal/httputil"
	"github.com/friendsofshopware/shopmon/api/internal/identity"
	credentialmail "github.com/friendsofshopware/shopmon/api/internal/identity/credentialmail"
	identitypasskey "github.com/friendsofshopware/shopmon/api/internal/identity/passkey"
	"github.com/friendsofshopware/shopmon/api/internal/identity/redisstate"
	"github.com/friendsofshopware/shopmon/api/internal/jobs"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/friendsofshopware/shopmon/api/internal/monitoring"
	monitoringqueue "github.com/friendsofshopware/shopmon/api/internal/monitoring/queue"
	monitoringshopware "github.com/friendsofshopware/shopmon/api/internal/monitoring/shopware"
	"github.com/friendsofshopware/shopmon/api/internal/notification"
	"github.com/friendsofshopware/shopmon/api/internal/organization"
	organizationmail "github.com/friendsofshopware/shopmon/api/internal/organization/invitationmail"
	organizationsso "github.com/friendsofshopware/shopmon/api/internal/organization/sso"
	ssooidc "github.com/friendsofshopware/shopmon/api/internal/organization/sso/oidc"
	"github.com/friendsofshopware/shopmon/api/internal/packagesmirror"
	packagesapi "github.com/friendsofshopware/shopmon/api/internal/packagesmirror/packagesapi"
	accountread "github.com/friendsofshopware/shopmon/api/internal/readmodel/account"
	adminread "github.com/friendsofshopware/shopmon/api/internal/readmodel/admin"
	advisoryread "github.com/friendsofshopware/shopmon/api/internal/readmodel/advisory"
	environmentread "github.com/friendsofshopware/shopmon/api/internal/readmodel/environment"
	"github.com/friendsofshopware/shopmon/api/internal/shopwareaccount"
	"github.com/friendsofshopware/shopmon/api/internal/suppression"
	"github.com/friendsofshopware/shopmon/api/internal/testutil/testdb"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	goredis "github.com/redis/go-redis/v9"
	"github.com/shyim/go-mailer/mailertest"
	goqueue "github.com/shyim/go-queue"
	"github.com/testcontainers/testcontainers-go/modules/redis"
)

// TestEnv holds all test dependencies.
type TestEnv struct {
	Pool    *pgxpool.Pool
	Queries *queries.Queries
	Handler *handler.Handler
	Server  *httptest.Server
	Cfg     *config.Config
	// Mail captures every email sent during the test, in place of real
	// delivery. Use the mailertest assertion helpers against it.
	Mail *mailertest.RecordingTransport
}

// shared holds the singleton Redis container shared across tests. Postgres is
// provided by the testdb package.
var shared struct {
	once     sync.Once
	redisURL string
	err      error
}

func initContainers() {
	ctx := context.Background()

	redisContainer, err := redis.Run(ctx, "redis:7-alpine")
	if err != nil {
		shared.err = fmt.Errorf("start redis: %w", err)
		return
	}
	redisEndpoint, err := redisContainer.ConnectionString(ctx)
	if err != nil {
		shared.err = fmt.Errorf("redis conn string: %w", err)
		return
	}
	shared.redisURL = redisEndpoint
}

// Setup creates a new TestEnv using shared Postgres and Redis containers.
// Each test gets a clean database via TRUNCATE.
// Optional cfgFn modifiers are applied to the config before building the handler.
func Setup(t *testing.T, cfgFn ...func(*config.Config)) *TestEnv {
	t.Helper()
	ctx := context.Background()

	shared.once.Do(initContainers)
	if shared.err != nil {
		t.Fatalf("failed to start containers: %v", shared.err)
	}

	pool, err := pgxpool.New(ctx, testdb.ConnString(t))
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Truncate all data for a clean slate
	testdb.TruncateAll(t, pool)
	flushRedis(t, shared.redisURL)

	q := queries.New(pool)

	cfg := &config.Config{
		AppSecret:         "test-secret-key-32-bytes-long!!!", // 32 bytes for AES-256
		FrontendURL:       "http://localhost:3000",
		MailFrom:          "test@shopmon.io",
		RedisURL:          shared.redisURL,
		WebAuthnRPID:      "localhost",
		WebAuthnRPName:    "Shopmon",
		WebAuthnRPOrigins: []string{"http://localhost:3000"},
		// Production always mounts the auth rate limiter. Use a high cap so
		// ordinary tests are not 429'd; tight-cap tests override this.
		AuthRateLimitMax: 10_000,
	}

	for _, fn := range cfgFn {
		fn(cfg)
	}

	mailRecorder := mailertest.NewRecordingTransport("")
	mailSender, err := mail.NewServiceWithTransport(mailRecorder, "noreply@shopmon.test", "", cfg.FrontendURL)
	if err != nil {
		t.Fatalf("create mail service: %v", err)
	}

	bus := goqueue.NewBus()
	// Handlers that enqueue work only need a transport: jobs.Dispatch supplies
	// the explicit route, so tests do not construct fake worker handlers.
	bus.AddTransport(jobs.TransportName, &noopTransport{})
	bus.AddTransport(jobs.ScrapeTransportName, &noopTransport{})
	organizationAuthorizer := organization.NewAuthorizer(q)
	organizationMailer := organizationmail.New(mailSender, cfg.FrontendURL)
	organizationService := organization.NewService(pool, q, organizationAuthorizer, organizationMailer)
	monitoringGateway := monitoringshopware.NewGateway(cfg.AppSecret)
	monitoringDispatcher := monitoringqueue.NewDispatcher(bus)
	monitoringService := monitoring.NewService(pool, q, organizationAuthorizer, monitoringGateway, monitoringDispatcher, nil)
	notificationService := notification.NewService(q)
	packagesClient := packagesapi.NewClient(cfg.PackagesAPIURL, cfg.PackagesAPIToken, httputil.NewHTTPClient(httputil.WithTimeout(30*time.Second)))
	packagesService := packagesmirror.NewService(q, organizationAuthorizer, packagesClient, packagesmirror.Config{
		BaseURL: cfg.PackagesAPIURL,
		Token:   cfg.PackagesAPIToken,
	})
	deploymentDispatcher := deploymentqueue.NewDispatcher(bus)
	deploymentService := deployment.NewService(pool, q, organizationAuthorizer, nil, deploymentDispatcher, cfg.FrontendURL, cfg.DeploymentScrapeDelay)
	catalogGateway := catalogshopware.NewGateway(shopwareaccount.NewClient(cfg.ShopwareAPIURL, httputil.NewHTTPClient(httputil.WithTimeout(30*time.Second))))
	catalogService := catalog.NewService(q, catalogGateway)
	accountReadModel := accountread.NewService(q)
	adminReadModel := adminread.NewService(q)
	advisoryReadModel := advisoryread.NewService(q)
	suppressionService := suppression.NewService(q, organizationAuthorizer)
	environmentReadModel := environmentread.NewService(q, organizationAuthorizer, cfg)
	jobDispatcher := jobs.NewDispatcher(bus)
	ssoGateway := ssooidc.NewGateway(15 * time.Second)
	ssoService := organizationsso.NewService(q, organizationAuthorizer, ssoGateway)
	identityAccountService := identity.NewAccountService(q)
	identityAdminService := identity.NewAdminService(pool, q)
	sessionService := identity.NewSessionService(q)
	credentialMailer := credentialmail.New(mailSender, cfg.FrontendURL)
	credentialService := identity.NewCredentialService(pool, q, sessionService, credentialMailer, identity.CredentialConfig{
		RegistrationEnabled: !cfg.DisableRegistration,
	})
	federatedService := identity.NewFederatedService(pool, q)
	challengeStore, err := redisstate.New(cfg.RedisURL, "shopmon:challenge:")
	if err != nil {
		t.Fatalf("create identity state store: %v", err)
	}
	t.Cleanup(func() { _ = challengeStore.Close() })
	passkeyService, err := identitypasskey.NewService(identitypasskey.Config{
		RPID: cfg.WebAuthnRPID, RPName: cfg.WebAuthnRPName, RPOrigins: cfg.WebAuthnRPOrigins,
	}, q, challengeStore, sessionService)
	if err != nil {
		t.Fatalf("create passkey service: %v", err)
	}
	auditService := audit.NewService(q)
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
		Advisories:    advisoryReadModel,
		Suppressions:  suppressionService,
		AdvisorySync:  jobDispatcher,
	})

	// Build chi router matching production setup
	r := chi.NewRouter()
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
	sessionAuthenticator := identity.NewSessionAuthenticator(q)

	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.OptionalAuthMiddleware(sessionAuthenticator))

		opts := apirouter.Options{}
		if cfg.AuthRateLimitMax > 0 {
			opts.AuthRateLimit = auth.RateLimitMiddleware(auth.NewRateLimiter(t.Context(), time.Minute, cfg.AuthRateLimitMax))
		}
		apirouter.Mount(apiRouter, h, authHandler, opts)
	})

	srv := httptest.NewServer(r)

	t.Cleanup(func() {
		srv.Close()
		pool.Close()
	})

	return &TestEnv{
		Pool:    pool,
		Queries: q,
		Handler: h,
		Server:  srv,
		Cfg:     cfg,
		Mail:    mailRecorder,
	}
}

// flushRedis clears the shared Redis instance so each test starts with an empty cache.
func flushRedis(t *testing.T, redisURL string) {
	t.Helper()

	opts, err := goredis.ParseURL(redisURL)
	if err != nil {
		t.Fatalf("failed to parse redis url: %v", err)
	}

	client := goredis.NewClient(opts)
	defer func() { _ = client.Close() }()

	if err := client.FlushAll(context.Background()).Err(); err != nil {
		t.Fatalf("failed to flush redis: %v", err)
	}
}

// SeedUser creates a test user and session, returning the session token.
func (e *TestEnv) SeedUser(t *testing.T, id, name, email, role string) string {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	_, err := e.Pool.Exec(ctx, `
		INSERT INTO "user" (id, name, email, email_verified, created_at, updated_at, role, notifications)
		VALUES ($1, $2, $3, true, $4, $4, $5, '[]'::jsonb)
	`, id, name, email, now, role)
	if err != nil {
		t.Fatalf("failed to seed user: %v", err)
	}

	token := fmt.Sprintf("test-token-%s", id)
	_, err = e.Pool.Exec(ctx, `
		INSERT INTO session (id, expires_at, token, created_at, updated_at, user_id)
		VALUES ($1, $2, $3, $4, $4, $5)
	`, fmt.Sprintf("session-%s", id), now.Add(24*time.Hour), token, now, id)
	if err != nil {
		t.Fatalf("failed to seed session: %v", err)
	}

	return token
}

// SeedOrganization creates a test organization and adds the user as a member.
func (e *TestEnv) SeedOrganization(t *testing.T, orgID, name, slug, userID string) {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	_, err := e.Pool.Exec(ctx, `
		INSERT INTO organization (id, name, slug, created_at)
		VALUES ($1, $2, $3, $4)
	`, orgID, name, slug, now)
	if err != nil {
		t.Fatalf("failed to seed organization: %v", err)
	}

	_, err = e.Pool.Exec(ctx, `
		INSERT INTO member (id, organization_id, user_id, role, created_at)
		VALUES ($1, $2, $3, 'owner', $4)
	`, fmt.Sprintf("member-%s-%s", orgID, userID), orgID, userID, now)
	if err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	// Set as active organization on the user's session
	_, err = e.Pool.Exec(ctx, `
		UPDATE session SET active_organization_id = $1 WHERE user_id = $2 AND active_organization_id IS NULL
	`, orgID, userID)
	if err != nil {
		t.Fatalf("failed to set active organization: %v", err)
	}
}

// SeedMember adds a user to an existing organization with an explicit role.
// SeedOrganization always creates an owner, so this is what makes role-gated
// behaviour (member vs owner/admin) testable.
func (e *TestEnv) SeedMember(t *testing.T, orgID, userID, role string) {
	t.Helper()
	ctx := context.Background()

	_, err := e.Pool.Exec(ctx, `
		INSERT INTO member (id, organization_id, user_id, role, created_at)
		VALUES ($1, $2, $3, $4, $5)
	`, fmt.Sprintf("member-%s-%s", orgID, userID), orgID, userID, role, time.Now())
	if err != nil {
		t.Fatalf("failed to seed member: %v", err)
	}

	_, err = e.Pool.Exec(ctx, `
		UPDATE session SET active_organization_id = $1 WHERE user_id = $2 AND active_organization_id IS NULL
	`, orgID, userID)
	if err != nil {
		t.Fatalf("failed to set active organization: %v", err)
	}
}

// SeedShop creates a test shop in an organization.
func (e *TestEnv) SeedShop(t *testing.T, orgID, name string) int {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	var id int
	err := e.Pool.QueryRow(ctx, `
		INSERT INTO shop (organization_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id
	`, orgID, name, now).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed shop: %v", err)
	}
	return id
}

// SeedEnvironment creates a test environment in an organization/shop.
// It also sets the environment as the shop's default if none is set.
func (e *TestEnv) SeedEnvironment(t *testing.T, orgID string, shopID int, name, url string) int {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	var id int
	err := e.Pool.QueryRow(ctx, `
		INSERT INTO environment (organization_id, shop_id, name, url, client_id, client_secret, shopware_version, environment_token, created_at)
		VALUES ($1, $2, $3, $4, 'test-client', 'test-secret', '6.5.0.0', 'test-environment-token', $5)
		RETURNING id
	`, orgID, shopID, name, url, now).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed environment: %v", err)
	}

	// Auto-set as shop default if not yet set
	_, _ = e.Pool.Exec(ctx, `
		UPDATE shop SET default_environment_id = $1 WHERE id = $2 AND default_environment_id IS NULL
	`, id, shopID)

	return id
}

// SeedNotification creates a test notification for a user.
func (e *TestEnv) SeedNotification(t *testing.T, userID, key, level, title, message string) int {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	var id int
	err := e.Pool.QueryRow(ctx, `
		INSERT INTO user_notification (user_id, key, level, title, message, link, read, created_at)
		VALUES ($1, $2, $3, $4, $5, '{"url":"http://example.com","label":"View"}'::jsonb, false, $6)
		RETURNING id
	`, userID, key, level, title, message, now).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed notification: %v", err)
	}
	return id
}

// NewMockShopwareServer creates a mock Shopware API server that handles OAuth token and info endpoints.
func NewMockShopwareServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	mux.HandleFunc("/api/oauth/token", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"access_token":"mock-token","expires_in":600}`))
	})

	mux.HandleFunc("/api/_info/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"version":"6.5.0.0"}`))
	})

	return httptest.NewServer(mux)
}

// noopTransport is an in-memory goqueue.Transport that accepts sends and never
// delivers. It lets handlers dispatch jobs in tests without a real broker.
type noopTransport struct{}

func (noopTransport) Send(context.Context, *goqueue.Envelope) error { return nil }
func (noopTransport) Receive(context.Context) (<-chan *goqueue.Envelope, error) {
	ch := make(chan *goqueue.Envelope)
	close(ch)
	return ch, nil
}
func (noopTransport) Ack(context.Context, *goqueue.Envelope) error        { return nil }
func (noopTransport) Nack(context.Context, *goqueue.Envelope, bool) error { return nil }
func (noopTransport) Retry(context.Context, *goqueue.Envelope) error      { return nil }

// Get performs a context-aware GET, replacing http.Get in tests so requests
// carry the test's context (and satisfy the noctx linter).
func Get(t *testing.T, url string) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	return http.DefaultClient.Do(req)
}

// Post performs a context-aware POST, replacing http.Post in tests.
func Post(t *testing.T, url, contentType string, body io.Reader) (*http.Response, error) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodPost, url, body)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", contentType)
	return http.DefaultClient.Do(req)
}

// NewRequest builds a context-aware request, replacing http.NewRequest in tests.
func NewRequest(t *testing.T, method, url string, body io.Reader) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, url, body)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	return req
}

// AuthRequest creates a new HTTP request with the session token set.
func (e *TestEnv) AuthRequest(t *testing.T, method, path, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, e.Server.URL+path, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}
