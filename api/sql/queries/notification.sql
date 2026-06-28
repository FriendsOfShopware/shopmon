-- name: ListNotifications :many
SELECT id, user_id, key, level, title, message, title_key, message_key, params, link, read, created_at
FROM user_notification WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteNotification :exec
DELETE FROM user_notification WHERE id = $1 AND user_id = $2;

-- name: DeleteAllNotifications :exec
DELETE FROM user_notification WHERE user_id = $1;

-- name: MarkAllNotificationsRead :exec
UPDATE user_notification SET read = true WHERE user_id = $1;

-- name: UpsertNotification :exec
INSERT INTO user_notification (user_id, key, level, title, message, title_key, message_key, params, link, read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, false, NOW())
ON CONFLICT (user_id, key) DO UPDATE SET
  level = $3, title = $4, message = $5, title_key = $6, message_key = $7, params = $8, link = $9, read = false, created_at = NOW();

-- name: DeleteUserOrgNotifications :exec
-- Removes a user's inbox notifications tied to environments of an organization.
DELETE FROM user_notification
WHERE user_id = $1
  AND (link->'params'->>'environmentId') IN (
    SELECT id::text FROM environment WHERE organization_id = $2
  );

-- name: DeleteEnvironmentNotifications :exec
-- Removes all users' inbox notifications tied to a single environment.
DELETE FROM user_notification
WHERE (link->'params'->>'environmentId') = @environment_id::text;
