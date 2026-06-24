-- name: AdminGetStats :one
SELECT
  (SELECT COUNT(*) FROM "user")::int AS total_users,
  (SELECT COUNT(*) FROM organization)::int AS total_organizations,
  (SELECT COUNT(*) FROM environment)::int AS total_environments,
  (SELECT COUNT(*) FROM environment WHERE status = 'green')::int AS environments_green,
  (SELECT COUNT(*) FROM environment WHERE status = 'yellow')::int AS environments_yellow,
  (SELECT COUNT(*) FROM environment WHERE status = 'red')::int AS environments_red;

-- name: AdminListOrganizations :many
SELECT o.id, o.name, o.slug, o.logo, o.created_at,
       (SELECT COUNT(*) FROM environment WHERE organization_id = o.id)::int AS environment_count,
       (SELECT COUNT(*) FROM member WHERE organization_id = o.id)::int AS member_count
FROM organization o
WHERE (sqlc.narg('search')::text IS NULL OR o.name ILIKE '%' || sqlc.narg('search') || '%' OR o.slug ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_dir')::text = 'asc' THEN o.name END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_dir')::text = 'desc' THEN o.name END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'memberCount' AND sqlc.arg('sort_dir')::text = 'asc' THEN (SELECT COUNT(*) FROM member WHERE organization_id = o.id) END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'memberCount' AND sqlc.arg('sort_dir')::text = 'desc' THEN (SELECT COUNT(*) FROM member WHERE organization_id = o.id) END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'createdAt' AND sqlc.arg('sort_dir')::text = 'asc' THEN o.created_at END ASC,
  o.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountOrganizations :one
SELECT COUNT(*)::int FROM organization o
WHERE (sqlc.narg('search')::text IS NULL OR o.name ILIKE '%' || sqlc.narg('search') || '%' OR o.slug ILIKE '%' || sqlc.narg('search') || '%');

-- name: AdminListEnvironments :many
SELECT e.id, e.name, e.url, e.status, e.shopware_version, e.last_scraped_at, e.created_at,
       e.organization_id, o.name AS organization_name,
       e.shop_id, s.name AS shop_name
FROM environment e
JOIN organization o ON o.id = e.organization_id
JOIN shop s ON s.id = e.shop_id
WHERE (sqlc.narg('search')::text IS NULL OR e.name ILIKE '%' || sqlc.narg('search') || '%' OR e.url ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_dir')::text = 'asc' THEN e.name END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'name' AND sqlc.arg('sort_dir')::text = 'desc' THEN e.name END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'status' AND sqlc.arg('sort_dir')::text = 'asc' THEN e.status END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'status' AND sqlc.arg('sort_dir')::text = 'desc' THEN e.status END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'organizationName' AND sqlc.arg('sort_dir')::text = 'asc' THEN o.name END ASC,
  CASE WHEN sqlc.arg('sort_by')::text = 'organizationName' AND sqlc.arg('sort_dir')::text = 'desc' THEN o.name END DESC,
  CASE WHEN sqlc.arg('sort_by')::text = 'createdAt' AND sqlc.arg('sort_dir')::text = 'asc' THEN e.created_at END ASC,
  e.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountEnvironments :one
SELECT COUNT(*)::int FROM environment e
WHERE (sqlc.narg('search')::text IS NULL OR e.name ILIKE '%' || sqlc.narg('search') || '%' OR e.url ILIKE '%' || sqlc.narg('search') || '%');

-- name: AdminGetGrowthUsers :many
SELECT to_char(created_at, 'YYYY-MM') AS month, COUNT(*)::int AS count
FROM "user" GROUP BY month ORDER BY month;

-- name: AdminGetGrowthEnvironments :many
SELECT to_char(created_at, 'YYYY-MM') AS month, COUNT(*)::int AS count
FROM environment GROUP BY month ORDER BY month;

-- name: AdminGetRecentUsers :many
SELECT id, name, email, created_at FROM "user" ORDER BY created_at DESC LIMIT 10;

-- name: AdminGetRecentEnvironments :many
SELECT e.id, e.name, e.url, o.name AS organization_name, e.created_at
FROM environment e
JOIN organization o ON o.id = e.organization_id
ORDER BY e.created_at DESC LIMIT 10;

-- name: AdminGetShopwareVersions :many
SELECT shopware_version AS version, COUNT(*)::int AS count
FROM environment GROUP BY shopware_version ORDER BY count DESC;

-- Phase 3: Organization detail --------------------------------------------

-- name: AdminGetOrganizationDetail :one
SELECT o.id, o.name, o.slug, o.logo, o.created_at,
       (SELECT COUNT(*) FROM environment WHERE organization_id = o.id)::int AS environment_count,
       (SELECT COUNT(*) FROM member WHERE organization_id = o.id)::int AS member_count
FROM organization o
WHERE o.id = $1;

-- name: AdminGetOrganizationMembers :many
SELECT m.user_id, u.name, u.email, u.image, m.role, m.created_at
FROM member m
JOIN "user" u ON u.id = m.user_id
WHERE m.organization_id = $1
ORDER BY m.created_at;

-- name: AdminGetOrganizationEnvironments :many
SELECT e.id, e.name, e.url, e.status, e.shopware_version, e.last_scraped_at
FROM environment e
WHERE e.organization_id = $1
ORDER BY e.name;

-- name: AdminGetOrganizationInvitations :many
SELECT i.id, i.email, i.role, i.status, i.expires_at, i.created_at,
       u.name AS inviter_name, u.email AS inviter_email
FROM invitation i
JOIN "user" u ON u.id = i.inviter_id
WHERE i.organization_id = $1
ORDER BY i.created_at DESC;

-- name: AdminGetOrganizationSSO :many
SELECT id, issuer, provider_id, domain
FROM sso_provider
WHERE organization_id = $1
ORDER BY domain;

-- name: AdminGetOrganizationShops :many
SELECT id, name, description, default_environment_id, created_at
FROM shop
WHERE organization_id = $1
ORDER BY name;

-- Phase 4: Environment detail ---------------------------------------------

-- name: AdminGetEnvironmentDetail :one
SELECT e.id, e.name, e.url, e.status, e.shopware_version, e.created_at,
       e.last_scraped_at, e.last_scraped_error, e.connection_issue_count,
       e.organization_id, o.name AS organization_name,
       e.shop_id, s.name AS shop_name
FROM environment e
JOIN organization o ON o.id = e.organization_id
JOIN shop s ON s.id = e.shop_id
WHERE e.id = $1;

-- name: AdminGetEnvironmentChecks :many
SELECT id, check_id, level, message, source, link
FROM environment_check
WHERE environment_id = $1
ORDER BY level, check_id;

-- name: AdminGetEnvironmentExtensions :many
SELECT id, name, label, active, version, latest_version, installed, store_link
FROM environment_extension
WHERE environment_id = $1
ORDER BY name;

-- name: AdminGetEnvironmentTasks :many
SELECT id, task_id, name, status, "interval", overdue, last_execution_time, next_execution_time
FROM environment_scheduled_task
WHERE environment_id = $1
ORDER BY overdue DESC, name;

-- name: AdminGetEnvironmentLastDeployment :one
SELECT id, name, command, return_code, execution_time, reference, created_at
FROM deployment
WHERE environment_id = $1
ORDER BY created_at DESC
LIMIT 1;
