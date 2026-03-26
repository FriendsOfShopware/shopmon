-- name: ListShopsByOrganization :many
SELECT s.id, s.name, s.url, s.favicon, s.status, s.shopware_version,
       s.last_scraped_at, s.last_scraped_error, s.created_at, s.project_id, s.organization_id,
       o.name AS organization_name,
       p.name AS project_name, p.description AS project_description
FROM shop s
JOIN organization o ON o.id = s.organization_id
LEFT JOIN project p ON p.id = s.project_id
WHERE s.organization_id = $1
ORDER BY s.name;

-- name: GetShopByID :one
SELECT s.*, o.name AS organization_name, p.name AS project_name, p.description AS project_description
FROM shop s
JOIN organization o ON o.id = s.organization_id
LEFT JOIN project p ON p.id = s.project_id
WHERE s.id = $1;

-- name: CreateShop :one
INSERT INTO shop (organization_id, project_id, name, url, client_id, client_secret, shopware_version, shop_token, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
RETURNING id;

-- name: UpdateShop :exec
UPDATE shop SET name = $1, url = $2, client_id = $3, client_secret = $4, ignores = $5, project_id = $6 WHERE id = $7;

-- name: UpdateShopName :exec
UPDATE shop SET name = $1 WHERE id = $2;

-- name: DeleteShop :exec
DELETE FROM shop WHERE id = $1;

-- name: GetShopExtensions :many
SELECT id, shop_id, name, label, active, version, latest_version, installed, rating_average, store_link, changelog, installed_at
FROM shop_extension WHERE shop_id = $1 ORDER BY name;

-- name: GetShopScheduledTasks :many
SELECT id, shop_id, task_id, name, status, interval, overdue, last_execution_time, next_execution_time
FROM shop_scheduled_task WHERE shop_id = $1 ORDER BY name;

-- name: GetShopQueues :many
SELECT id, shop_id, name, size FROM shop_queue WHERE shop_id = $1 ORDER BY name;

-- name: GetShopCache :one
SELECT id, shop_id, environment, http_cache, cache_adapter FROM shop_cache WHERE shop_id = $1;

-- name: GetShopChecks :many
SELECT id, shop_id, check_id, level, message, source, link FROM shop_check WHERE shop_id = $1;

-- name: GetShopSitespeeds :many
SELECT id, shop_id, deployment_id, created_at, ttfb, fully_loaded, largest_contentful_paint, first_contentful_paint, cumulative_layout_shift, transfer_size
FROM shop_sitespeed WHERE shop_id = $1 ORDER BY created_at DESC LIMIT 100;

-- name: GetShopChangelogs :many
SELECT id, shop_id, extensions, old_shopware_version, new_shopware_version, date
FROM shop_changelog WHERE shop_id = $1 ORDER BY date DESC LIMIT 10;

-- name: CountShopDeployments :one
SELECT COUNT(*)::int FROM deployment WHERE shop_id = $1;

-- name: UpdateShopSitespeedSettings :exec
UPDATE shop SET sitespeed_enabled = $1, sitespeed_urls = $2 WHERE id = $3;

-- name: GetAllShops :many
SELECT s.id, s.name, s.url, s.client_id, s.client_secret, s.shopware_version,
       s.organization_id, s.ignores, s.last_scraped_at, s.last_scraped_error,
       s.connection_issue_count, s.shop_token, s.sitespeed_enabled, s.sitespeed_urls
FROM shop s;

-- name: GetShopForScrape :one
SELECT s.id, s.name, s.url, s.client_id, s.client_secret, s.shopware_version,
       s.organization_id, s.ignores, s.last_scraped_at, s.last_scraped_error,
       s.connection_issue_count, s.shop_token, s.sitespeed_enabled, s.sitespeed_urls
FROM shop s WHERE s.id = $1;

-- name: UpdateShopAfterScrape :exec
UPDATE shop SET status = $1, shopware_version = $2, last_scraped_at = NOW(), last_scraped_error = $3,
       favicon = $4, shop_image = $5, last_changelog = $6, connection_issue_count = 0
WHERE id = $7;

-- name: UpdateShopImage :exec
UPDATE shop SET shop_image = $1 WHERE id = $2;

-- name: UpdateShopScrapeError :exec
UPDATE shop SET last_scraped_error = $1, last_scraped_at = NOW(), connection_issue_count = connection_issue_count + 1 WHERE id = $2;

-- name: UpsertShopExtension :exec
INSERT INTO shop_extension (shop_id, name, label, active, version, latest_version, installed, rating_average, store_link, changelog, installed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (shop_id, name) DO UPDATE SET label = $3, active = $4, version = $5, latest_version = $6, installed = $7, rating_average = $8, store_link = $9, changelog = $10, installed_at = $11;

-- name: DeleteShopExtensionsNotIn :exec
DELETE FROM shop_extension WHERE shop_id = $1 AND name != ALL($2::text[]);

-- name: UpsertShopQueue :exec
INSERT INTO shop_queue (shop_id, name, size) VALUES ($1, $2, $3)
ON CONFLICT (shop_id, name) DO UPDATE SET size = $3;

-- name: DeleteShopQueuesNotIn :exec
DELETE FROM shop_queue WHERE shop_id = $1 AND name != ALL($2::text[]);

-- name: UpsertShopScheduledTask :exec
INSERT INTO shop_scheduled_task (shop_id, task_id, name, status, interval, overdue, last_execution_time, next_execution_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (shop_id, task_id) DO UPDATE SET name = $3, status = $4, interval = $5, overdue = $6, last_execution_time = $7, next_execution_time = $8;

-- name: DeleteShopScheduledTasksNotIn :exec
DELETE FROM shop_scheduled_task WHERE shop_id = $1 AND task_id != ALL($2::text[]);

-- name: UpsertShopCache :exec
INSERT INTO shop_cache (shop_id, environment, http_cache, cache_adapter)
VALUES ($1, $2, $3, $4)
ON CONFLICT (shop_id) DO UPDATE SET environment = $2, http_cache = $3, cache_adapter = $4;

-- name: DeleteShopChecks :exec
DELETE FROM shop_check WHERE shop_id = $1;

-- name: InsertShopCheck :exec
INSERT INTO shop_check (shop_id, check_id, level, message, source, link)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (shop_id, check_id) DO UPDATE SET level = $3, message = $4, source = $5, link = $6;

-- name: InsertShopChangelog :exec
INSERT INTO shop_changelog (shop_id, extensions, old_shopware_version, new_shopware_version, date)
VALUES ($1, $2, $3, $4, NOW());

-- name: UpdateShopActiveDeployment :exec
UPDATE shop SET active_deployment_id = $1 WHERE id = $2;

-- name: GetShopCredentials :one
SELECT id, url, client_id, client_secret, shop_token FROM shop WHERE id = $1;

-- name: GetShopsWithSitespeedEnabled :many
SELECT s.id, s.url, s.sitespeed_urls, s.active_deployment_id, s.shop_token FROM shop s WHERE s.sitespeed_enabled = true;

-- name: InsertShopSitespeed :exec
INSERT INTO shop_sitespeed (shop_id, deployment_id, created_at, ttfb, fully_loaded, largest_contentful_paint, first_contentful_paint, cumulative_layout_shift, transfer_size)
VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8);

-- name: GetShopNotificationSubscribers :many
SELECT u.id, u.name, u.email
FROM "user" u
JOIN member m ON m.user_id = u.id
WHERE m.organization_id = @organization_id
  AND u.notifications @> to_jsonb(ARRAY['shop-' || @shop_id::text])::jsonb;
