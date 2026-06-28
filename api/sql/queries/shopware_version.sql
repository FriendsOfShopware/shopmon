-- name: UpsertShopwareVersion :exec
INSERT INTO shopware_version (version, release_date, php_versions, title, body, updated_at)
VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  php_versions = EXCLUDED.php_versions,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  updated_at = NOW();

-- name: GetKnownShopwareVersions :many
SELECT version FROM shopware_version;

-- name: ListShopwareVersions :many
SELECT version, release_date, php_versions
FROM shopware_version
ORDER BY release_date DESC, version DESC;
