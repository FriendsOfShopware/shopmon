-- name: ListEnvironmentsByOrganization :many
SELECT e.id, e.name, e.url, e.favicon, e.status, e.shopware_version,
       e.last_scraped_at, e.last_scraped_error, e.created_at, e.shop_id, e.organization_id,
       o.name AS organization_name,
       s.name AS shop_name, s.description AS shop_description
FROM environment e
JOIN organization o ON o.id = e.organization_id
LEFT JOIN shop s ON s.id = e.shop_id
WHERE e.organization_id = $1
ORDER BY e.name;

-- name: GetEnvironmentByID :one
SELECT e.*, o.name AS organization_name, s.name AS shop_name, s.description AS shop_description
FROM environment e
JOIN organization o ON o.id = e.organization_id
LEFT JOIN shop s ON s.id = e.shop_id
WHERE e.id = $1;

-- name: CreateEnvironment :one
INSERT INTO environment (organization_id, shop_id, name, url, client_id, client_secret, shopware_version, environment_token, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, NOW())
RETURNING id;

-- name: UpdateEnvironment :exec
UPDATE environment SET name = $1, url = $2, client_id = $3, client_secret = $4, ignores = $5, shop_id = $6 WHERE id = $7;

-- name: UpdateEnvironmentName :exec
UPDATE environment SET name = $1 WHERE id = $2;

-- name: DeleteEnvironment :exec
DELETE FROM environment WHERE id = $1;

-- name: GetEnvironmentExtensions :many
SELECT id, environment_id, name, label, active, version, latest_version, installed, rating_average, store_link, changelog, installed_at
FROM environment_extension WHERE environment_id = $1 ORDER BY name;

-- name: GetEnvironmentScheduledTasks :many
SELECT id, environment_id, task_id, name, status, interval, overdue, last_execution_time, next_execution_time
FROM environment_scheduled_task WHERE environment_id = $1 ORDER BY name;

-- name: GetEnvironmentQueues :many
SELECT id, environment_id, name, size FROM environment_queue WHERE environment_id = $1 ORDER BY name;

-- name: GetEnvironmentCache :one
SELECT id, environment_id, environment, http_cache, cache_adapter FROM environment_cache WHERE environment_id = $1;

-- name: GetEnvironmentChecks :many
SELECT id, environment_id, check_id, level, message, source, link FROM environment_check WHERE environment_id = $1;

-- name: GetEnvironmentSitespeeds :many
SELECT id, environment_id, deployment_id, created_at, ttfb, fully_loaded, largest_contentful_paint, first_contentful_paint, cumulative_layout_shift, transfer_size
FROM environment_sitespeed WHERE environment_id = $1 ORDER BY created_at DESC LIMIT 100;

-- name: GetEnvironmentChangelogs :many
SELECT id, environment_id, extensions, old_shopware_version, new_shopware_version, date
FROM environment_changelog WHERE environment_id = $1 ORDER BY date DESC LIMIT 10;

-- name: CountEnvironmentDeployments :one
SELECT COUNT(*)::int FROM deployment WHERE environment_id = $1;

-- name: UpdateEnvironmentSitespeedSettings :exec
UPDATE environment SET sitespeed_enabled = $1, sitespeed_urls = $2 WHERE id = $3;

-- name: GetAllEnvironments :many
SELECT e.id, e.name, e.url, e.client_id, e.client_secret, e.shopware_version,
       e.organization_id, e.ignores, e.last_scraped_at, e.last_scraped_error,
       e.connection_issue_count, e.environment_token, e.sitespeed_enabled, e.sitespeed_urls
FROM environment e;

-- name: GetEnvironmentForScrape :one
SELECT e.id, e.name, e.url, e.client_id, e.client_secret, e.shopware_version,
       e.organization_id, e.ignores, e.last_scraped_at, e.last_scraped_error,
       e.connection_issue_count, e.environment_token, e.sitespeed_enabled, e.sitespeed_urls
FROM environment e WHERE e.id = $1;

