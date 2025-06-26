# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shopmon is a SaaS application for monitoring multiple Shopware instances. It tracks shop versions, extensions, security updates, scheduled tasks, and performance metrics.

## Architecture

### API (Bun + TypeScript)
- Located in `/api/`
- Built with TypeScript, tRPC, Hono framework
- Uses Drizzle ORM with SQLite (Cloudflare D1 in production)
- Sentry for error tracking
- Deployed via Docker containers

### Frontend (Vue 3 SPA)
- Located in `/frontend/`
- Vue 3 with TypeScript, Pinia stores, Vue Router
- tRPC client for type-safe API communication
- PostCSS with nested syntax for styling
- Built with Vite

### Database
- SQLite with Drizzle ORM
- Migrations in `/api/drizzle/`
- Schema defined in `/api/src/db.ts`

## Common Commands

### Setup & Development
```bash
# Install dependencies for both API and frontend
make setup

# Run database migrations
make migrate

# Start development servers (API on :5789, Frontend on :3000)
make dev

# Run frontend against production API
make dev-to-prod

# Start demo environment with Docker
make up
```

### API-specific Commands
```bash
cd api
bun run dev               # Development server
bun run db:generate       # Generate database migrations
bun run db:migrate        # Apply migrations
bun run cron              # Run cron jobs manually
bun run build:app         # Build for production
bun run build:cron        # Build cron worker
```

### Database Migration Commands
```bash
# Generate a new migration after schema changes
cd api
bunx --bun drizzle-kit generate

# Apply pending migrations
cd api
bun run db:migrate

# Creating custom SQL migrations
# 1. Run bunx --bun drizzle-kit generate --custom --name=<migration_name>
# 2. Edit the generated file in api/drizzle/migrations/
# 3. Run: bun run db:migrate
```

### Frontend-specific Commands
```bash
cd frontend
bun run dev               # Development server
bun run dev:local         # Dev with local API
bun run build             # Production build
bun run preview           # Preview production build
```

### Code Quality
```bash
# Run from root directory
bun run biome:check       # Lint and format check
bun run biome:fix         # Auto-fix issues
```

## Key Patterns & Conventions

### tRPC Router Structure
- Main router: `/api/src/router.ts`
- Feature routers in `/api/src/trpc/*/index.ts`
- Context and middleware in `/api/src/trpc/`
- Authentication middleware in `/api/src/trpc/middleware.ts`

### Frontend State Management
- Pinia stores in `/frontend/src/stores/`
- Each major feature has its own store
- tRPC client configured in `/frontend/src/helpers/trpc.ts`
- Persistent auth state in localStorage

### Authentication
- JWT-based authentication with database sessions
- Passkey/WebAuthn support in `/api/src/trpc/auth/passkey.ts`
- Session cleanup via cron job
- Password reset with expiring tokens

### Database Schema
- User management: `user`, `user_notification`, `sessions`, `user_passkeys`
- Organization/Shop: `organization`, `shop`, `shop_scrape_info`
- Monitoring data: `shop_changelog`, `shop_sitespeed`, `shop_extension`
- Task tracking: `scheduled_task`, `scheduled_task_run`
- Notifications: User's `notifications` column stores array of strings like `shop-123` for subscriptions

### Email Templates
- MJML templates in `/api/src/mail/sources/`
- Compiled to JS during build process
- Templates: alert, confirmation, password-reset

### Cron Jobs
- Located in `/api/src/cron/jobs/`
- Jobs: shopScrape, sessionCleanup, passwordResetCleanup
- Configured in `/api/src/cron/index.ts`

### Notification System
- User subscription-based notifications stored in `user.notifications` as JSON array
- Format: `["shop-123", "shop-456"]` for subscribed shops
- Shop notification logic in `/api/src/repository/shops.ts` filters by subscriptions
- Frontend watch/unwatch functionality in shop detail pages
- Settings page shows all subscribed shops with unsubscribe option
- tRPC endpoints: `subscribeToNotifications`, `unsubscribeFromNotifications`, `isSubscribedToNotifications`

## Environment Configuration

### API Environment Variables
```bash
APP_SECRET              # Encryption key (required)
APP_DATABASE_PATH       # SQLite database path
APP_FILES_DIR           # File storage directory
SMTP_HOST/PORT/USER/PASS # Email configuration
APP_SITESPEED_ENDPOINT  # Sitespeed.io service URL (default: http://localhost:3001)
FRONTEND_URL            # Frontend URL for emails
SENTRY_DSN              # Error tracking
```

### Sitespeed Service Environment Variables

### Sitespeed Results Access
- Results are accessible via `/sitespeed/result/<shop-id>/` endpoint
- API serves sitespeed files and directories from the configured data folder
- Frontend proxy passes `/sitespeed` requests to API server
- Detailed reports, screenshots, and videos are available through the web interface

### Frontend Environment
```bash
SHOPMON_API_URL         # API URL (defaults to production)
```

## Deployment

- Docker multi-stage build in root `Dockerfile`
- GitHub Actions workflow in `.github/workflows/`
- Staging deployment to `shea.shyim.de`
- Uses 1Password for secrets management
- Production uses Docker Compose (`compose.deploy.yml`)

## Security Patterns

- Client secrets encrypted using Node crypto module
- JWT tokens with configurable expiry
- Foreign key constraints in SQLite
- Session cleanup removes expired tokens
- Password reset tokens expire after 24 hours
