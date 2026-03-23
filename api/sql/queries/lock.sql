-- name: AcquireLock :exec
INSERT INTO lock (key, expires, created_at) VALUES ($1, $2, NOW())
ON CONFLICT (key) DO UPDATE SET expires = $2, created_at = NOW()
WHERE lock.expires < NOW();

-- name: IsLocked :one
SELECT COUNT(*) > 0 AS locked FROM lock WHERE key = $1 AND expires > NOW();

-- name: ReleaseLock :exec
DELETE FROM lock WHERE key = $1;

-- name: CleanupExpiredLocks :exec
DELETE FROM lock WHERE expires < NOW();

-- name: CleanupExpiredInvitations :exec
DELETE FROM invitation WHERE expires_at < NOW() AND status = 'pending';
