-- name: GetUserByID :one
SELECT id, name, email, email_verified, image, created_at, updated_at, role, banned, ban_reason, ban_expires, notifications FROM "user" WHERE id = $1;

-- name: GetUserNotifications :one
SELECT notifications FROM "user" WHERE id = $1;

-- name: UpdateUserNotifications :exec
UPDATE "user" SET notifications = $1 WHERE id = $2;

-- name: GetSessionByToken :one
SELECT s.id, s.expires_at, s.token, s.user_id, s.active_organization_id, s.impersonated_by,
       u.id as user_id, u.name as user_name, u.email as user_email, u.role as user_role, u.notifications as user_notifications, u.image as user_image, u.banned as user_banned, u.ban_expires as user_ban_expires
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

-- name: GetUserStoreExtensionsByOrg :many
SELECT ese.extension_name AS name, ese.label, ese.version, ese.latest_version, ese.active, ese.installed,
       se.store_link, se.rating_average, ese.installed_at,
       ese.environment_id, e.name AS environment_name, e.url AS environment_url,
       o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_store_extension ese
JOIN store_extension se ON se.name = ese.extension_name
JOIN environment e ON e.id = ese.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1 AND e.organization_id = $2
ORDER BY ese.extension_name, e.name;

-- name: GetUserUnknownExtensionsByOrg :many
SELECT ee.name, ee.label, ee.version, ee.latest_version, ee.active, ee.installed,
       NULL::text AS store_link, NULL::integer AS rating_average, ee.installed_at,
       ee.environment_id, e.name AS environment_name, e.url AS environment_url,
       o.name AS environment_organization_name, e.organization_id AS environment_organization_id
FROM environment_extension ee
JOIN environment e ON e.id = ee.environment_id
JOIN organization o ON o.id = e.organization_id
JOIN member m ON m.organization_id = e.organization_id
WHERE m.user_id = $1 AND e.organization_id = $2
ORDER BY ee.name, e.name;

-- name: GetUserStoreExtensionChangelogsByOrg :many
SELECT DISTINCT sev.extension_name, sev.version, sev.changelog_en, sev.changelog_de, sev.released_at
FROM store_extension_version sev
WHERE sev.extension_name IN (
  SELECT ese.extension_name
  FROM environment_store_extension ese
  JOIN environment e ON e.id = ese.environment_id
  JOIN member m ON m.organization_id = e.organization_id
  WHERE m.user_id = $1 AND e.organization_id = $2
)
ORDER BY sev.extension_name, sev.version;

-- name: IsOrganizationMember :one
SELECT COUNT(*) > 0 AS is_member FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: GetEnvironmentOrganizationID :one
SELECT organization_id FROM environment WHERE id = $1;
