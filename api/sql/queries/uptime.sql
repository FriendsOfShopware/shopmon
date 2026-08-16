-- name: UpsertUptimeSettings :exec
INSERT INTO environment_uptime (
  environment_id, enabled, url, interval_seconds, expected_status, content_match,
  failure_threshold, recovery_threshold, status, next_check_at
)
VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, 'unknown',
  CASE WHEN $2 THEN NOW() ELSE NULL END
)
ON CONFLICT (environment_id) DO UPDATE SET
  enabled = EXCLUDED.enabled,
  url = EXCLUDED.url,
  interval_seconds = EXCLUDED.interval_seconds,
  expected_status = EXCLUDED.expected_status,
  content_match = EXCLUDED.content_match,
  failure_threshold = EXCLUDED.failure_threshold,
  recovery_threshold = EXCLUDED.recovery_threshold,
  -- Enabling (re-)arms the monitor: fresh state, first probe immediately.
  status = 'unknown',
  consecutive_failures = 0,
  consecutive_successes = 0,
  next_check_at = CASE WHEN EXCLUDED.enabled THEN NOW() ELSE NULL END;

-- name: GetUptimeSettings :one
SELECT
  u.environment_id, u.enabled, u.url, u.interval_seconds, u.expected_status,
  u.content_match, u.failure_threshold, u.recovery_threshold, u.status,
  u.last_checked_at, u.last_status_code, u.last_latency_ms, u.last_error,
  e.url AS environment_url, e.name AS environment_name, e.organization_id
FROM environment_uptime u
JOIN environment e ON e.id = u.environment_id
WHERE u.environment_id = $1;

-- name: ClaimDueUptimeMonitors :many
-- Atomically claims due monitors by pushing next_check_at forward before they
-- are probed, so concurrent ticks (or a slow probe overlapping the next tick)
-- can never probe the same monitor twice.
UPDATE environment_uptime u SET
  next_check_at = NOW() + make_interval(secs => u.interval_seconds::int)
FROM environment e
WHERE e.id = u.environment_id
  AND u.enabled = true
  AND u.status <> 'paused'
  AND u.next_check_at IS NOT NULL
  AND u.next_check_at <= @now
  AND (u.environment_id % @shard_total) = @shard_index
RETURNING
  u.environment_id, u.url, u.interval_seconds, u.expected_status, u.content_match,
  u.failure_threshold, u.recovery_threshold, u.status,
  u.consecutive_failures, u.consecutive_successes,
  e.url AS environment_url, e.name AS environment_name, e.organization_id;

-- name: UpdateUptimeProbeState :exec
UPDATE environment_uptime SET
  status = $2,
  consecutive_failures = $3,
  consecutive_successes = $4,
  last_checked_at = NOW(),
  last_status_code = $5,
  last_latency_ms = $6,
  last_error = $7
WHERE environment_id = $1;

-- name: PauseUptimeMonitor :exec
UPDATE environment_uptime SET
  status = 'paused',
  next_check_at = NULL
WHERE environment_id = $1;

-- name: ResumeUptimeMonitor :exec
UPDATE environment_uptime SET
  status = 'unknown',
  consecutive_failures = 0,
  consecutive_successes = 0,
  next_check_at = NOW()
WHERE environment_id = $1 AND enabled = true;

-- name: InsertUptimeEvent :one
INSERT INTO environment_uptime_event (environment_id, started_at, status_code, latency_ms, error)
VALUES ($1, $2, $3, $4, $5)
RETURNING id;

-- name: ResolveUptimeEvents :exec
UPDATE environment_uptime_event SET resolved_at = $2
WHERE environment_id = $1 AND resolved_at IS NULL;

-- name: GetOpenUptimeEvent :one
SELECT * FROM environment_uptime_event
WHERE environment_id = $1 AND resolved_at IS NULL
ORDER BY started_at DESC LIMIT 1;

-- name: ListUptimeEvents :many
SELECT * FROM environment_uptime_event
WHERE environment_id = $1 AND started_at >= $2
ORDER BY started_at DESC;

-- name: ListUptimeEventsOverlapping :many
SELECT * FROM environment_uptime_event
WHERE environment_id = $1
  AND started_at < @range_to
  AND (resolved_at IS NULL OR resolved_at > @range_from);

-- name: UpsertUptimeHourlyRollup :exec
-- Additive on conflict: the same hour can be flushed more than once when a
-- probe started in hour H completes after H was already flushed; late samples
-- must merge into the existing row, not replace it.
INSERT INTO environment_uptime_rollup_hourly (
  environment_id, hour, probes_expected, probes_run, probes_ok, paused_seconds,
  latency_avg_ms, latency_p95_ms
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, hour) DO UPDATE SET
  probes_expected = GREATEST(environment_uptime_rollup_hourly.probes_expected, EXCLUDED.probes_expected),
  probes_run = environment_uptime_rollup_hourly.probes_run + EXCLUDED.probes_run,
  probes_ok = environment_uptime_rollup_hourly.probes_ok + EXCLUDED.probes_ok,
  paused_seconds = environment_uptime_rollup_hourly.paused_seconds + EXCLUDED.paused_seconds,
  latency_avg_ms = CASE
    WHEN environment_uptime_rollup_hourly.latency_avg_ms IS NULL THEN EXCLUDED.latency_avg_ms
    WHEN EXCLUDED.latency_avg_ms IS NULL THEN environment_uptime_rollup_hourly.latency_avg_ms
    ELSE (environment_uptime_rollup_hourly.latency_avg_ms * environment_uptime_rollup_hourly.probes_run
        + EXCLUDED.latency_avg_ms * EXCLUDED.probes_run)
        / (environment_uptime_rollup_hourly.probes_run + EXCLUDED.probes_run)
  END,
  latency_p95_ms = GREATEST(environment_uptime_rollup_hourly.latency_p95_ms, EXCLUDED.latency_p95_ms);

-- name: ListUptimeHourlyRollups :many
SELECT * FROM environment_uptime_rollup_hourly
WHERE environment_id = $1 AND hour >= $2 AND hour < $3
ORDER BY hour;

-- name: ListUptimeDailyRollups :many
SELECT * FROM environment_uptime_rollup_daily
WHERE environment_id = $1 AND day >= $2 AND day <= $3
ORDER BY day;

-- name: ListUptimeDailyAggregateTargets :many
SELECT DISTINCT environment_id FROM (
  SELECT environment_id FROM environment_uptime_event
  WHERE started_at < @range_to AND (resolved_at IS NULL OR resolved_at > @range_from)
  UNION
  SELECT environment_id FROM environment_uptime_rollup_hourly
  WHERE hour >= @range_from AND hour < @range_to
) t;

-- name: UpsertUptimeDailyRollup :exec
INSERT INTO environment_uptime_rollup_daily (
  environment_id, day, up_seconds, down_seconds, paused_seconds, nodata_seconds,
  latency_avg_ms, latency_p95_ms
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, day) DO UPDATE SET
  up_seconds = EXCLUDED.up_seconds,
  down_seconds = EXCLUDED.down_seconds,
  paused_seconds = EXCLUDED.paused_seconds,
  nodata_seconds = EXCLUDED.nodata_seconds,
  latency_avg_ms = EXCLUDED.latency_avg_ms,
  latency_p95_ms = EXCLUDED.latency_p95_ms;

-- name: DeleteUptimeHourlyRollupsOlderThan :exec
DELETE FROM environment_uptime_rollup_hourly WHERE hour < $1;

-- name: DeleteUptimeEventsOlderThan :exec
DELETE FROM environment_uptime_event WHERE started_at < $1 AND resolved_at IS NOT NULL;
