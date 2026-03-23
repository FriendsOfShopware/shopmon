-- name: ListApiKeysByProject :many
SELECT id, project_id, name, scopes, created_at, last_used_at
FROM project_api_key WHERE project_id = $1 ORDER BY created_at DESC;

-- name: CreateApiKey :one
INSERT INTO project_api_key (id, project_id, name, token, scopes, created_at)
VALUES ($1, $2, $3, $4, $5, NOW())
RETURNING id;

-- name: DeleteApiKey :exec
DELETE FROM project_api_key WHERE id = $1 AND project_id = $2;

-- name: GetApiKeyByToken :one
SELECT pak.id, pak.project_id, pak.name, pak.token, pak.scopes, pak.created_at, pak.last_used_at,
       p.organization_id
FROM project_api_key pak
JOIN project p ON p.id = pak.project_id
WHERE pak.token = $1;

-- name: UpdateApiKeyLastUsed :exec
UPDATE project_api_key SET last_used_at = NOW() WHERE id = $1;
