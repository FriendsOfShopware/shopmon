-- Advisory suppressions: a shop owner's recorded decision that an advisory is
-- accepted or mitigated outside Shopmon.
--
-- "Active" is always (revoked_at IS NULL AND (expires_at IS NULL OR expires_at
-- > NOW())). Expiry is enforced at read time rather than by a cleanup job, so
-- there is no window in which an expired suppression still silences an alert.
-- The same predicate must be used everywhere or the scope tab counts will
-- disagree with the list.

-- name: CreateAdvisorySuppression :one
INSERT INTO advisory_suppression (
  organization_id, shop_id, environment_id, advisory_id,
  reason, expires_at, created_by, created_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW())
RETURNING *;

-- name: RevokeAdvisorySuppression :one
UPDATE advisory_suppression
SET revoked_at = NOW(), revoked_by = $2
WHERE id = $1 AND revoked_at IS NULL
  AND organization_id = ANY(sqlc.arg('organization_ids')::text[])
RETURNING *;

-- The partial unique indexes only exclude revoked rows (an "active" predicate
-- cannot use NOW() in an index), so an expired-but-unrevoked row still blocks a
-- new suppression for the same scope. Retiring it on conflict keeps the audit
-- trail and lets the scope be suppressed again.
-- name: RevokeExpiredAdvisorySuppressions :execrows
UPDATE advisory_suppression
SET revoked_at = NOW(), revoked_by = $3
WHERE shop_id = $1
  AND advisory_id = $2
  AND (
    (sqlc.narg('environment_id')::int IS NULL AND environment_id IS NULL)
    OR environment_id = sqlc.narg('environment_id')::int
  )
  AND revoked_at IS NULL
  AND expires_at IS NOT NULL AND expires_at <= NOW();


-- Loaded before a revoke so the same role policy that governed creation can be
-- applied: the scope (shop-wide vs environment) decides which roles may act.
-- name: GetAdvisorySuppressionInOrgs :one
SELECT * FROM advisory_suppression
WHERE id = $1 AND organization_id = ANY(sqlc.arg('organization_ids')::text[]);

-- Active suppressions covering one environment: shop-wide rows plus rows
-- narrowed to this environment.
-- name: ListActiveSuppressionsForEnvironment :many
SELECT s.*
FROM advisory_suppression s
JOIN environment e ON e.id = sqlc.arg('environment_id')::int
WHERE s.shop_id = e.shop_id
  AND (s.environment_id IS NULL OR s.environment_id = e.id)
  AND s.revoked_at IS NULL
  AND (s.expires_at IS NULL OR s.expires_at > NOW())
ORDER BY s.advisory_id;

-- name: ListSuppressionsForOrganizations :many
SELECT s.*,
       a.title AS advisory_title,
       a.cve AS advisory_cve,
       a.severity AS advisory_severity,
       a.severity_override AS advisory_severity_override,
       sh.name AS shop_name,
       e.name AS environment_name,
       u.name AS created_by_name
FROM advisory_suppression s
JOIN composer_advisory a ON a.advisory_id = s.advisory_id
JOIN shop sh ON sh.id = s.shop_id
LEFT JOIN environment e ON e.id = s.environment_id
LEFT JOIN "user" u ON u.id = s.created_by
WHERE s.organization_id = ANY(sqlc.arg('organization_ids')::text[])
  AND (
    sqlc.arg('include_inactive')::boolean
    OR (s.revoked_at IS NULL AND (s.expires_at IS NULL OR s.expires_at > NOW()))
  )
ORDER BY s.created_at DESC;

-- Advisories actively suppressed for one environment, with the identifiers the
-- checker uses to build a check id. Consumed by the scrape so a suppressed
-- advisory neither escalates status nor triggers an alert.
-- name: ListSuppressedAdvisoriesForEnvironment :many
SELECT a.advisory_id, a.cve, a.ghsa_id
FROM advisory_suppression s
JOIN environment e ON e.id = sqlc.arg('environment_id')::int
JOIN composer_advisory a ON a.advisory_id = s.advisory_id
WHERE s.shop_id = e.shop_id
  AND (s.environment_id IS NULL OR s.environment_id = e.id)
  AND s.revoked_at IS NULL
  AND (s.expires_at IS NULL OR s.expires_at > NOW())
ORDER BY a.advisory_id;
