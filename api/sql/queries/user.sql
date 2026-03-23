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
       (SELECT COUNT(*) FROM shop WHERE organization_id = o.id)::int AS shop_count,
       (SELECT COUNT(*) FROM member WHERE organization_id = o.id)::int AS member_count
FROM organization o
JOIN member m ON m.organization_id = o.id
WHERE m.user_id = $1
ORDER BY o.name;

-- name: GetUserShops :many
SELECT s.id, s.name, s.url, s.favicon, s.status, s.shopware_version, s.last_scraped_at, s.last_scraped_error,
       s.organization_id, o.name AS organization_name, p.id AS project_id, p.name AS project_name
FROM shop s
JOIN member m ON m.organization_id = s.organization_id
JOIN organization o ON o.id = s.organization_id
LEFT JOIN project p ON p.id = s.project_id
WHERE m.user_id = $1
ORDER BY s.name;

-- name: GetUserProjects :many
SELECT p.id, p.name, p.description, p.git_url, p.organization_id, o.name AS organization_name
FROM project p
JOIN member m ON m.organization_id = p.organization_id
JOIN organization o ON o.id = p.organization_id
WHERE m.user_id = $1
ORDER BY p.name;

-- name: GetUserChangelogs :many
SELECT sc.id, sc.shop_id, sc.extensions, sc.old_shopware_version, sc.new_shopware_version, sc.date,
       s.name AS shop_name, o.name AS shop_organization_name, s.organization_id AS shop_organization_id
FROM shop_changelog sc
JOIN shop s ON s.id = sc.shop_id
JOIN organization o ON o.id = s.organization_id
JOIN member m ON m.organization_id = s.organization_id
WHERE m.user_id = $1
ORDER BY sc.date DESC
LIMIT 10;

-- name: GetUserExtensions :many
SELECT se.name, se.label, se.version, se.latest_version, se.active, se.installed,
       se.store_link, se.rating_average, se.installed_at, se.changelog,
       se.shop_id, s.name AS shop_name, s.url AS shop_url, o.name AS shop_organization_name, s.organization_id AS shop_organization_id
FROM shop_extension se
JOIN shop s ON s.id = se.shop_id
JOIN organization o ON o.id = s.organization_id
JOIN member m ON m.organization_id = s.organization_id
WHERE m.user_id = $1
ORDER BY se.name, s.name;

-- name: IsOrganizationMember :one
SELECT COUNT(*) > 0 AS is_member FROM member WHERE organization_id = $1 AND user_id = $2;

-- name: GetShopOrganizationID :one
SELECT organization_id FROM shop WHERE id = $1;
