# Self-Hosting Shopmon

Shopmon is distributed as a single Docker image that bundles the API server, background worker, and frontend.

**Image:** `ghcr.io/friendsofshopware/shopmon:staging`

## Requirements

- **PostgreSQL** 18+
- **Redis** 8+
- A reverse proxy (Traefik, Caddy, nginx) for TLS termination

## Quick Start

Create a `compose.yml`:

```yaml
x-env: &env
  DATABASE_URL: postgres://shopmon:shopmon@db:5432/shopmon?sslmode=disable
  REDIS_URL: redis://redis:6379
  FRONTEND_URL: https://shopmon.example.com
  APP_SECRET: change-me-to-a-32-byte-secret!! # must be exactly 16, 24, or 32 bytes
  MAIL_FROM: "Shopmon <noreply@example.com>"
  SMTP_HOST: smtp.example.com
  SMTP_PORT: "587"
  SMTP_SECURE: "true"
  SMTP_USER: your-smtp-user
  SMTP_PASS: your-smtp-password

services:
  migrate:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on:
      - db
    environment:
      <<: *env
    command: ["migrate", "up"]

  api:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on:
      migrate:
        condition: service_completed_successfully
    ports:
      - "8080:8080"
    environment:
      <<: *env

  worker:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on:
      - db
      - redis
    environment:
      <<: *env
    command: ["worker"]

  redis:
    image: redis:8-alpine
    volumes:
      - redis-data:/data

  db:
    image: postgres:18
    environment:
      POSTGRES_DB: shopmon
      POSTGRES_USER: shopmon
      POSTGRES_PASSWORD: shopmon
    volumes:
      - db-data:/var/lib/postgresql/data

volumes:
  db-data:
  redis-data:
```

```bash
docker compose up -d
```

Your instance is now available at `http://localhost:8080`. Put a reverse proxy in front for TLS.

## Services

The Docker image runs three different commands:

| Command | Description |
|---|---|
| `server` (default) | HTTP API server on port 8080, serves the frontend |
| `worker` | Background job processor (shop scraping, notifications, cleanup) |
| `migrate up` | Applies database migrations (run before starting the API) |

## Environment Variables

### Required

| Variable | Description |
|---|---|
| `DATABASE_URL` | PostgreSQL connection string |
| `REDIS_URL` | Redis connection string |
| `FRONTEND_URL` | Public URL of your instance (used for email links, OAuth callbacks, WebAuthn) |
| `APP_SECRET` | AES encryption key for secrets. Must be exactly 16, 24, or 32 bytes |

### Email (SMTP)

| Variable | Default | Description |
|---|---|---|
| `SMTP_HOST` | `localhost` | SMTP server hostname |
| `SMTP_PORT` | `1025` | SMTP server port |
| `SMTP_SECURE` | `false` | Use TLS (`true`/`false`) |
| `SMTP_USER` | | SMTP username |
| `SMTP_PASS` | | SMTP password |
| `MAIL_FROM` | `noreply@shopmon.io` | Sender address |
| `SMTP_REPLY_TO` | | Reply-to address |

### Feature Toggles

All features below are **disabled by default** unless configured. The frontend automatically hides UI for disabled features via the `GET /api/info/config` endpoint.

#### GitHub OAuth

| Variable | Description |
|---|---|
| `APP_OAUTH_GITHUB_CLIENT_ID` | GitHub OAuth App Client ID |
| `APP_OAUTH_GITHUB_CLIENT_SECRET` | GitHub OAuth App Client Secret |

Create an OAuth App at https://github.com/settings/developers with the callback URL: `{FRONTEND_URL}/auth/callback`

When not set, the GitHub login button is hidden from the login page.

#### Sitespeed Analytics

| Variable | Description |
|---|---|
| `APP_SITESPEED_ENDPOINT` | URL to your [sitespeed.io](https://www.sitespeed.io) service |
| `APP_SITESPEED_PREFIX` | Prefix for sitespeed result paths (e.g. `prod-`) |
| `APP_SITESPEED_API_KEY` | API key for the sitespeed service |

When not set, the sitespeed tab, settings, and documentation section are hidden.

#### Package Mirror

| Variable | Description |
|---|---|
| `PACKAGES_API_URL` | URL to the packages mirror API |
| `PACKAGES_API_TOKEN` | Authentication token for the packages API |

When not set, the packages token management UI and documentation section are hidden.

#### Registration

| Variable | Default | Description |
|---|---|---|
| `DISABLE_REGISTRATION` | `false` | Set to `true` to disable email/password registration |

When enabled, the registration page is blocked, the "Create account" link is hidden from the login page, and the API returns 403 on sign-up attempts. Useful when you want to restrict access to invited users only.

### S3 Storage (Deployments)

Required only if you use the deployment tracking feature.

| Variable | Default | Description |
|---|---|---|
| `APP_S3_ENDPOINT` | | S3-compatible endpoint URL |
| `APP_S3_ACCESS_KEY_ID` | | Access key |
| `APP_S3_SECRET_ACCESS_KEY` | | Secret key |
| `APP_S3_BUCKET` | `shopmon` | Bucket name |
| `APP_S3_REGION` | `auto` | Region |

### Observability (OpenTelemetry)

| Variable | Default | Description |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | | OTLP endpoint for traces and logs |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | | Override for trace-specific endpoint |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | | Override for log-specific endpoint |
| `OTEL_SERVICE_NAME` | `shopmon` | Service name in traces |


## Reverse Proxy Examples

### Caddy

```
shopmon.example.com {
    reverse_proxy api:8080
}
```

### nginx

```nginx
server {
    listen 443 ssl;
    server_name shopmon.example.com;

    location / {
        proxy_pass http://api:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

### Traefik (labels)

```yaml
labels:
  - traefik.enable=true
  - traefik.http.routers.shopmon.rule=Host(`shopmon.example.com`)
  - traefik.http.routers.shopmon.tls.certresolver=letsencrypt
```

## Minimal Self-Hosting Example

If you want a minimal instance with just email/password login and no optional features:

```yaml
x-env: &env
  DATABASE_URL: postgres://shopmon:shopmon@db:5432/shopmon
  REDIS_URL: redis://redis:6379
  FRONTEND_URL: https://shopmon.example.com
  APP_SECRET: change-me-to-a-32-byte-secret!!

services:
  migrate:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on: [db]
    environment:
      <<: *env
    command: ["migrate", "up"]

  api:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on:
      migrate:
        condition: service_completed_successfully
    ports: ["8080:8080"]
    environment:
      <<: *env

  worker:
    image: ghcr.io/friendsofshopware/shopmon:staging
    depends_on: [db, redis]
    environment:
      <<: *env
    command: ["worker"]

  redis:
    image: redis:8-alpine

  db:
    image: postgres:18
    environment:
      POSTGRES_DB: shopmon
      POSTGRES_USER: shopmon
      POSTGRES_PASSWORD: shopmon
    volumes:
      - db-data:/var/lib/postgresql/data

volumes:
  db-data:
```

This gives you shop monitoring, health checks, extensions, scheduled tasks, queue monitoring, changelog, deployments, and SSO -- with GitHub auth, sitespeed, and package mirror UI hidden.
