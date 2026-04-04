-- name: ListApiKeysByShop :many
SELECT id, shop_id, name, scopes, created_at, last_used_at
FROM shop_api_key WHERE shop_id = $1 ORDER BY created_at DESC;

-- name: CreateApiKey :one
INSERT INTO shop_api_key (id, shop_id, name, token, scopes, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id;

-- name: DeleteApiKey :exec
DELETE FROM shop_api_key WHERE id = $1 AND shop_id = $2;

-- name: GetApiKeyByToken :one
SELECT sak.id, sak.shop_id, sak.name, sak.token, sak.scopes, sak.created_at, sak.last_used_at,
       s.organization_id
FROM shop_api_key sak
JOIN shop s ON s.id = sak.shop_id
WHERE sak.token = $1;

-- name: UpdateApiKeyLastUsed :exec
UPDATE shop_api_key SET last_used_at = NOW() WHERE id = $1;
