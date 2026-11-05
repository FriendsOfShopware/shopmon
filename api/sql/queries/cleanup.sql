-- name: CleanupExpiredSessions :exec
DELETE FROM session WHERE expires_at < NOW();

-- name: CleanupOldNotifications :exec
DELETE FROM user_notification WHERE read = true AND created_at < NOW() - interval '90 days';

-- name: CleanupOldSitespeedData :exec
DELETE FROM environment_sitespeed WHERE created_at < NOW() - interval '90 days';

-- name: CleanupExpiredBans :exec
UPDATE "user" SET banned = false, ban_reason = NULL, ban_expires = NULL
WHERE banned = true AND ban_expires IS NOT NULL AND ban_expires < NOW();

-- name: CleanupOldEnvironmentAccountEvents :exec
-- Trims the account history. Open findings are never deleted: an unreviewed
-- security change must not disappear because it got old.
DELETE FROM environment_account_event
WHERE detected_at < NOW() - interval '180 days'
  AND (level = 'info' OR acknowledged_at IS NOT NULL);
