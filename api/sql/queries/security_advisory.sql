-- name: UpsertSecurityAdvisory :exec
INSERT INTO security_advisory (
  advisory_id, origin, package_name, title, link, cve,
  affected_versions, source_name, source_remote_id, severity, reported_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, NOW()
)
ON CONFLICT (advisory_id) DO UPDATE SET
  origin = EXCLUDED.origin,
  package_name = EXCLUDED.package_name,
  title = EXCLUDED.title,
  link = EXCLUDED.link,
  cve = EXCLUDED.cve,
  affected_versions = EXCLUDED.affected_versions,
  source_name = EXCLUDED.source_name,
  source_remote_id = EXCLUDED.source_remote_id,
  severity = EXCLUDED.severity,
  reported_at = EXCLUDED.reported_at,
  updated_at = NOW();

-- name: ListSecurityAdvisories :many
SELECT advisory_id, origin, package_name, title, link, cve,
       affected_versions, source_name, source_remote_id, severity, reported_at, created_at, updated_at
FROM security_advisory
ORDER BY reported_at DESC NULLS LAST, advisory_id;

-- name: GetSecurityAdvisory :one
SELECT advisory_id, origin, package_name, title, link, cve,
       affected_versions, source_name, source_remote_id, severity, reported_at, created_at, updated_at
FROM security_advisory
WHERE advisory_id = $1;
