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
ORDER BY o.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountOrganizations :one
SELECT COUNT(*)::int FROM organization;

-- name: AdminListEnvironments :many
SELECT e.id, e.name, e.url, e.status, e.shopware_version, e.last_scraped_at, e.created_at,
       e.organization_id, o.name AS organization_name
FROM environment e
JOIN organization o ON o.id = e.organization_id
ORDER BY e.name
LIMIT $1 OFFSET $2;

-- name: AdminCountEnvironments :one
SELECT COUNT(*)::int FROM environment;

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
