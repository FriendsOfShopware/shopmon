# End-to-end tests (Playwright)

Browser-driven tests covering the main click-paths of the app: login/logout,
sidebar navigation, the dashboard, the environment-detail tabs, and public pages.

## Prerequisites

The full dev stack must be running (frontend on :3000, API on :5789) with
fixtures seeded:

```sh
mise run up            # infra (db, redis, mailpit, …)
mise run load-fixtures # reset db + migrate + seed Acme Corp / Acme Shop
mise run dev           # api + worker + frontend
```

Fixtures seed deterministic accounts (all password `password`):
`owner@fos.gg` (admin), `admin@fos.gg`, `member@fos.gg`, `regular@fos.gg`.

> The dev API raises the auth rate limit via `AUTH_RATE_LIMIT_MAX` (set in
> `.mise.toml`). Without it, a full run trips the default 20-req/min limiter and
> the session fetch gets 429'd. Production keeps the default of 20.

## Running

```sh
npm run test:e2e          # headless
npm run test:e2e:ui       # interactive UI mode
npm run test:e2e:report   # open the last HTML report
```

If the dev server isn't already running, Playwright starts one (see
`webServer` in `playwright.config.ts`).

## Layout

| File                   | Purpose                                                                                                                             |
| ---------------------- | ----------------------------------------------------------------------------------------------------------------------------------- |
| `auth.setup.ts`        | Logs in once as the owner and saves the session to `.auth/user.json`. The `setup` project runs this before the authenticated specs. |
| `support/test.ts`      | Extended `test` that pre-dismisses the "What's New" modal so it never overlays the page under test.                                 |
| `support/constants.ts` | Seeded fixture data (users, org, shop, environments). Keep in sync with `api/fixtures.go`.                                          |
| `*.spec.ts`            | The test suites.                                                                                                                    |

## CI

A nightly GitHub Actions workflow (`.github/workflows/e2e.yml`, also runnable
on demand via _workflow_dispatch_) runs this suite against a fresh stack: it
boots Postgres + Redis, migrates, seeds fixtures, starts the API on
`127.0.0.1:5789` (`mise run e2e:server`), and lets Playwright start the
frontend. The HTML report is uploaded as a build artifact.

## Conventions

- Prefer role/label/text locators (`getByRole`, `getByLabel`) — they double as
  accessibility coverage.
- Reach for `data-testid` only when no good accessible name exists (e.g. the
  dashboard summary numbers: `dashboard-stat-shops` / `-healthy` / `-warnings` /
  `-errors`).
- Specs that log in/out themselves opt out of the shared session with
  `test.use({ storageState: { cookies: [], origins: [] } })` so they don't
  invalidate the token other specs reuse.
