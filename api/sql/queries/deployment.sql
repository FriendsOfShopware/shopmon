-- name: ListDeployments :many
SELECT id, shop_id, name, command, return_code, start_date, end_date, execution_time, reference, created_at
FROM deployment WHERE shop_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;

-- name: GetDeploymentByID :one
SELECT id, shop_id, name, command, return_code, start_date, end_date, execution_time, composer, reference, created_at
FROM deployment WHERE id = $1 AND shop_id = $2;

-- name: CreateDeployment :one
INSERT INTO deployment (shop_id, name, command, return_code, start_date, end_date, execution_time, composer, reference, created_at)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, NOW())
RETURNING id;

-- name: DeleteDeployment :exec
DELETE FROM deployment WHERE id = $1 AND shop_id = $2;
