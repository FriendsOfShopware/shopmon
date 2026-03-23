# Project Architecture & Development Guidelines

## Tech Stack

- **API**: Go (chi router, sqlc, oapi-codegen, asynq)
- **Frontend**: Vue.js + TypeScript
- **Database**: PostgreSQL
- **Queue**: Redis (asynq)
- **Storage**: S3-compatible (deployment outputs)

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
    database/queries/         <-- sqlc-generated data access
    handler/                  <-- API endpoint handlers (implements OpenAPI ServerInterface)
    httputil/                 <-- Shared HTTP helpers (WriteJSON, WriteError, ExtractToken)
    jobs/                     <-- Background jobs (shop scrape, sitespeed, cleanup)
    mail/                     <-- SMTP service + email templates
    middleware/               <-- HTTP middleware (auth, org membership, shop access)
    queue/                    <-- Asynq task type definitions
    shopware/                 <-- Shopware HTTP client
      checker/                <-- Shop health check system (env, security, tasks, worker, frosh_tools)
    storage/                  <-- S3 storage for deployment outputs
    telemetry/                <-- OpenTelemetry tracing setup
    testutil/                 <-- Test infrastructure (testcontainers for Postgres + Redis)
  migrations/                 <-- SQL migration files (golang-migrate)
  openapi/
    spec.yaml                 <-- OpenAPI 3.0.3 specification
    generated/                <-- oapi-codegen generated server interface
  sql/
    schema.sql                <-- Full DDL for sqlc
    queries/                  <-- sqlc query definitions
frontend/                     <-- Vue.js frontend
```

## API Architecture

The Go API uses a handler-based architecture with sqlc for type-safe database queries and oapi-codegen for OpenAPI server interface generation.

### Handler (`internal/handler/`)
- Implements the generated `openapi.ServerInterface`
- Uses `requireUser()` and `requireOrgMembership()` helpers for auth checks
- Delegates to sqlc queries for data access
- Returns responses using `httputil.WriteJSON()` / `httputil.WriteError()`

### Auth (`internal/auth/`)
- Standalone auth system (no external auth libraries)
- Email/password with bcrypt, GitHub OAuth, SSO/OIDC, WebAuthn passkeys
- Organization management (CRUD, invitations, roles)
- Admin endpoints (user management, ban/unban, impersonation)
- Redis-backed challenge store for OAuth/WebAuthn flows
- Rate limiting (20 req/60s per IP)

### Background Jobs (`internal/jobs/`)
- Shop scraping: fetches Shopware API data, runs health checkers, computes diffs, sends notifications
- Sitespeed: performance measurement via external sitespeed service
- Cleanup: expired locks and invitations

### Database Queries (`sql/queries/`)
- All database access goes through sqlc-generated code
- Query files organized by domain: shop.sql, user.sql, project.sql, etc.
- Generate with: `cd api && make generate`

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
