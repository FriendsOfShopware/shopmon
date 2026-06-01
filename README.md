# Shopmon

Shopmon is an application from FriendsOfShopware to manage multiple Shopware instances. It tracks versions, extensions, scheduled tasks, queue health, and performance trends across all your shops.

## Tech Stack

- **API**: Go (chi router, sqlc, oapi-codegen)
- **Frontend**: Vue 3 + TypeScript + Tailwind CSS
- **Database**: PostgreSQL
- **Queue**: PostgreSQL (go-queue)
- **Storage**: S3-compatible
- **Observability**: OpenTelemetry (traces + logs via OTLP)

## Features

- Overview of all your Shopware instances
- Shopware version and security update tracking
- Installed extensions and available updates
- Scheduled task and queue monitoring
- Daily sitespeed checks to catch performance regressions
- Cache clearing

## Managed / SaaS

https://shopmon.fos.gg

## Local Development

### Prerequisites

- [mise](https://mise.jdx.dev) — manages Go, Node.js, and tool versions
- [Docker](https://www.docker.com/) — for PostgreSQL, Redis, Mailpit, and demo shop

All other tools (Go 1.26, Node 26, sqlc, golangci-lint, air, oapi-codegen) are installed automatically by mise.

### Quick Start

```bash
git clone https://github.com/FriendsOfShopware/shopmon.git
cd shopmon

mise trust              # trust the mise config
mise install            # install all tools
mise run setup          # install dependencies (Go modules + npm)
mise run up             # start Docker services
mise run load-fixtures  # migrate DB + seed test data
mise run dev            # start API, worker, and frontend
```

Open http://localhost:3000 in your browser.

### Useful Commands

```bash
mise run up              # Start infrastructure (DB, Redis, Mailpit, demo shop)
mise run down            # Stop infrastructure
mise run migrate         # Apply database migrations
mise run fixtures        # Seed test fixtures
mise run load-fixtures   # Drop DB → migrate → seed (full reset)
mise run test            # Run Go integration tests
mise run lint            # Run all linters + type checks + tests
mise run lint-fix        # Auto-fix lint and formatting issues
mise run generate        # Regenerate sqlc + oapi-codegen + frontend API types
mise run build           # Build the API binary
```

Run `mise tasks` to see all available tasks.

### Dev Services

| Service | URL | Notes |
|---|---|---|
| Shopmon Frontend | http://localhost:3000 | |
| Shopmon API | http://localhost:5789 | |
| Mailpit | http://localhost:8025 | Catches all outgoing mail |
| Demo Shop Frontend | http://localhost:3889 | |
| Demo Shop Admin | http://localhost:3889/admin | `admin` / `shopware` |
| Jaeger (traces) | http://localhost:16686 | OpenTelemetry traces |

### Environment Variables

Local dev defaults are in [`.mise.toml`](.mise.toml) under `[env]`. They point to the Docker Compose services. No `.env` file needed for local development.

For production/staging, see [`.env.production`](.env.production) and [`.env.staging`](.env.staging).

## Production Deployment

See [SELF_HOSTING.md](SELF_HOSTING.md) for a complete self-hosting guide with Docker Compose, environment variables, and reverse proxy examples.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and development workflow.

## License

MIT
