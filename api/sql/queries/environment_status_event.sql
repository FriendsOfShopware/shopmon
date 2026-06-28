-- name: InsertEnvironmentStatusEvent :exec
INSERT INTO environment_status_event (environment_id, old_status, new_status, reasons)
VALUES ($1, $2, $3, $4);

-- name: ListEnvironmentStatusEvents :many
SELECT id, environment_id, old_status, new_status, reasons, created_at
FROM environment_status_event
WHERE environment_id = $1
ORDER BY created_at DESC
LIMIT 50;
