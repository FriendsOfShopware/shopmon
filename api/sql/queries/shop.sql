-- name: ListShopsByOrganization :many
SELECT id, organization_id, name, description, git_url, default_environment_id, created_at, updated_at
FROM shop WHERE organization_id = $1 ORDER BY name;

-- name: GetShopByID :one
SELECT id, organization_id, name, description, git_url, default_environment_id, created_at, updated_at
FROM shop WHERE id = $1;

-- name: CreateShop :one
INSERT INTO shop (organization_id, name, description, git_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id;

-- name: UpdateShop :exec
UPDATE shop SET name = $1, description = $2, git_url = $3, updated_at = NOW() WHERE id = $4 AND organization_id = $5;

-- name: DeleteShop :exec
DELETE FROM shop WHERE id = $1 AND organization_id = $2;

-- name: CountEnvironmentsInShop :one
SELECT COUNT(*)::int FROM environment WHERE shop_id = $1;

-- name: GetShopOrganizationID :one
SELECT organization_id FROM shop WHERE id = $1;

-- name: SetShopDefaultEnvironment :exec
UPDATE shop SET default_environment_id = $1, updated_at = NOW() WHERE id = $2 AND organization_id = $3;
