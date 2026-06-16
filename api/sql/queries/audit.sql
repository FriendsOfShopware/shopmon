-- name: CreateAuditLog :exec
INSERT INTO audit_log (actor_user_id, action, target_user_id, detail, ip_address)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditLog :many
SELECT id, actor_user_id, action, target_user_id, detail, ip_address, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;
