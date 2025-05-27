# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

Shopmon is a SaaS application for monitoring multiple Shopware instances. It tracks shop versions, extensions, security updates, scheduled tasks, and performance metrics.

## Architecture

### API (Cloudflare Workers)
- Located in `/api/`
- Built with TypeScript, tRPC, Hono framework
- Uses Drizzle ORM with Cloudflare D1 (SQLite)
- Cloudflare Durable Objects for real-time features
- Deployed via Wrangler to Cloudflare Workers

### Frontend (Vue 3 SPA)
- Located in `/frontend/`
- Vue 3 with TypeScript, Pinia stores, Vue Router
- tRPC client for type-safe API communication
- Tailwind CSS for styling
- Built with Vite

### Database
- Cloudflare D1 (SQLite in production)
- Migrations in `/api/drizzle/`
- Schema defined via Drizzle ORM

## Common Commands

### Setup & Development
```bash
# Install dependencies for both API and frontend
make setup

# Run database migrations
make migrate

# Start development servers (API on :8787, Frontend on :5173)
make dev

# Run frontend against production API
make dev-to-prod
```

### API-specific Commands
```bash
cd api
pnpm run dev:local        # Local development
pnpm run generate         # Generate TypeScript types
pnpm run db:generate      # Generate database migrations
pnpm run migrate          # Apply migrations
pnpm run api:deploy       # Deploy to Cloudflare
pnpm run biome:check      # Lint/format check
pnpm run biome:fix        # Auto-fix lint/format issues
```

### Frontend-specific Commands
```bash
cd frontend
pnpm run dev              # Development server
pnpm run build            # Production build
pnpm run preview          # Preview production build
```

## Key Patterns & Conventions

### tRPC Router Structure
- Main router: `/api/src/router.ts`
- Feature routers in `/api/src/trpc/*/index.ts`
- Context and middleware in `/api/src/trpc/`

### Frontend State Management
- Pinia stores in `/frontend/src/stores/`
- Each major feature has its own store
- tRPC client configured in `/frontend/src/helpers/trpc.ts`

### Authentication
- JWT-based authentication
- Passkey support
- Auth logic in `/api/src/trpc/auth/`


### Email Templates
- MJML templates in `/api/src/mail/sources/`
- Compiled during build process

## Environment Configuration

### API Environment Variables
- `VITE_*` variables are exposed to frontend
- Critical variables: DATABASE, JWT_SECRET, CRYPTO_KEY
- Email: SMTP_* or RESEND_API_TOKEN
- Optional: DISABLE_REGISTRATION, PAGESPEED_API_KEY

### Frontend Environment
- API URL configured via VITE_API_URL
- Defaults to production API if not set

## Deployment

- API deploys to Cloudflare Workers via `wrangler`
- Frontend can be deployed to any static hosting
- Database migrations must be run before deployment