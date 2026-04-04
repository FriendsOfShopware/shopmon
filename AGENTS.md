# Project Architecture & Development Guidelines

## Tech Stack

- **API**: Go (chi router, sqlc, oapi-codegen, go-queue)
- **Frontend**: Vue.js + TypeScript
- **Database**: PostgreSQL (pgx)
- **Queue**: PostgreSQL (go-queue)
- **Storage**: S3-compatible (deployment outputs)
- **Observability**: OpenTelemetry (traces + logs via OTLP)

## Directory Structure

```
api/                          <-- Go API (single binary)
  main.go                     <-- Cobra CLI entry point
  server.go                   <-- HTTP server command
  worker.go                   <-- Background worker command
  migrate.go                  <-- Database migration command
  fixtures.go                 <-- Test fixture seeding
  internal/
    auth/                     <-- Authentication (credentials, OAuth, SSO, passkeys, orgs, admin)
    config/                   <-- Environment configuration
    crypto/                   <-- AES-GCM encryption
    database/queries/         <-- sqlc-generated data access (DO NOT EDIT)
    handler/                  <-- API endpoint handlers (implements OpenAPI ServerInterface)
    httputil/                 <-- Shared HTTP helpers (WriteJSON, WriteError, ExtractToken)
    jobs/                     <-- Background jobs (environment scrape, sitespeed, cleanup)
    mail/                     <-- SMTP service + email templates
    middleware/               <-- HTTP middleware (auth, org membership, environment access)
    queue/                    <-- Queue task type definitions
    shopware/                 <-- Shopware HTTP client
      checker/                <-- Environment health check system
    storage/                  <-- S3 storage for deployment outputs
    telemetry/                <-- OpenTelemetry tracing + logging setup
    testutil/                 <-- Test infrastructure (testcontainers for Postgres + Redis)
    webui/                    <-- Embedded frontend serving
  migrations/                 <-- SQL migration files (golang-migrate)
  openapi/
    spec.yaml                 <-- OpenAPI 3.0.3 specification (source of truth for API)
    generated/                <-- oapi-codegen generated server interface (DO NOT EDIT)
  sql/
    schema.sql                <-- Full DDL for sqlc
    queries/                  <-- sqlc query definitions
frontend/                     <-- Vue.js frontend
```

## Code Conventions

### Error Handling

- Wrap errors with context: `fmt.Errorf("create shop: %w", err)`
- In handlers, log with `slog.Error` then respond with `httputil.WriteError`
- In background jobs, record errors on the OTel span AND return them
- **Never** silently discard errors — at minimum use `_ =` for intentional ignoring (e.g. `defer func() { _ = resp.Body.Close() }()`)

### Logging

- Use `log/slog` exclusively (no `log`, no `fmt.Println`)
- Always include structured context: `slog.Error("msg", "shopId", id, "error", err)`
- Logs are automatically exported to OTLP when telemetry is enabled (dual stderr + OTLP output)

### Observability (OpenTelemetry)

- HTTP server traces are handled by `otelhttp` middleware (automatic)
- Background jobs get traces via `go-queue` OTel middleware (automatic)
- For **manual spans** in jobs or complex operations:
  ```go
  var tracer = otel.Tracer("shopmon/jobs")

  func (h *Handler) DoWork(ctx context.Context) error {
      ctx, span := tracer.Start(ctx, "operation.name")
      defer span.End()

      span.SetAttributes(attribute.Int("key", value))

      if err != nil {
          span.RecordError(err)
          span.SetStatus(codes.Error, err.Error())
          return fmt.Errorf("operation: %w", err)
      }
      return nil
  }
  ```
- **Errors that are logged with `slog.Error` automatically appear in OTel** — no extra work needed

### Handler Pattern

Handlers implement the generated `openapi.ServerInterface`:

```go
func (h *Handler) GetThing(w http.ResponseWriter, r *http.Request, id string) {
    user := h.requireUser(w, r)       // returns nil + writes 401 if unauthenticated
    if user == nil {
        return
    }

    thing, err := h.queries.GetThing(r.Context(), id)
    if err != nil {
        slog.Error("failed to get thing", "id", id, "error", err)
        httputil.WriteError(w, http.StatusInternalServerError, "failed to get thing")
        return
    }

    httputil.WriteJSON(w, http.StatusOK, mapToResponse(thing))
}
```

Key rules:
- Use `h.requireUser()` / `h.requireOrgMembership()` for auth — they handle writing error responses
- Use `httputil.WriteJSON()` and `httputil.WriteError()` for all responses
- Use `r.Context()` for all database calls (propagates traces)

### Database Access (sqlc)

- All queries live in `sql/queries/*.sql` — one file per domain (shop.sql, user.sql, etc.)
- Generated Go code is in `internal/database/queries/` — **never edit generated files**
- To add/change a query: edit the `.sql` file, then run `make generate`
- Query naming convention: `-- name: VerbNoun :one|:many|:exec`

### Background Jobs

- Job message types are plain structs (serialized as JSON by go-queue)
- Handlers follow: `func (h *Handler) HandleX(ctx context.Context, msg MsgType) error`
- Always create OTel spans for top-level job handlers
- Register handlers in `internal/jobs/register.go`

### Testing

- Integration tests using testcontainers (real Postgres + Redis)
- Test helpers in `internal/testutil/` — `Setup(t)` returns a `TestEnv` with seeded DB
- Seed data with `env.SeedUser()`, `env.SeedOrganization()`, `env.SeedShop()`, etc.
- Run tests: `make test`

### API Changes

When adding or modifying API endpoints:

1. Edit `openapi/spec.yaml` (add paths, schemas, parameters)
2. Run `make generate` (regenerates server interface)
3. Implement the new method on `Handler` in `internal/handler/`
4. Add sqlc queries if needed (`sql/queries/`, then `make generate`)

## CLI Commands

```
shopmon server              # Start HTTP API server
shopmon worker              # Start background worker
shopmon migrate up          # Apply database migrations
shopmon migrate down        # Rollback last migration
shopmon migrate status      # Show current migration version
shopmon fixtures            # Seed test data
shopmon fixtures --skip-shop # Seed without shop data
```

## Development

```bash
make up              # Start infrastructure (Postgres, Redis, demo shop, Mailpit)
make migrate         # Apply migrations
make load-fixtures   # Reset DB + migrate + seed fixtures
make dev             # Run API server + frontend
make dev-worker      # Run background worker
make test            # Run integration tests
make generate        # Regenerate sqlc + oapi-codegen
```
