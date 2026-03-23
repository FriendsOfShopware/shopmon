-- name: AdminGetStats :one
SELECT
  (SELECT COUNT(*) FROM "user")::int AS total_users,
  (SELECT COUNT(*) FROM organization)::int AS total_organizations,
  (SELECT COUNT(*) FROM shop)::int AS total_shops,
  (SELECT COUNT(*) FROM shop WHERE status = 'green')::int AS shops_green,
  (SELECT COUNT(*) FROM shop WHERE status = 'yellow')::int AS shops_yellow,
  (SELECT COUNT(*) FROM shop WHERE status = 'red')::int AS shops_red;

-- name: AdminListOrganizations :many
SELECT o.id, o.name, o.slug, o.logo, o.created_at,
       (SELECT COUNT(*) FROM shop WHERE organization_id = o.id)::int AS shop_count,
       (SELECT COUNT(*) FROM member WHERE organization_id = o.id)::int AS member_count
FROM organization o
ORDER BY o.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountOrganizations :one
SELECT COUNT(*)::int FROM organization;

-- name: AdminListShops :many
SELECT s.id, s.name, s.url, s.status, s.shopware_version, s.last_scraped_at, s.created_at,
       s.organization_id, o.name AS organization_name
FROM shop s
JOIN organization o ON o.id = s.organization_id
ORDER BY s.name
LIMIT $1 OFFSET $2;

-- name: AdminCountShops :one
SELECT COUNT(*)::int FROM shop;

-- name: AdminGetGrowthUsers :many
SELECT to_char(created_at, 'YYYY-MM') AS month, COUNT(*)::int AS count
FROM "user" GROUP BY month ORDER BY month;

-- name: AdminGetGrowthShops :many
SELECT to_char(created_at, 'YYYY-MM') AS month, COUNT(*)::int AS count
FROM shop GROUP BY month ORDER BY month;

-- name: AdminGetRecentUsers :many
SELECT id, name, email, created_at FROM "user" ORDER BY created_at DESC LIMIT 10;

-- name: AdminGetRecentShops :many
SELECT s.id, s.name, s.url, o.name AS organization_name, s.created_at
FROM shop s
JOIN organization o ON o.id = s.organization_id
ORDER BY s.created_at DESC LIMIT 10;

-- name: AdminGetShopwareVersions :many
SELECT shopware_version AS version, COUNT(*)::int AS count
FROM shop GROUP BY shopware_version ORDER BY count DESC;
