-- name: GetUserByID :one
SELECT id, name, email, email_verified, image, created_at, updated_at, role, banned, ban_reason, ban_expires, notifications FROM "user" WHERE id = $1;

-- name: GetUserNotifications :one
SELECT notifications FROM "user" WHERE id = $1;

-- name: UpdateUserNotifications :exec
UPDATE "user" SET notifications = $1 WHERE id = $2;

-- name: GetSessionByToken :one
SELECT s.id, s.expires_at, s.token, s.user_id, s.active_organization_id, s.impersonated_by,
       u.id as user_id, u.name as user_name, u.email as user_email, u.role as user_role, u.notifications as user_notifications, u.image as user_image, u.banned as user_banned
FROM session s
JOIN "user" u ON u.id = s.user_id
WHERE s.token = $1 AND s.expires_at > NOW();

-- name: GetUserOrganizations :many
SELECT o.id, o.name, o.slug, o.logo, o.created_at,
       (SELECT COUNT(*) FROM environment WHERE organization_id = o.id)::int AS environment_count,
       (SELECT COUNT(*) FROM member WHERE organization_id = o.id)::int AS member_count
FROM organization o
JOIN member m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.name;

-- name: GetUserEnvironments :many
SELECT e.id, e.name, e.url, e.favicon, e.status, e.shopware_version, e.last_scraped_at, e.last_scraped_error,
       e.organization_id, o.name AS organization_name, s.id AS shop_id, s.name AS shop_name
FROM environment e
JOIN member m ON m.organization_id = e.organization_id
JOIN organization o ON o.id = e.organization_id
LEFT JOIN shop s ON s.id = e.shop_id
WHERE m.user_id = $1
ORDER BY e.name;

-- name: GetUserShops :many
SELECT s.id, s.name, s.description, s.git_url, s.organization_id, o.name AS organization_name
FROM shop s
JOIN member m ON m.organization_id = s.organization_id
JOIN organization o ON o.id = s.organization_id
WHERE m.user_id = $1
ORDER BY s.name;

-- name: GetUserChangelogs :many
SELECT ec.id, ec.environment_id, ec.extensions, ec.old_shopware_version, ec.new_shopware_version, ec.date,
       e.name AS environment_name, o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_changelog ec
JOIN environment e ON e.id = ec.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1
ORDER BY ec.date DESC
LIMIT 10;

-- name: GetUserExtensions :many
SELECT ee.name, ee.label, ee.version, ee.latest_version, ee.active, ee.installed,
       ee.store_link, ee.rating_average, ee.installed_at, ee.changelog,
       ee.environment_id, e.name AS environment_name, e.url AS environment_url, o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_extension ee
JOIN environment e ON e.id = ee.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1
ORDER BY ee.name, e.name;

-- name: GetUserEnvironmentsByOrg :many
SELECT e.id, e.name, e.url, e.favicon, e.status, e.shopware_version, e.last_scraped_at, e.last_scraped_error,
       e.organization_id, o.name AS organization_name, s.id AS shop_id, s.name AS shop_name
FROM environment e
JOIN member m ON m.organization_id = e.organization_id
JOIN organization o ON o.id = e.organization_id
LEFT JOIN shop s ON s.id = e.shop_id
WHERE m.user_id = $1 AND e.organization_id = $2
ORDER BY e.name;

-- name: GetUserShopsByOrg :many
SELECT s.id, s.name, s.description, s.git_url, s.default_environment_id, s.organization_id, o.name AS organization_name
FROM shop s
JOIN member m ON m.organization_id = s.organization_id
JOIN organization o ON o.id = s.organization_id
WHERE m.user_id = $1 AND s.organization_id = $2
ORDER BY s.name;

-- name: GetUserChangelogsByOrg :many
SELECT ec.id, ec.environment_id, ec.extensions, ec.old_shopware_version, ec.new_shopware_version, ec.date,
       e.name AS environment_name, o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_changelog ec
JOIN environment e ON e.id = ec.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1 AND e.organization_id = $2
ORDER BY ec.date DESC
LIMIT 10;

-- name: GetUserExtensionsByOrg :many
SELECT ee.name, ee.label, ee.version, ee.latest_version, ee.active, ee.installed,
       ee.store_link, ee.rating_average, ee.installed_at, ee.changelog,
       ee.environment_id, e.name AS environment_name, e.url AS environment_url, o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_extension ee
JOIN environment e ON e.id = ee.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1 AND e.organization_id = $2
ORDER BY ee.name, e.name;

-- name: IsOrganizationMember :one
SELECT COUNT(*) > 0 AS is_member FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: GetEnvironmentOrganizationID :one
SELECT organization_id FROM environment WHERE id = $1;
