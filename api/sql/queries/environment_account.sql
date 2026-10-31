-- Account monitoring: baseline, capability state and the append-only event
-- history. See migrations/000030_environment_account_monitoring.up.sql for why
-- the three account kinds share one table.

-- name: UpsertEnvironmentAccount :exec
-- The WHERE ... IS DISTINCT FROM guard keeps the hourly scrape from writing a
-- new row version (and WAL) when nothing changed.
INSERT INTO environment_account (
  environment_id, account_type, account_id, label, active, admin,
  role_ids, roles_hash, access_key_count, privilege_count, privileges_hash,
  managed_by_app, app_name, password_recovery_pending,
  last_usage_at, password_changed_at, remote_created_at, remote_updated_at, admin_path
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19)
ON CONFLICT (environment_id, account_type, account_id) DO UPDATE SET
  label = EXCLUDED.label,
  active = EXCLUDED.active,
  admin = EXCLUDED.admin,
  role_ids = EXCLUDED.role_ids,
  roles_hash = EXCLUDED.roles_hash,
  access_key_count = EXCLUDED.access_key_count,
  privilege_count = EXCLUDED.privilege_count,
  privileges_hash = EXCLUDED.privileges_hash,
  managed_by_app = EXCLUDED.managed_by_app,
  app_name = EXCLUDED.app_name,
  password_recovery_pending = EXCLUDED.password_recovery_pending,
  last_usage_at = EXCLUDED.last_usage_at,
  password_changed_at = EXCLUDED.password_changed_at,
  remote_created_at = EXCLUDED.remote_created_at,
  remote_updated_at = EXCLUDED.remote_updated_at,
  admin_path = EXCLUDED.admin_path,
  synced_at = NOW()
WHERE (environment_account.label, environment_account.active, environment_account.admin,
       environment_account.role_ids, environment_account.roles_hash,
       environment_account.access_key_count, environment_account.privilege_count,
       environment_account.privileges_hash, environment_account.managed_by_app,
       environment_account.app_name, environment_account.password_recovery_pending,
       environment_account.last_usage_at, environment_account.password_changed_at,
       environment_account.remote_created_at, environment_account.remote_updated_at,
       environment_account.admin_path)
  IS DISTINCT FROM
      (EXCLUDED.label, EXCLUDED.active, EXCLUDED.admin,
       EXCLUDED.role_ids, EXCLUDED.roles_hash,
       EXCLUDED.access_key_count, EXCLUDED.privilege_count,
       EXCLUDED.privileges_hash, EXCLUDED.managed_by_app,
       EXCLUDED.app_name, EXCLUDED.password_recovery_pending,
       EXCLUDED.last_usage_at, EXCLUDED.password_changed_at,
       EXCLUDED.remote_created_at, EXCLUDED.remote_updated_at,
       EXCLUDED.admin_path);

-- name: DeleteEnvironmentAccountsNotIn :exec
-- Keys are "type:id" so one text[] covers all three account kinds. Callers must
-- never pass an empty list for a failed or empty fetch: that would wipe the
-- baseline and make every account look new on the next successful run.
DELETE FROM environment_account
WHERE environment_id = $1
  AND (account_type || ':' || account_id) != ALL(@keys::text[]);

-- name: ListEnvironmentAccounts :many
SELECT * FROM environment_account
WHERE environment_id = $1
ORDER BY account_type, account_id;

-- name: GetEnvironmentAccountState :one
SELECT * FROM environment_account_state WHERE environment_id = $1;

-- name: UpsertEnvironmentAccountState :exec
-- baseline_at is only ever set once: COALESCE keeps the original seeding time,
-- which is what tells later runs that a diff is meaningful.
INSERT INTO environment_account_state (
  environment_id, supported, baseline_at, account_count, generated_at, last_synced_at, last_error
) VALUES ($1, true, NOW(), $2, $3, NOW(), NULL)
ON CONFLICT (environment_id) DO UPDATE SET
  supported = true,
  baseline_at = COALESCE(environment_account_state.baseline_at, NOW()),
  account_count = EXCLUDED.account_count,
  generated_at = EXCLUDED.generated_at,
  last_synced_at = NOW(),
  last_error = NULL;