-- name: UpdateEnvironmentAfterScrape :exec
UPDATE environment SET status = $1, shopware_version = $2, last_scraped_at = NOW(), last_scraped_error = $3,
       favicon = $4, environment_image = $5, last_changelog = $6, connection_issue_count = 0
WHERE id = $7;

-- name: UpdateEnvironmentImage :exec
UPDATE environment SET environment_image = $1 WHERE id = $2;

-- name: UpdateEnvironmentScrapeError :exec
UPDATE environment SET last_scraped_error = $1, last_scraped_at = NOW(), connection_issue_count = connection_issue_count + 1 WHERE id = $2;

-- name: UpsertEnvironmentExtension :exec
INSERT INTO environment_extension (environment_id, name, label, active, version, latest_version, installed, rating_average, store_link, changelog, installed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
ON CONFLICT (environment_id, name) DO UPDATE SET label = $3, active = $4, version = $5, latest_version = $6, installed = $7, rating_average = $8, store_link = $9, changelog = $10, installed_at = $11;

-- name: DeleteEnvironmentExtensionsNotIn :exec
DELETE FROM environment_extension WHERE environment_id = $1 AND name != ALL($2::text[]);

-- name: UpsertEnvironmentQueue :exec
INSERT INTO environment_queue (environment_id, name, size) VALUES ($1, $2, $3)
ON CONFLICT (environment_id, name) DO UPDATE SET size = $3;

-- name: DeleteEnvironmentQueuesNotIn :exec
DELETE FROM environment_queue WHERE environment_id = $1 AND name != ALL($2::text[]);

-- name: UpsertEnvironmentScheduledTask :exec
INSERT INTO environment_scheduled_task (environment_id, task_id, name, status, interval, overdue, last_execution_time, next_execution_time)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, task_id) DO UPDATE SET name = $3, status = $4, interval = $5, overdue = $6, last_execution_time = $7, next_execution_time = $8;

-- name: DeleteEnvironmentScheduledTasksNotIn :exec
DELETE FROM environment_scheduled_task WHERE environment_id = $1 AND task_id != ALL($2::text[]);

-- name: UpsertEnvironmentCache :exec
INSERT INTO environment_cache (environment_id, environment, http_cache, cache_adapter)
VALUES ($1, $2, $3, $4)
ON CONFLICT (environment_id) DO UPDATE SET environment = $2, http_cache = $3, cache_adapter = $4;

-- name: DeleteEnvironmentChecks :exec
DELETE FROM environment_check WHERE environment_id = $1;

-- name: InsertEnvironmentCheck :exec
INSERT INTO environment_check (environment_id, check_id, level, message, source, link)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (environment_id, check_id) DO UPDATE SET level = $3, message = $4, source = $5, link = $6;

-- name: InsertEnvironmentChangelog :exec
INSERT INTO environment_changelog (environment_id, extensions, old_shopware_version, new_shopware_version, date)
VALUES ($1, $2, $3, $4, NOW());

-- name: UpdateEnvironmentActiveDeployment :exec
UPDATE environment SET active_deployment_id = $1 WHERE id = $2;

-- name: GetEnvironmentCredentials :one
SELECT id, url, client_id, client_secret, environment_token FROM environment WHERE id = $1;

-- name: GetEnvironmentsWithSitespeedEnabled :many
SELECT e.id, e.url, e.sitespeed_urls, e.active_deployment_id, e.environment_token FROM environment e WHERE e.sitespeed_enabled = true;

-- name: InsertEnvironmentSitespeed :exec
INSERT INTO environment_sitespeed (environment_id, deployment_id, created_at, ttfb, fully_loaded, largest_contentful_paint, first_contentful_paint, cumulative_layout_shift, transfer_size)
VALUES ($1, $2, NOW(), $3, $4, $5, $6, $7, $8);

-- name: GetEnvironmentNotificationSubscribers :many
SELECT u.id, u.name, u.email
FROM "user" u
JOIN member m ON m.user_id = u.id
WHERE m.organization_id = @organization_id
  AND u.notifications @> to_jsonb(ARRAY['environment-' || @environment_id::text])::jsonb;
