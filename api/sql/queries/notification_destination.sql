-- name: ListNotificationDestinations :many
SELECT * FROM notification_destination
WHERE organization_id = $1 AND COALESCE(shop_id, 0)::integer = sqlc.arg(shop_id)::integer
ORDER BY created_at, id;

-- name: GetEventNotificationDestinations :many
SELECT d.* FROM notification_destination d
WHERE d.organization_id = $1 AND (d.shop_id IS NULL OR d.shop_id IN (
  SELECT e.shop_id FROM environment e WHERE e.id = sqlc.arg(environment_id)::integer AND e.organization_id = $1
)) ORDER BY d.shop_id NULLS FIRST, d.created_at, d.id;

-- name: CreateNotificationDestination :execrows
INSERT INTO notification_destination (id, organization_id, shop_id, name, dsn)
VALUES ($1, $2, NULLIF(sqlc.arg(shop_id)::integer, 0), $3, $4);

-- name: UpdateNotificationDestination :execrows
UPDATE notification_destination SET name = sqlc.arg(name),
dsn = COALESCE(NULLIF(sqlc.arg(dsn)::text, ''), dsn), updated_at = NOW()
WHERE id = sqlc.arg(id) AND organization_id = sqlc.arg(organization_id)
AND COALESCE(shop_id, 0)::integer = sqlc.arg(shop_id)::integer;

-- name: DeleteNotificationDestination :execrows
DELETE FROM notification_destination WHERE id = $1 AND organization_id = $2
AND COALESCE(shop_id, 0)::integer = sqlc.arg(shop_id)::integer;

-- name: ShopBelongsToOrganization :one
SELECT EXISTS (SELECT 1 FROM shop WHERE id = $1 AND organization_id = $2);

-- name: GetNotificationDestinationInOrganization :one
SELECT * FROM notification_destination WHERE id = $1 AND organization_id = $2;

-- name: ListDestinationNotificationPreferences :many
SELECT * FROM destination_notification_preference WHERE destination_id = $1;

-- name: UpsertDestinationNotificationPreference :exec
INSERT INTO destination_notification_preference (destination_id, scope_type, scope_id, event_type, channel, enabled)
VALUES ($1, $2, $3, $4, 'destination', $5)
ON CONFLICT (destination_id, scope_type, scope_id, event_type, channel)
DO UPDATE SET enabled = EXCLUDED.enabled;

-- name: DeleteDestinationNotificationPreference :exec
DELETE FROM destination_notification_preference
WHERE destination_id = $1 AND scope_type = $2 AND scope_id = $3 AND event_type = $4;

-- name: EnvironmentMatchesNotificationDestination :one
SELECT EXISTS (
 SELECT 1 FROM environment e JOIN notification_destination d ON d.organization_id = e.organization_id
 WHERE d.id = $1 AND e.id = $2 AND (d.shop_id IS NULL OR d.shop_id = e.shop_id)
);