-- name: UpsertEnvironmentAccountStateStatus :exec
-- Records capability or failure without touching the counts, so an unreachable
-- endpoint does not report the shop as having zero accounts.
INSERT INTO environment_account_state (environment_id, supported, last_synced_at, last_error)
VALUES ($1, $2, NOW(), $3)
ON CONFLICT (environment_id) DO UPDATE SET
  supported = EXCLUDED.supported,
  last_synced_at = NOW(),
  last_error = EXCLUDED.last_error;

-- name: ResetEnvironmentAccountBaseline :exec
-- Clearing baseline_at makes the next scrape seed again instead of reporting
-- everything that changed while monitoring was switched off as a flood of new
-- findings.
UPDATE environment_account_state SET baseline_at = NULL WHERE environment_id = $1;

-- name: InsertEnvironmentAccountEvent :exec
-- DO NOTHING makes recording idempotent: two overlapping scrapes of one
-- environment diff the same baseline and reach for the same event, and the
-- reviewer must not have to acknowledge the same finding twice. The index
-- predicate has to be repeated for Postgres to infer the partial index.
--
-- Informational rows fall outside that index, so they never conflict and are
-- appended as plain history. A duplicate history line from two overlapping
-- scrapes is cosmetic; an informational row blocking a later finding would not
-- be.
INSERT INTO environment_account_event (
  environment_id, account_type, account_id, label, change, level, params, admin_path
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, account_type, account_id, change)
  WHERE acknowledged_at IS NULL AND level <> 'info'
  DO NOTHING;

-- name: ListOpenEnvironmentAccountEvents :many
-- Findings that still need review. info rows are history only and never become
-- a check.
SELECT * FROM environment_account_event
WHERE environment_id = $1 AND acknowledged_at IS NULL AND level <> 'info'
ORDER BY detected_at, id;

-- name: ListUnnotifiedEnvironmentAccountEvents :many
-- Acknowledged events are excluded: a failed alert leaves its events pending
-- for the next scrape, and in between a reviewer may well have confirmed one.
-- Alerting about a finding somebody already reviewed is pure noise.
SELECT * FROM environment_account_event
WHERE environment_id = $1
  AND notified_at IS NULL
  AND acknowledged_at IS NULL
  AND level <> 'info'
ORDER BY detected_at, id;

-- name: MarkEnvironmentAccountEventsNotified :exec
UPDATE environment_account_event
SET notified_at = NOW()
WHERE environment_id = $1 AND id = ANY(@ids::bigint[]);

-- name: ListEnvironmentAccountEvents :many
-- Full history for the UI timeline, acknowledged and informational rows
-- included. id breaks ties because rows written in one transaction share
-- detected_at.
SELECT e.*, u.name AS acknowledged_by_name
FROM environment_account_event e
LEFT JOIN "user" u ON u.id = e.acknowledged_by
WHERE e.environment_id = $1
ORDER BY e.detected_at DESC, e.id DESC
LIMIT $2;

-- name: GetEnvironmentAccountEvent :one
SELECT * FROM environment_account_event WHERE id = $1 AND environment_id = $2;

-- name: AcknowledgeEnvironmentAccountEvent :execrows
-- Returns zero rows when the event is already acknowledged, which the handler
-- turns into a 409 rather than silently succeeding twice.
UPDATE environment_account_event
SET acknowledged_at = NOW(), acknowledged_by = $3, acknowledge_note = $4
WHERE id = $1 AND environment_id = $2 AND acknowledged_at IS NULL;

-- name: AcknowledgeAllEnvironmentAccountEvents :exec
-- Used when monitoring is switched off: open findings would otherwise stay red
-- with no UI left to clear them from.
UPDATE environment_account_event
SET acknowledged_at = NOW(), acknowledged_by = $2, acknowledge_note = $3
WHERE environment_id = $1 AND acknowledged_at IS NULL;

-- name: CountOpenEnvironmentAccountFindings :one
SELECT COUNT(*)::int FROM environment_account_event
WHERE environment_id = $1 AND acknowledged_at IS NULL AND level <> 'info';
