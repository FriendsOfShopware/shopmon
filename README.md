# Shopmon

Shopmon is an application from FriendsOfShopware to manage multiple Shopware instances. It tracks versions, extensions, scheduled tasks, queue health, and performance trends across all your shops.

## Tech Stack

- **API**: Go (chi router, Huma v2, sqlc)
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

All other tools (Go 1.27, Node 26, sqlc, golangci-lint, air) are installed automatically by mise.

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

### Docker Sandbox

With [Docker Sandboxes](https://docs.docker.com/ai/sandboxes/configuration/environment-files/)
and the `sbx` CLI installed, use the committed environment from the repository root:

```bash
sbx env plan .                  # Review the environment
sbx env create .                # Install mise, project tools, and dependencies
sbx env exec . -- mise run lint
sbx env exec . -- mise run test
sbx env run .                   # Attach to Codex
```

The sandbox mounts your checkout, so source edits appear on the host. Its own
Docker Engine runs Compose services and testcontainers; Linux frontend dependencies
live in a sandbox volume, separate from the host's `frontend/node_modules`.
Tool versions and development settings come from `.mise.toml`.

To start the application, run these commands in the sandbox (or prefix them with
`sbx env exec . --` on the host):

```bash
mise run up
mise run load-fixtures
mise run dev
```

The frontend, Mailpit, demo shop, Jaeger, and LavinMQ management UI use the same
host ports listed below. Stop any host development services using those ports
before creating the sandbox. The frontend proxies `/api` to the sandbox API.
The root `sbxenv.yaml` stays read-only inside the sandbox. After changing its
ports, resources, or kit, use `sbx env rm .` and create it again; removal also
deletes the sandbox's installed tools, dependencies, and database volumes.

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
mise run generate        # Regenerate sqlc + dump OpenAPI (gitignored) + frontend API types
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
| LavinMQ | http://localhost:15672 | `guest` / `guest`. Only used with `QUEUE_TRANSPORT=amqp` |

### Environment Variables

Local dev defaults are in [`.mise.toml`](.mise.toml) under `[env]`. They point to the Docker Compose services. No `.env` file needed for local development.

For production/staging, see [`.env.production`](.env.production) and [`.env.staging`](.env.staging).

## Production Deployment

See [SELF_HOSTING.md](SELF_HOSTING.md) for a complete self-hosting guide with Docker Compose, environment variables, and reverse proxy examples.

## Contributing

See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines and development workflow.
The Go API's package boundaries and extension workflow are documented in
[api/ARCHITECTURE.md](api/ARCHITECTURE.md).

## License

MIT
