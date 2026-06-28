-- name: UpsertShopwareVersion :exec
INSERT INTO shopware_version (version, release_date, title, body, updated_at)
VALUES ($1, $2, $3, $4, NOW())
ON CONFLICT (version) DO UPDATE SET
  release_date = EXCLUDED.release_date,
  title = EXCLUDED.title,
  body = EXCLUDED.body,
  updated_at = NOW();

-- name: GetKnownShopwareVersions :many
SELECT version FROM shopware_version;

-- name: ListShopwareVersions :many
SELECT version
FROM shopware_version
ORDER BY release_date DESC, version DESC;
