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

-- name: ReassignShopDefaultEnvironment :exec
-- Before deleting an environment, move any shop that uses it as its default to
-- another environment of the same shop (lowest id), or NULL if none remain. This
-- avoids violating the RESTRICT foreign key shop.default_environment_id -> environment.
UPDATE shop SET default_environment_id = (
    SELECT e.id FROM environment e
    WHERE e.shop_id = shop.id AND e.id != $1
    ORDER BY e.id LIMIT 1
), updated_at = NOW()
WHERE default_environment_id = $1;

-- name: DeleteEnvironment :exec
DELETE FROM environment WHERE id = $1;

-- name: GetEnvironmentExtensions :many
-- Extensions unknown to the Shopware store (everything else lives in the
-- normalized store_extension* tables, linked via environment_store_extension).
SELECT id, environment_id, name, label, active, version, latest_version, installed, installed_at
FROM environment_extension WHERE environment_id = $1 ORDER BY name;

-- name: GetEnvironmentStoreExtensions :many
SELECT ese.id, ese.environment_id, ese.extension_name, ese.label, ese.active, ese.version,
       ese.latest_version, ese.installed, ese.installed_at,
       se.store_link, se.rating_average, se.icon_url,
       se.label_en, se.label_de, se.short_description_en, se.short_description_de
FROM environment_store_extension ese
JOIN store_extension se ON se.name = ese.extension_name
WHERE ese.environment_id = $1
ORDER BY ese.extension_name;

-- name: GetEnvironmentStoreExtensionChangelogs :many
SELECT sev.extension_name, sev.version, sev.changelog_en, sev.changelog_de, sev.released_at
FROM store_extension_version sev
JOIN environment_store_extension ese ON ese.extension_name = sev.extension_name
WHERE ese.environment_id = $1
ORDER BY sev.extension_name, sev.version;

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
INSERT INTO environment_extension (environment_id, name, label, active, version, latest_version, installed, installed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, name) DO UPDATE SET label = $3, active = $4, version = $5, latest_version = $6, installed = $7, installed_at = $8;

-- name: DeleteEnvironmentExtensionsNotIn :exec
DELETE FROM environment_extension WHERE environment_id = $1 AND name != ALL($2::text[]);

-- name: UpsertStoreExtension :exec
-- Localized text fields use COALESCE so a failed locale fetch (NULL value) does
-- not erase the previously stored translation.
INSERT INTO store_extension (
  name, store_id, icon_url, producer_name, producer_website, rating_average, store_link,
  release_date, label_en, label_de, short_description_en, short_description_de,
  description_en, description_de, installation_manual_en, installation_manual_de,
  latest_version, last_refreshed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, NOW())
ON CONFLICT (name) DO UPDATE SET
  store_id = $2, icon_url = $3, producer_name = $4, producer_website = $5, rating_average = $6,
  store_link = $7, release_date = $8,
  label_en = COALESCE($9, store_extension.label_en),
  label_de = COALESCE($10, store_extension.label_de),
  short_description_en = COALESCE($11, store_extension.short_description_en),
  short_description_de = COALESCE($12, store_extension.short_description_de),
  description_en = COALESCE($13, store_extension.description_en),
  description_de = COALESCE($14, store_extension.description_de),
  installation_manual_en = COALESCE($15, store_extension.installation_manual_en),
  installation_manual_de = COALESCE($16, store_extension.installation_manual_de),
  latest_version = $17,
  last_refreshed_at = NOW();

-- name: UpsertStoreExtensionVersion :exec
-- COALESCE keeps the existing per-language changelog text and date when a locale
-- fetch fails (NULL), so a transient single-locale store outage does not erase a
-- previously stored translation.
INSERT INTO store_extension_version (extension_name, version, changelog_en, changelog_de, released_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (extension_name, version) DO UPDATE SET
  changelog_en = COALESCE($3, store_extension_version.changelog_en),
  changelog_de = COALESCE($4, store_extension_version.changelog_de),
  released_at = COALESCE($5, store_extension_version.released_at);

-- name: UpsertStoreExtensionImage :exec
INSERT INTO store_extension_image (extension_name, url, preview, priority)
VALUES ($1, $2, $3, $4)
ON CONFLICT (extension_name, url) DO UPDATE SET preview = $3, priority = $4;

-- name: DeleteStoreExtensionImagesNotIn :exec
DELETE FROM store_extension_image WHERE extension_name = $1 AND url != ALL($2::text[]);

-- name: UpsertEnvironmentStoreExtension :exec
INSERT INTO environment_store_extension (environment_id, extension_name, label, version, latest_version, active, installed, installed_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (environment_id, extension_name) DO UPDATE SET
  label = $3, version = $4, latest_version = $5, active = $6, installed = $7, installed_at = $8;

-- name: DeleteEnvironmentStoreExtensionsNotIn :exec
DELETE FROM environment_store_extension WHERE environment_id = $1 AND extension_name != ALL($2::text[]);

-- name: GetEnvironmentStoreExtensionImages :many
SELECT sei.extension_name, sei.url, sei.preview, sei.priority
FROM store_extension_image sei
JOIN environment_store_extension ese ON ese.extension_name = sei.extension_name
WHERE ese.environment_id = $1
ORDER BY sei.extension_name, sei.priority DESC;

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
