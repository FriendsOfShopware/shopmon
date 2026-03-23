-- name: ListNotifications :many
SELECT id, user_id, key, level, title, message, link, read, created_at
FROM user_notification WHERE user_id = $1 ORDER BY created_at DESC;

-- name: DeleteNotification :exec
DELETE FROM user_notification WHERE id = $1 AND user_id = $2;

-- name: DeleteAllNotifications :exec
DELETE FROM user_notification WHERE user_id = $1;

-- name: MarkAllNotificationsRead :exec
UPDATE user_notification SET read = true WHERE user_id = $1;

-- name: UpsertNotification :exec
INSERT INTO user_notification (user_id, key, level, title, message, link, read, created_at)
VALUES ($1, $2, $3, $4, $5, $6, false, NOW())
ON CONFLICT (user_id, key) DO UPDATE SET level = $3, title = $4, message = $5, link = $6, read = false, created_at = NOW();
