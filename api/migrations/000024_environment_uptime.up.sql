-- External storefront uptime monitoring.
--
-- environment_uptime holds monitor configuration and the live probe state in a
-- single 1:1 row per environment. Probes run outside the queue (ticker loop in
-- the worker); this row is the coordination point via next_check_at.
CREATE TABLE "environment_uptime" (
  "environment_id" integer PRIMARY KEY REFERENCES "environment"("id") ON DELETE cascade,
  "enabled" boolean NOT NULL DEFAULT false,
  -- NULL means the environment's own URL is probed.
  "url" text,
  "interval_seconds" integer NOT NULL DEFAULT 60,
  -- 0 accepts any 2xx/3xx response; otherwise the exact status is required.
  "expected_status" integer NOT NULL DEFAULT 0,
  -- Optional substring that must appear in the response body.
  "content_match" text,
  -- Consecutive failing probes before an up->down transition fires.
  "failure_threshold" integer NOT NULL DEFAULT 3,
  -- Consecutive successful probes before a down->up recovery fires.
  "recovery_threshold" integer NOT NULL DEFAULT 2,
  -- unknown | up | down | paused
  "status" text NOT NULL DEFAULT 'unknown',
  "consecutive_failures" integer NOT NULL DEFAULT 0,
  "consecutive_successes" integer NOT NULL DEFAULT 0,
  "next_check_at" timestamp,
  "last_checked_at" timestamp,
  "last_status_code" integer,
  "last_latency_ms" integer,
  "last_error" text
);

-- Supports the due-monitor claim scan.
CREATE INDEX "idx_environment_uptime_next_check" ON "environment_uptime" ("next_check_at")
  WHERE "enabled" AND "next_check_at" IS NOT NULL;

-- Downtime incidents: one row per up->down transition, resolved_at set on
-- recovery. An open incident has resolved_at IS NULL.
CREATE TABLE "environment_uptime_event" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "started_at" timestamp NOT NULL,
  "resolved_at" timestamp,
  "status_code" integer,
  "latency_ms" integer,
  "error" text
);

CREATE INDEX "idx_environment_uptime_event_env" ON "environment_uptime_event" ("environment_id", "started_at" DESC);

-- Hourly probe aggregates. probes_expected lets readers distinguish "no data"
-- from "everything fine" when the worker was not probing for the full hour.
CREATE TABLE "environment_uptime_rollup_hourly" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "hour" timestamp NOT NULL,
  "probes_expected" integer NOT NULL DEFAULT 0,
  "probes_run" integer NOT NULL DEFAULT 0,
  "probes_ok" integer NOT NULL DEFAULT 0,
  "paused_seconds" integer NOT NULL DEFAULT 0,
  "latency_avg_ms" integer,
  "latency_p95_ms" integer,
  PRIMARY KEY ("environment_id", "hour")
);

-- Daily availability aggregates drive the status history (90-day strip) and
-- availability percentages. Availability is computed at read time as
-- up_seconds / (up_seconds + down_seconds); paused and nodata stay out of the
-- denominator.
CREATE TABLE "environment_uptime_rollup_daily" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "day" date NOT NULL,
  "up_seconds" integer NOT NULL DEFAULT 0,
  "down_seconds" integer NOT NULL DEFAULT 0,
  "paused_seconds" integer NOT NULL DEFAULT 0,
  "nodata_seconds" integer NOT NULL DEFAULT 0,
  "latency_avg_ms" integer,
  "latency_p95_ms" integer,
  PRIMARY KEY ("environment_id", "day")
);
