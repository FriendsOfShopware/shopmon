-- name: ListDeployments :many
SELECT id, environment_id, name, command, return_code, start_date, end_date, execution_time, reference, created_at
FROM deployment WHERE environment_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetDeploymentByID :one
SELECT id, environment_id, name, command, return_code, start_date, end_date, execution_time, composer, reference, created_at
FROM deployment WHERE id = $1 AND environment_id = $2;

-- name: CreateDeployment :one
INSERT INTO deployment (environment_id, name, command, return_code, start_date, end_date, execution_time, composer, reference, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
RETURNING id;

-- name: DeleteDeployment :exec
DELETE FROM deployment WHERE id = $1 AND environment_id = $2;
