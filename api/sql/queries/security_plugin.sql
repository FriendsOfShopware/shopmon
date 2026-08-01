-- SwagPlatformSecurity backport map, derived from the store changelog.

-- name: UpsertSecurityPluginFix :exec
INSERT INTO security_plugin_fix (
  ghsa_id, plugin_branch, plugin_version, shopware_branch, released_at, synced_at
) VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (ghsa_id, plugin_branch) DO UPDATE SET
  plugin_version = EXCLUDED.plugin_version,
  shopware_branch = EXCLUDED.shopware_branch,
  released_at = EXCLUDED.released_at,
  synced_at = NOW();

-- Rows for branches that were re-derived but no longer name the GHSA. Scoped to
-- the observed branches so a partial sync never deletes another branch's map.
-- name: DeleteSecurityPluginFixesNotIn :exec
DELETE FROM security_plugin_fix
WHERE plugin_branch = ANY(sqlc.arg('branches')::text[])
  AND ghsa_id != ALL(sqlc.arg('ghsa_ids')::text[]);

-- Records why a branch's parse result was rejected (e.g. the sharp-drop
-- guard) without touching its accepted fixes or counters.
-- name: SetSecurityPluginCoverageError :exec
UPDATE security_plugin_coverage SET last_error = $2 WHERE plugin_branch = $1;

-- name: UpsertSecurityPluginCoverage :exec
INSERT INTO security_plugin_coverage (
  plugin_branch, min_parsed_version, max_parsed_version, ghsa_count, parsed_at, last_error
) VALUES ($1, $2, $3, $4, NOW(), $5)
ON CONFLICT (plugin_branch) DO UPDATE SET
  min_parsed_version = EXCLUDED.min_parsed_version,
  max_parsed_version = EXCLUDED.max_parsed_version,
  ghsa_count = EXCLUDED.ghsa_count,
  parsed_at = NOW(),
  last_error = EXCLUDED.last_error;

-- name: ListSecurityPluginCoverage :many
SELECT * FROM security_plugin_coverage ORDER BY plugin_branch;

-- name: ListSecurityPluginFixes :many
SELECT * FROM security_plugin_fix ORDER BY ghsa_id, plugin_branch;

-- name: ListSecurityPluginFixesForGhsas :many
SELECT * FROM security_plugin_fix
WHERE ghsa_id = ANY(sqlc.arg('ghsa_ids')::text[])
ORDER BY ghsa_id, plugin_branch;
