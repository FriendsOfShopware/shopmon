package testutil

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	apiserver "github.com/friendsofshopware/shopmon/api/internal/api"
	"github.com/friendsofshopware/shopmon/api/internal/auth"
	"github.com/friendsofshopware/shopmon/api/internal/authapi"
	"github.com/friendsofshopware/shopmon/api/internal/config"
	"github.com/friendsofshopware/shopmon/api/internal/database/queries"
	"github.com/friendsofshopware/shopmon/api/internal/handler"
	"github.com/friendsofshopware/shopmon/api/internal/mail"
	"github.com/friendsofshopware/shopmon/api/internal/middleware"
	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/redis"
	"github.com/testcontainers/testcontainers-go/wait"
)

// TestEnv holds all test dependencies.
type TestEnv struct {
	Pool    *pgxpool.Pool
	Queries *queries.Queries
	Handler *handler.Handler
	Server  *httptest.Server
	Cfg     *config.Config
}

// shared holds the singleton containers shared across tests.
var shared struct {
	once     sync.Once
	pgConn   string
	redisURL string
	err      error
}

func initContainers() {
	ctx := context.Background()

	// Start Postgres
	pgContainer, err := postgres.Run(ctx,
		"postgres:17-alpine",
		postgres.WithDatabase("shopmon_test"),
		postgres.WithUsername("test"),
		postgres.WithPassword("test"),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(30*time.Second),
		),
	)
	if err != nil {
		shared.err = fmt.Errorf("start postgres: %w", err)
		return
	}
	shared.pgConn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		shared.err = fmt.Errorf("postgres conn string: %w", err)
		return
	}

	// Start Redis
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

	// Apply schema once
	pool, err := pgxpool.New(ctx, shared.pgConn)
	if err != nil {
		shared.err = fmt.Errorf("pool for schema: %w", err)
		return
	}
	defer pool.Close()

	schemaSQL, err := os.ReadFile(schemaPath())
	if err != nil {
		shared.err = fmt.Errorf("read schema: %w", err)
		return
	}
	if _, err := pool.Exec(ctx, string(schemaSQL)); err != nil {
		shared.err = fmt.Errorf("apply schema: %w", err)
		return
	}
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

	pool, err := pgxpool.New(ctx, shared.pgConn)
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}

	// Truncate all data for a clean slate
	truncateAll(t, pool)

	q := queries.New(pool)

	cfg := &config.Config{
		AppSecret:         "test-secret-key-32-bytes-long!!", // 32 bytes for AES-256
		FrontendURL:       "http://localhost:3000",
		MailFrom:          "test@shopmon.io",
		RedisURL:          shared.redisURL,
		WebAuthnRPID:      "localhost",
		WebAuthnRPName:    "Shopmon",
		WebAuthnRPOrigins: []string{"http://localhost:3000"},
	}

	for _, fn := range cfgFn {
		fn(cfg)
	}

	mailSvc := mail.NewService(mail.SMTPConfig{From: "test@shopmon.io"})

	h := handler.New(pool, q, nil, cfg, mailSvc, nil)

	// Build chi router matching production setup
	r := chi.NewRouter()
	authHandler := auth.NewAuthHandler(pool, q, cfg, mailSvc)

	r.Route("/api", func(apiRouter chi.Router) {
		apiRouter.Use(middleware.OptionalAuthMiddleware(q))

		// Auth routes (generated from OpenAPI spec)
		authapi.HandlerWithOptions(authHandler, authapi.ChiServerOptions{
			BaseRouter: apiRouter,
		})
		// Callback routes not in generated spec
		apiRouter.Get("/auth/callback/github", authHandler.GithubCallback)
		apiRouter.Get("/auth/sso/callback/{providerId}", authHandler.SSOCallback)

		// Generated API routes
		apiserver.HandlerWithOptions(h, apiserver.ChiServerOptions{
			BaseRouter: apiRouter,
		})
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
	}
}

// schemaPath returns the absolute path to sql/schema.sql relative to this file.
func schemaPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "sql", "schema.sql")
}

func truncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	ctx := context.Background()

	_, err := pool.Exec(ctx, `
		DO $$
		DECLARE r RECORD;
		BEGIN
			FOR r IN (SELECT tablename FROM pg_tables WHERE schemaname = 'public') LOOP
				EXECUTE 'TRUNCATE TABLE ' || quote_ident(r.tablename) || ' CASCADE';
			END LOOP;
		END $$;
	`)
	if err != nil {
		t.Fatalf("failed to truncate tables: %v", err)
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
}

// SeedProject creates a test project in an organization.
func (e *TestEnv) SeedProject(t *testing.T, orgID, name string) int {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	var id int
	err := e.Pool.QueryRow(ctx, `
		INSERT INTO project (organization_id, name, created_at, updated_at)
		VALUES ($1, $2, $3, $3)
		RETURNING id
	`, orgID, name, now).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed project: %v", err)
	}
	return id
}

// SeedShop creates a test shop in an organization/project.
func (e *TestEnv) SeedShop(t *testing.T, orgID string, projectID int, name, url string) int {
	t.Helper()
	ctx := context.Background()
	now := time.Now()

	var id int
	err := e.Pool.QueryRow(ctx, `
		INSERT INTO shop (organization_id, project_id, name, url, client_id, client_secret, shopware_version, shop_token, created_at)
		VALUES ($1, $2, $3, $4, 'test-client', 'test-secret', '6.5.0.0', 'test-shop-token', $5)
		RETURNING id
	`, orgID, projectID, name, url, now).Scan(&id)
	if err != nil {
		t.Fatalf("failed to seed shop: %v", err)
	}
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

// AuthRequest creates a new HTTP request with the session token set.
func (e *TestEnv) AuthRequest(t *testing.T, method, path, token string) *http.Request {
	t.Helper()
	req, err := http.NewRequest(method, e.Server.URL+path, nil)
	if err != nil {
		t.Fatalf("failed to create request: %v", err)
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return req
}
