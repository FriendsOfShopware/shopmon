-- Queries backing the store extension catalog sync job. The catalog is shared
-- across environments and refreshed by a dedicated queue job, never by the
-- environment scrapes themselves.

-- name: GetFreshStoreExtensionSyncNames :many
-- Names among the given set whose store lookup happened within the last hour
-- (hit or miss). These do not need another sync.
SELECT extension_name FROM store_extension_sync
WHERE extension_name = ANY($1::text[]) AND last_synced_at > NOW() - INTERVAL '1 hour';

-- name: UpsertStoreExtensionSyncStates :exec
-- Marks all given names as looked up now, whether or not the store knew them.
INSERT INTO store_extension_sync (extension_name, last_synced_at)
SELECT unnest($1::text[]), NOW()
ON CONFLICT (extension_name) DO UPDATE SET last_synced_at = NOW();

-- name: GetStoreExtensionNamesIn :many
-- Which of the given names exist in the shared store catalog. Membership here
-- is what classifies a scraped extension as store-known.
SELECT name FROM store_extension WHERE name = ANY($1::text[]);

-- name: GetStoreExtensionCompatibility :many
-- Compatible latest versions for the given extensions on one Shopware version.
-- A row with NULL latest_version means "checked, no compatible release".
SELECT extension_name, latest_version FROM store_extension_compatibility
WHERE extension_name = ANY($1::text[]) AND shopware_version = $2;

-- name: UpsertStoreExtensionCompatibility :exec
INSERT INTO store_extension_compatibility (extension_name, shopware_version, latest_version)
VALUES ($1, $2, $3)
ON CONFLICT (extension_name, shopware_version) DO UPDATE SET latest_version = $3
WHERE store_extension_compatibility.latest_version IS DISTINCT FROM $3;

-- name: GetDistinctEnvironmentShopwareVersions :many
SELECT DISTINCT shopware_version FROM environment WHERE shopware_version != '';

-- name: GetStoreExtensionChangelogs :many
-- All catalog versions of one extension with both changelog translations, used
-- to attach update changelogs to extension diffs directly from the database.
SELECT sev.version, sev.released_at,
       en.changelog AS changelog_en, de.changelog AS changelog_de
FROM store_extension_version sev
LEFT JOIN store_extension_version_translation en ON en.extension_version_id = sev.id AND en.language = 'en'
LEFT JOIN store_extension_version_translation de ON de.extension_version_id = sev.id AND de.language = 'de'
WHERE sev.extension_name = $1;

-- name: CleanupOrphanedStoreExtensionSyncStates :exec
-- Sync bookkeeping for names no environment has anymore.
DELETE FROM store_extension_sync
WHERE last_synced_at < NOW() - INTERVAL '7 days'
  AND NOT EXISTS (
    SELECT 1 FROM environment_store_extension ese WHERE ese.extension_name = store_extension_sync.extension_name
  )
  AND NOT EXISTS (
    SELECT 1 FROM environment_extension ee WHERE ee.name = store_extension_sync.extension_name
  );

-- name: CleanupUnusedStoreExtensionCompatibility :exec
-- Compatibility rows for Shopware versions no environment runs anymore.
DELETE FROM store_extension_compatibility
WHERE shopware_version NOT IN (SELECT DISTINCT shopware_version FROM environment);
