-- name: GetOrganizationSlackWebhook :one
SELECT webhook_url FROM organization_slack_webhook WHERE organization_id = $1;

-- name: UpsertOrganizationSlackWebhook :exec
INSERT INTO organization_slack_webhook (organization_id, webhook_url, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (organization_id) DO UPDATE SET webhook_url = EXCLUDED.webhook_url, updated_at = NOW();

-- name: DeleteOrganizationSlackWebhook :exec
DELETE FROM organization_slack_webhook WHERE organization_id = $1;

-- name: GetShopSlackWebhook :one
SELECT w.webhook_url FROM shop_slack_webhook w
JOIN shop s ON s.id = w.shop_id
WHERE w.shop_id = $1 AND s.organization_id = $2;

-- name: GetEnvironmentShopSlackWebhook :one
SELECT w.webhook_url FROM shop_slack_webhook w
JOIN environment e ON e.shop_id = w.shop_id
WHERE e.id = $1;

-- name: ShopBelongsToOrganization :one
SELECT EXISTS (SELECT 1 FROM shop WHERE id = $1 AND organization_id = $2);

-- name: UpsertShopSlackWebhook :exec
INSERT INTO shop_slack_webhook (shop_id, webhook_url, created_at, updated_at)
VALUES ($1, $2, NOW(), NOW())
ON CONFLICT (shop_id) DO UPDATE SET webhook_url = EXCLUDED.webhook_url, updated_at = NOW();

-- name: DeleteShopSlackWebhook :exec
DELETE FROM shop_slack_webhook WHERE shop_id = $1;
