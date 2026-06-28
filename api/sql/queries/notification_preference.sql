-- name: ListNotificationPreferences :many
SELECT id, user_id, scope_type, scope_id, event_type, channel, enabled, config
FROM notification_preference
WHERE user_id = $1;

-- name: UpsertNotificationPreference :exec
INSERT INTO notification_preference (user_id, scope_type, scope_id, event_type, channel, enabled, config)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (user_id, scope_type, scope_id, event_type, channel)
DO UPDATE SET enabled = $6, config = $7;

-- name: DeleteNotificationPreference :exec
DELETE FROM notification_preference
WHERE user_id = $1 AND scope_type = $2 AND scope_id = $3 AND event_type = $4 AND channel = $5;

-- name: SubscribeEnvironment :exec
INSERT INTO notification_preference (user_id, scope_type, scope_id, event_type, channel, enabled)
VALUES ($1, 'environment', $2, '', '', true)
ON CONFLICT (user_id, scope_type, scope_id, event_type, channel)
DO UPDATE SET enabled = true;

-- name: UnsubscribeEnvironment :exec
DELETE FROM notification_preference
WHERE user_id = $1 AND scope_type = 'environment' AND scope_id = $2;

-- name: IsEnvironmentSubscribed :one
SELECT EXISTS (
  SELECT 1 FROM notification_preference
  WHERE user_id = $1 AND scope_type = 'environment' AND scope_id = $2 AND channel = '' AND enabled = true
) AS subscribed;

-- name: ListSubscribedEnvironmentIDs :many
SELECT scope_id FROM notification_preference
WHERE user_id = $1 AND scope_type = 'environment' AND channel = '' AND enabled = true;

-- name: DeleteUserOrgNotificationPreferences :exec
-- Removes a user's environment-scoped preferences for every environment in an
-- organization (used when the user is removed from / leaves the org).
DELETE FROM notification_preference
WHERE user_id = $1
  AND scope_type = 'environment'
  AND scope_id IN (SELECT id::text FROM environment WHERE organization_id = $2);

-- name: DeleteEnvironmentNotificationPreferences :exec
-- Removes all users' preferences scoped to a single environment (used when the
-- environment is deleted). scope_id is text, so there is no FK to cascade on.
DELETE FROM notification_preference
WHERE scope_type = 'environment' AND scope_id = $1;
