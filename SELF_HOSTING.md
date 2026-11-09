# Self-Hosting Shopmon

Shopmon is distributed as a single Docker image that bundles the API server, background worker, and frontend.

**Image:** `ghcr.io/friendsofshopware/shopmon:staging`

## Requirements

- **PostgreSQL** 18+
- **Redis** 8+
- A reverse proxy (Traefik, Caddy, nginx) for TLS termination
- *Optional:* an AMQP broker (LavinMQ or RabbitMQ) if you want background jobs off PostgreSQL — see [Background Job Queue](#background-job-queue)

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

All variables are read once at startup. A value that cannot be parsed (a
non-numeric count, an unparseable duration, a boolean that is not
`true`/`false`) or that is out of range makes the process exit with an error
instead of falling back to the default, so a typo cannot silently run the
instance with settings you did not choose.

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
| `MAIL_DSN` | | Full SMTP connection string, e.g. `smtp://user:pass@host:587` or `smtps://user:pass@host:465` (implicit TLS). Supported query options: `verify_peer`, `require_tls`, `auto_tls`. When set, it takes precedence over the `SMTP_*` variables below. |
| `SMTP_HOST` | `localhost` | SMTP server hostname (used to build the DSN when `MAIL_DSN` is unset) |
| `SMTP_PORT` | `1025` | SMTP server port |
| `SMTP_SECURE` | `false` | Use implicit TLS (`true`/`false`); selects the `smtps` scheme |
| `SMTP_USER` | | SMTP username |
| `SMTP_PASS` | | SMTP password |
| `MAIL_FROM` | `noreply@shopmon.io` | Sender address |
| `SMTP_REPLY_TO` | | Reply-to address |

### Notification destinations

Organization owners and admins can add multiple named notification destinations on the organization page or in a shop's settings. Delivery uses [go-notifier](https://github.com/shyim/go-notifier). Organization destinations can receive organization alerts; shop destinations receive that shop's alerts as well. Identical DSNs receive only one copy of an event, even when configured at both scopes. No environment subscribers are required. Shared destinations have their own admin-managed event preferences; a user's personal watch, email or in-app settings never control team delivery.

Use **Event settings** beside a destination to enable or disable each registered event type. Defaults are enabled, matching personal notification preferences. In an environment's **Notification settings**, the **Shared destination settings** section offers **Inherit / On / Off** overrides for each destination. Inherit uses that destination's default; an environment override affects only that environment. This section is available even when you are not watching the environment. Members can view shared settings, but only organization owners and admins can change them. Test messages intentionally bypass event preferences to verify the connection.

Choose a service in **Add destination**, then paste its webhook URL or enter the requested token and channel/server details. Shopmon builds and encodes the connection DSN automatically. Use **Send test** after saving to verify delivery.

For API clients, the stored connection formats are:

| Provider | DSN |
|----------|-----|
| Slack incoming webhook | `slack://default?webhook_url=URL_ENCODED_WEBHOOK_URL` |
| Slack bot | `slack://xoxb-TOKEN@default?channel=CHANNEL_ID` |
| Telegram | `telegram://BOT_TOKEN@default?channel=CHAT_ID` |
| Discord | `discord://WEBHOOK_TOKEN@default?webhook_id=WEBHOOK_ID` |
| Gotify | `gotify://APP_TOKEN@SERVER_HOST` |
| Microsoft Teams | `microsoftteams://WEBHOOK_ID@default?token=TOKEN` |

When configuring destinations through the API, URL-encode special characters in credentials and query values. For Slack webhooks, encode the complete `https://hooks.slack.com/services/…` URL as the `webhook_url` value. The Teams format above matches the pinned library's factory: for `https://outlook.office.com/webhook/ID/IncomingWebhook/TOKEN`, use `ID` and `TOKEN`.

DSNs are encrypted with `APP_SECRET` and are never returned to the browser. Editing a name keeps the existing credentials. Choose **Replace connection** to enter new credentials (API clients can leave the DSN empty to keep it). Changing `APP_SECRET` requires reconfiguring destinations. Saving alone does not send a message.

The server (test messages) and worker (alerts) need outbound access to the configured services. Destinations must resolve to public addresses; private, loopback and link-local addresses and HTTP redirects are blocked. This also applies to self-hosted Gotify. HTTP tracing is disabled for these requests because provider URLs can contain credentials; delivery outcome metrics remain available.

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

### Background Job Queue

Background jobs run on PostgreSQL by default — no extra infrastructure needed. If you already operate a message broker, or want the queue off your database, switch the transport to AMQP (RabbitMQ or LavinMQ).

| Variable | Default | Description |
|---|---|---|
| `QUEUE_TRANSPORT` | `postgres` | Job backend: `postgres` (jobs stored in `queue_messages` + `queue_messages_scrape`) or `amqp` |
| `QUEUE_AMQP_DSN` | `amqp://guest:guest@localhost:5672/` | Broker connection string |
| `QUEUE_AMQP_EXCHANGE` | `shopmon` | Exchange to publish jobs to (declared on startup) |
| `QUEUE_AMQP_QUEUE` | `shopmon` | Catalog-sync queue (and routing key). Shop-facing jobs use `<QUEUE_AMQP_QUEUE>-scrape` (default `shopmon-scrape`) |
| `QUEUE_AMQP_PREFETCH` | `10` | Unacknowledged deliveries per consumer per queue. Keep it at or above the worker concurrency (10) |
| `QUEUE_AMQP_DELAYED_EXCHANGE` | `true` | Declare the exchange as `x-delayed-message` so delayed jobs are held by the broker. Accepts `true`/`false`, `1`/`0`, `TRUE`/`FALSE`; an unparseable value keeps delayed delivery on |

Set the same values on the `api` and `worker` services — the API publishes jobs and the worker consumes them.

Shop-refresh jobs (`EnvironmentScrape`, sitespeed, changelog, cleanup, advisory sync) use a dedicated queue so a `StoreExtensionSync` backlog cannot starve shop updates. On PostgreSQL the worker creates `queue_messages_scrape` automatically. On AMQP it declares `<QUEUE_AMQP_QUEUE>-scrape`. After deploy, confirm the worker logs `enqueued environment scrapes` at `:00` and `scraping environment` shortly after. Leftover `EnvironmentScrape` rows in `queue_messages` (or messages still on the catalog AMQP queue) can be left to drain or deleted; the hourly cron republishes to the scrape queue.

Shopmon delays some jobs (post-deployment scrapes, sitespeed reruns), which needs broker-side delayed delivery: **LavinMQ supports it natively**, RabbitMQ needs the [rabbitmq_delayed_message_exchange](https://github.com/rabbitmq/rabbitmq-delayed-message-exchange) plugin. Only set `QUEUE_AMQP_DELAYED_EXCHANGE=false` if your broker has neither — delayed jobs are then delivered immediately, so a post-deployment scrape runs before the shop has settled.

Switching transports does not migrate jobs that are already queued. Drain the worker on the old transport before flipping the variable, or those jobs are left behind.

### S3 Storage (Deployments)

Required only if you use the deployment tracking feature.

| Variable | Default | Description |
|---|---|---|
| `APP_S3_ENDPOINT` | | S3-compatible endpoint URL |
| `APP_S3_ACCESS_KEY_ID` | | Access key |
| `APP_S3_SECRET_ACCESS_KEY` | | Secret key |
| `APP_S3_BUCKET` | `shopmon` | Bucket name |
| `APP_S3_REGION` | `auto` | Region |

### Deployments

| Variable | Default | Description |
|---|---|---|
| `DEPLOYMENT_SCRAPE_DELAY` | `5m` | How long to wait after a deployment is recorded before re-scraping the environment. Accepts a Go duration (e.g. `30s`, `5m`, `2m30s`). Gives post-deploy tasks (theme compile, indexing, cache warming) time to settle. Set to `0` to scrape immediately. |

### Observability (OpenTelemetry)

| Variable | Default | Description |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | | OTLP endpoint for traces and logs |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | | Override for trace-specific endpoint |
| `OTEL_EXPORTER_OTLP_LOGS_ENDPOINT` | | Override for log-specific endpoint |
| `OTEL_SERVICE_NAME` | `shopmon` | Base APM service name. The API reports as `${OTEL_SERVICE_NAME}-api` and the worker as `${OTEL_SERVICE_NAME}-worker` (defaults: `shopmon-api` / `shopmon-worker`) |
| `OTEL_DEPLOYMENT_ENVIRONMENT` | `$DD_ENV` | Deployment environment reported as a resource attribute |
| `OTEL_SERVICE_VERSION` | `$DD_VERSION`, else the build revision | Service version reported as a resource attribute |
| `OTEL_TRACES_SAMPLER_RATIO` | `1` | Head sampling ratio, clamped to `[0, 1]` |


## Public Composer advisory feed

`GET /api/advisories/composer` returns every published advisory for
`store.shopware.com/` packages, grouped by package name under `advisories` in
Composer's security-advisories API format. No authentication or pagination is
required. Hidden drafts and non-store package rows are excluded. Package-specific
Packagist IDs are preserved; advisories without one use their Shopmon/GHSA ID.

The same URL accepts `POST` requests from Composer and returns the complete feed
(package filters are ignored). To use it with `composer audit`, a Composer
repository's `packages.json` must advertise the endpoint, for example:

```json
{
  "packages": {},
  "security-advisories": {
    "metadata": false,
    "api-url": "https://your-shopmon.example/api/advisories/composer"
  }
}
```

Configure that repository in your project's `composer.json`. The feed itself is
an advisory API response, not a `packages.json` repository index.

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
