-- name: CleanupExpiredSessions :exec
DELETE FROM session WHERE expires_at < NOW();

-- name: CleanupOldNotifications :exec
DELETE FROM user_notification WHERE read = true AND created_at < NOW() - interval '90 days';

-- name: CleanupOldSitespeedData :exec
DELETE FROM environment_sitespeed WHERE created_at < NOW() - interval '90 days';

-- name: CleanupOldChangelogData :exec
DELETE FROM environment_changelog WHERE date < NOW() - interval '180 days';

-- name: CleanupExpiredBans :exec
UPDATE "user" SET banned = false, ban_reason = NULL, ban_expires = NULL
WHERE banned = true AND ban_expires IS NOT NULL AND ban_expires < NOW();
