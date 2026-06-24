-- name: CreateAuditLog :exec
INSERT INTO audit_log (actor_user_id, action, target_user_id, detail, ip_address)
VALUES ($1, $2, $3, $4, $5);

-- name: ListAuditLog :many
SELECT id, actor_user_id, action, target_user_id, detail, ip_address, created_at
FROM audit_log
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminListUserAuditLog :many
SELECT a.id, a.actor_user_id, actor.name AS actor_name, actor.email AS actor_email,
       a.action, a.target_user_id, target.name AS target_name, target.email AS target_email,
       a.detail, a.ip_address, a.created_at
FROM audit_log a
LEFT JOIN "user" actor ON actor.id = a.actor_user_id
LEFT JOIN "user" target ON target.id = a.target_user_id
WHERE a.actor_user_id = $1 OR a.target_user_id = $1
ORDER BY a.created_at DESC
LIMIT $2;

-- name: AdminListAuditLog :many
SELECT a.id, a.actor_user_id, actor.name AS actor_name, actor.email AS actor_email,
       a.action, a.target_user_id, target.name AS target_name, target.email AS target_email,
       a.detail, a.ip_address, a.created_at
FROM audit_log a
LEFT JOIN "user" actor ON actor.id = a.actor_user_id
LEFT JOIN "user" target ON target.id = a.target_user_id
WHERE (sqlc.narg('action')::text IS NULL OR a.action = sqlc.narg('action'))
  AND (sqlc.narg('actor_user_id')::text IS NULL OR a.actor_user_id = sqlc.narg('actor_user_id'))
  AND (sqlc.narg('target_user_id')::text IS NULL OR a.target_user_id = sqlc.narg('target_user_id'))
ORDER BY a.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountAuditLog :one
SELECT COUNT(*)::int FROM audit_log a
WHERE (sqlc.narg('action')::text IS NULL OR a.action = sqlc.narg('action'))
  AND (sqlc.narg('actor_user_id')::text IS NULL OR a.actor_user_id = sqlc.narg('actor_user_id'))
  AND (sqlc.narg('target_user_id')::text IS NULL OR a.target_user_id = sqlc.narg('target_user_id'));
