# Contributing to Shopmon

Thank you for your interest in contributing to Shopmon! This document provides guidelines and instructions for contributing to the project.

## Table of Contents

- [Getting Started](#getting-started)
- [Development Setup](#development-setup)
- [Making Changes](#making-changes)
- [Code Style](#code-style)
- [Testing](#testing)
- [Submitting Changes](#submitting-changes)
- [Reporting Issues](#reporting-issues)

## Getting Started

1. Fork the repository on GitHub
2. Clone your fork locally:
   ```bash
   git clone https://github.com/your-username/shopmon.git
   cd shopmon
   ```
3. Add the upstream repository as a remote:
   ```bash
   git remote add upstream https://github.com/original-owner/shopmon.git
   ```

## Development Setup

### Prerequisites

- [mise](https://mise.jdx.dev) — manages Go, Node.js, and other tool versions
- Docker and Docker Compose — for running the development infrastructure
- All other tools (Go, Node, sqlc, etc.) are managed by mise and installed automatically

### Initial Setup

1. Trust the mise config and install tools:
   ```bash
   mise trust
   mise install
   ```

2. Install dependencies:
   ```bash
   mise run setup
   ```

3. Start the development infrastructure (PostgreSQL, Redis, Mailpit, demo shop):
   ```bash
   mise run up
   ```

4. Set up the database and seed fixtures:
   ```bash
   mise run load-fixtures
   ```

   This creates test users, organizations, and a shop with sample data.

5. Start the development environment:
   ```bash
   mise run dev
   ```

   This starts three processes in parallel:
   - API server with live reload on http://localhost:5789
   - Background worker with live reload
   - Frontend dev server on http://localhost:3000

### Running Individual Services

```bash
mise run dev:api        # API server only (with air live reload)
mise run dev:worker     # Background worker only (with air live reload)
mise run dev:frontend   # Frontend only (Vite dev server)
```

### Environment Variables

Environment variables for local development are pre-configured in [`.mise.toml`](.mise.toml) under `[env]`. Defaults point to the Docker Compose services (Postgres, Redis, Mailpit, S3-compatible storage).

For production/staging, see [`.env.production`](.env.production) and [`.env.staging`](.env.staging).

## Making Changes

### Branch Naming

Create a new branch for your feature or fix:

```bash
git checkout -b feature/your-feature-name
# or
git checkout -b fix/your-bug-fix
```

### Development Workflow

1. Make your changes in the appropriate directory:
   - API code: `api/`
   - Frontend code: `frontend/`
   - Database queries: `api/sql/queries/`
   - Database migrations: `api/migrations/`
   - API specification: `api/openapi/spec.yaml`

2. Run code quality checks:
   ```bash
   mise run lint
   ```

3. Fix any issues:
   ```bash
   mise run lint-fix
   ```

### API Changes

When adding or modifying API endpoints:

1. Edit `api/openapi/spec.yaml` (add paths, schemas, parameters)
2. Run `mise run generate` (regenerates sqlc queries + oapi-codegen server interface)
3. Implement the new method on the Handler in `api/internal/handler/`
4. Add sqlc queries if needed (`api/sql/queries/`, then `mise run generate`)

### Database Changes

If your changes require database modifications:

1. Add a new migration file in `api/migrations/`
2. Apply the migration:
   ```bash
   mise run migrate
   ```

## Code Style

### API (Go)

- Use `log/slog` for all logging — never `log` or `fmt.Println`
- Always include structured context: `slog.Error("msg", "key", value, "error", err)`
- Wrap errors with context: `fmt.Errorf("create shop: %w", err)`
- Handlers implement the generated `openapi.ServerInterface`
- Use `httputil.WriteJSON()` and `httputil.WriteError()` for all HTTP responses
- Use `r.Context()` for all database calls (propagates OpenTelemetry traces)
- See [AGENTS.md](AGENTS.md) for detailed Go conventions

### Frontend (Vue 3 + TypeScript)

- Use Composition API (`<script setup lang="ts">`)
- Keep components small and reusable
- Use Tailwind CSS for styling
- Run `mise run lint` before committing

### Code Generation

After changing SQL queries or the OpenAPI spec, regenerate code:

```bash
mise run generate
```

This runs `sqlc generate`, `oapi-codegen`, and regenerates frontend API types.

**Never edit files in `api/internal/database/queries/` or `api/openapi/generated/`** — they are auto-generated.

## Testing

### Go Integration Tests

Integration tests use testcontainers (real PostgreSQL + Redis). Run them with:

```bash
mise run test
```

Test helpers are in `api/internal/testutil/`. Use `Setup(t)` to get a `TestEnv` with a seeded database.

### Frontend Tests

```bash
cd frontend && npm run test:run
```

## Submitting Changes

1. Commit your changes with clear, descriptive messages:
   ```bash
   git commit -m "feat: add shop health dashboard"
   # or
   git commit -m "fix: resolve pagination issue in shop list"
   ```

2. Push to your fork:
   ```bash
   git push origin your-branch-name
   ```

3. Create a Pull Request on GitHub:
   - Provide a clear title and description
   - Reference any related issues
   - Include screenshots for UI changes
   - Describe testing performed

### Commit Message Format

We follow conventional commits:

- `feat:` — New features
- `fix:` — Bug fixes
- `docs:` — Documentation changes
- `style:` — Code style changes (formatting, etc.)
- `refactor:` — Code refactoring
- `test:` — Adding or updating tests
- `chore:` — Maintenance tasks

## Reporting Issues

### Bug Reports

When reporting bugs, please include:

1. A clear, descriptive title
2. Steps to reproduce the issue
3. Expected behavior
4. Actual behavior
5. Screenshots (if applicable)
6. Your environment details:
   - OS and version
   - Browser (for frontend issues)
   - Any relevant logs

### Feature Requests

For feature requests, please include:

1. A clear description of the feature
2. Use cases and benefits
3. Any implementation ideas (optional)
4. Mockups or examples (if applicable)

## Questions?

If you have questions about contributing, feel free to:

1. Check existing issues and pull requests
2. Open a discussion issue

Thank you for contributing to Shopmon! 🎉
