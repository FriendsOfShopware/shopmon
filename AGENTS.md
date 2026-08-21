# Project Architecture & Development Guidelines

## Tech Stack

- **API**: Go (chi router, Huma v2, sqlc, go-queue)
- **Frontend**: Vue.js + TypeScript
- **Database**: PostgreSQL (pgx)
- **Queue**: PostgreSQL (go-queue), optionally AMQP (LavinMQ/RabbitMQ) via `QUEUE_TRANSPORT=amqp`
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
    api/                      <-- Hand-written HTTP DTO types (API models)
    authapi/                  <-- Hand-written HTTP DTO types (auth models)
    apirouter/                <-- Huma-on-chi mount + OpenAPI metadata
    auth/                     <-- Authentication (credentials, OAuth, SSO, passkeys, orgs, admin)
    config/                   <-- Environment configuration
    crypto/                   <-- AES-GCM encryption
    database/queries/         <-- sqlc-generated data access (DO NOT EDIT)
    handler/                  <-- API endpoint handlers (Huma operations)
    httputil/                 <-- Shared HTTP helpers + Huma JSON error override
    jobs/                     <-- Background job handlers + task types (environment scrape, sitespeed, cleanup)
    mail/                     <-- SMTP service + email templates
    middleware/               <-- HTTP middleware (auth, org membership, environment access)
    shopware/                 <-- Shopware admin API client (per-shop)
      checker/                <-- Environment health check system
    shopwareaccount/          <-- Shopware account/store API client (api.shopware.com)
    storage/                  <-- S3 storage for deployment outputs
    telemetry/                <-- OpenTelemetry tracing + logging setup
    testutil/                 <-- Test infrastructure (testcontainers for Postgres + Redis)
    webui/                    <-- Embedded frontend serving
  migrations/                 <-- SQL migration files (golang-migrate)
  sql/
    schema.sql                <-- Full DDL for sqlc
    queries/                  <-- sqlc query definitions
frontend/                     <-- Vue.js frontend
```

## Code Conventions

### Error Handling

- Wrap errors with context: `fmt.Errorf("create shop: %w", err)`
- In handlers, log with `slog.Error` then return a Huma status error (`huma.Error404NotFound`, …)
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

Register each operation with `huma.Register` (existing `operationId`) and implement a Huma handler:

```go
func (h *Handler) GetThing(ctx context.Context, input *getThingInput) (*getThingOutput, error) {
    user, err := h.requireUser(ctx) // returns huma.Error401Unauthorized if unauthenticated
    if err != nil {
        return nil, err
    }

    thing, err := h.things.Get(ctx, user.ID, input.ID)
    if err != nil {
        slog.ErrorContext(ctx, "failed to get thing", "id", input.ID, "error", err)
        return nil, huma.Error500InternalServerError("failed to get thing")
    }

    return &getThingOutput{Body: mapToResponse(thing)}, nil
}
```

Key rules:
- Use `h.requireUser()` for auth — it returns a Huma 401 error when unauthenticated
- Return Huma outputs (`Body`, optional `Status`) or `huma.Error*` — error bodies stay `{"message":"..."}`
- Use `ctx` from the Huma handler for all database calls (propagates traces)
- Paths are relative to the `/api` chi mount (`/health`, `/organizations/{orgId}/shops`, …)

### Database Access (sqlc)

- All queries live in `sql/queries/*.sql` — one file per domain (shop.sql, user.sql, etc.)
- Generated Go code is in `internal/database/queries/` — **never edit generated files**
- To add/change a query: edit the `.sql` file, then run `mise run generate`
- Query naming convention: `-- name: VerbNoun :one|:many|:exec`

### Background Jobs

- Job message types are plain structs (serialized as JSON by go-queue)
- Handlers follow: `func (h *Handler) HandleX(ctx context.Context, msg MsgType) error`
- Always create OTel spans for top-level job handlers
- Register handlers in `internal/jobs/register.go`
- The transport is chosen in `internal/jobs/bus.go` (`QUEUE_TRANSPORT`): Postgres by default, AMQP when configured. Job code stays transport-agnostic — dispatch through `jobs.Dispatch`, never against a transport directly

### Testing

- Integration tests using testcontainers (real Postgres + Redis)
- Test helpers in `internal/testutil/` — `Setup(t)` returns a `TestEnv` with seeded DB
- Seed data with `env.SeedUser()`, `env.SeedOrganization()`, `env.SeedShop()`, etc.
- Run tests: `mise run test`

### API Changes

When adding or modifying API endpoints:

1. Register a Huma operation (`huma.Register`) with the `operationId`, path, method, and input/output structs in `internal/handler` or `internal/auth`
2. Implement the handler (return a Huma output or `huma.Error*`)
3. Add or extend DTO types in `internal/api` or `internal/authapi` if needed
4. Add sqlc queries if needed (`sql/queries/`)
5. Run `mise run generate` (dumps OpenAPI to gitignored `api/openapi/spec.yaml` and regenerates committed frontend types)

Do not commit or hand-edit `openapi/spec.yaml` — it is a local generate input. The reviewable client is `frontend/src/api/generated/`.

## CLI Commands

```
shopmon server              # Start HTTP API server
shopmon worker              # Start background worker
shopmon migrate up          # Apply database migrations
shopmon migrate down        # Rollback last migration
shopmon migrate status      # Show current migration version
shopmon fixtures            # Seed test data
shopmon fixtures --skip-shop # Seed without shop data
shopmon openapi             # Print OpenAPI YAML generated from Huma routes
```

## Development

```bash
mise run up              # Start infrastructure (Postgres, Redis, demo shop, Mailpit)
mise run migrate         # Apply migrations
mise run load-fixtures   # Reset DB + migrate + seed fixtures
mise run dev             # Run API server + worker + frontend
mise run dev:worker      # Run background worker only
mise run test            # Run API integration tests
mise run lint            # Lint/format/typecheck API + frontend (includes frontend unit tests)
mise run generate        # Regenerate sqlc + dump OpenAPI from Huma + frontend types
```

## Verification

Before considering a change complete (and before opening or updating a PR), always run:

```bash
mise run lint
mise run test
```

- **`mise run lint`** — API (`golangci-lint`) + frontend (oxlint, `vue-tsc`, oxfmt check, Vitest unit tests)
- **`mise run test`** — API integration tests (`go test ./internal/...`)

Do not rely only on targeted package/file test runs. Fix any failures (including format drift from oxfmt) and re-run both commands until they pass. If format check fails, run `cd frontend && npm run format:fix` (or `npx oxfmt <file>`) and include the formatting fix in the change.
