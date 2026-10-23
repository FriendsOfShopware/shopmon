-- Read the latest collected inventory directly so draft imports and edited
-- ranges have impact counts before the background rematch has run.
-- name: ListAdminAdvisoryInventory :many
SELECT c.environment_id, e.shop_id, c.package_name, c.version
FROM environment_sbom_component c
JOIN environment e ON e.id = c.environment_id
WHERE c.package_name = ANY(sqlc.arg('package_names')::text[]);

-- name: ListActiveSuppressionsForAdvisories :many
SELECT advisory_id, shop_id, environment_id
FROM advisory_suppression
WHERE advisory_id = ANY(sqlc.arg('advisory_ids')::text[])
  AND revoked_at IS NULL
  AND (expires_at IS NULL OR expires_at > NOW());
