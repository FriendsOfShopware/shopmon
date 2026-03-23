-- name: ListProjectsByOrganization :many
SELECT id, organization_id, name, description, git_url, created_at, updated_at
FROM project WHERE organization_id = $1 ORDER BY name;

-- name: GetProjectByID :one
SELECT id, organization_id, name, description, git_url, created_at, updated_at
FROM project WHERE id = $1;

-- name: CreateProject :one
INSERT INTO project (organization_id, name, description, git_url, created_at, updated_at)
VALUES ($1, $2, $3, $4, NOW(), NOW())
RETURNING id;

-- name: UpdateProject :exec
UPDATE project SET name = $1, description = $2, git_url = $3, updated_at = NOW() WHERE id = $4 AND organization_id = $5;

-- name: DeleteProject :exec
DELETE FROM project WHERE id = $1 AND organization_id = $2;

-- name: CountShopsInProject :one
SELECT COUNT(*)::int FROM shop WHERE project_id = $1;

-- name: GetProjectOrganizationID :one
SELECT organization_id FROM project WHERE id = $1;
