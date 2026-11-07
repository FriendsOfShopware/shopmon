-- name: ListNotificationDestinations :many
SELECT * FROM notification_destination
WHERE organization_id = $1 AND COALESCE(shop_id, 0)::integer = sqlc.arg(shop_id)::integer
ORDER BY created_at, id;

-- name: GetEventNotificationDestinations :many
SELECT d.* FROM notification_destination d
WHERE d.organization_id = $1 AND (d.shop_id IS NULL OR d.shop_id IN (
  SELECT e.shop_id FROM environment e WHERE e.id = sqlc.arg(environment_id)::integer AND e.organization_id = $1
)) ORDER BY d.shop_id NULLS FIRST, d.created_at, d.id;

-- name: SaveNotificationDestination :execrows
INSERT INTO notification_destination (id, organization_id, shop_id, name, dsn)
VALUES ($1, $2, NULLIF(sqlc.arg(shop_id)::integer, 0), $3, $4)
ON CONFLICT (id) DO UPDATE SET name = EXCLUDED.name, dsn = EXCLUDED.dsn, updated_at = NOW()
WHERE notification_destination.organization_id = EXCLUDED.organization_id
AND notification_destination.shop_id IS NOT DISTINCT FROM EXCLUDED.shop_id;

-- name: DeleteNotificationDestination :execrows
DELETE FROM notification_destination WHERE id = $1 AND organization_id = $2
AND COALESCE(shop_id, 0)::integer = sqlc.arg(shop_id)::integer;

-- name: ShopBelongsToOrganization :one
SELECT EXISTS (SELECT 1 FROM shop WHERE id = $1 AND organization_id = $2);
