// Package testdb provides a shared PostgreSQL testcontainer with the shopmon
// schema applied. It is separate from testutil so packages that testutil itself
// depends on (e.g. internal/jobs) can run database-backed tests without an
// import cycle.
package testdb

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

// shared holds the singleton container shared across tests of one package binary.
var shared struct {
	once sync.Once
	conn string
	err  error
}

func initContainer() {
	ctx := context.Background()

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
	shared.conn, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		shared.err = fmt.Errorf("postgres conn string: %w", err)
		return
	}

	pool, err := pgxpool.New(ctx, shared.conn)
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
	}
}

// ConnString returns the connection string of the shared container, starting it
// on first use.
func ConnString(t *testing.T) string {
	t.Helper()
	shared.once.Do(initContainer)
	if shared.err != nil {
		t.Fatalf("failed to start postgres container: %v", shared.err)
	}
	return shared.conn
}

// Setup returns a pool to the shared database with all tables truncated, so
// each test starts from a clean slate. The pool is closed on test cleanup.
func Setup(t *testing.T) *pgxpool.Pool {
	t.Helper()

	pool, err := pgxpool.New(context.Background(), ConnString(t))
	if err != nil {
		t.Fatalf("failed to create pool: %v", err)
	}
	t.Cleanup(pool.Close)

	TruncateAll(t, pool)
	return pool
}

// TruncateAll empties every table in the public schema.
func TruncateAll(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	_, err := pool.Exec(context.Background(), `
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

// schemaPath returns the absolute path to sql/schema.sql relative to this file.
func schemaPath() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "..", "..", "sql", "schema.sql")
}
