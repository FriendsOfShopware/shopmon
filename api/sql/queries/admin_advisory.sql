-- On-demand preview against the latest inventory, independent of saved advisories.
-- Includes Composer SBOM components as well as runtime versions (php, mysql)
-- stored on the environment.
-- name: ListAdminAdvisoryInventory :many
SELECT * FROM (
  SELECT c.environment_id, e.name AS environment_name,
         e.shop_id, s.name AS shop_name,
         e.organization_id, o.name AS organization_name,
         c.package_name, c.version
  FROM environment_sbom_component c
  JOIN environment e ON e.id = c.environment_id
  JOIN shop s ON s.id = e.shop_id
  JOIN organization o ON o.id = e.organization_id
  WHERE c.package_name = ANY(sqlc.arg('package_names')::text[])
  UNION ALL
  SELECT e.id AS environment_id, e.name AS environment_name,
         e.shop_id, s.name AS shop_name,
         e.organization_id, o.name AS organization_name,
         'php'::text AS package_name, e.php_version AS version
  FROM environment e
  JOIN shop s ON s.id = e.shop_id
  JOIN organization o ON o.id = e.organization_id
  WHERE 'php' = ANY(sqlc.arg('package_names')::text[])
    AND e.php_version IS NOT NULL AND e.php_version <> ''
  UNION ALL
  SELECT e.id AS environment_id, e.name AS environment_name,
         e.shop_id, s.name AS shop_name,
         e.organization_id, o.name AS organization_name,
         'mysql'::text AS package_name,
         regexp_replace(e.database_version, '^MySQL\s+', '', 'i') AS version
  FROM environment e
  JOIN shop s ON s.id = e.shop_id
  JOIN organization o ON o.id = e.organization_id
  WHERE 'mysql' = ANY(sqlc.arg('package_names')::text[])
    AND e.database_version IS NOT NULL AND e.database_version <> ''
    AND e.database_version !~* '(mariadb|^(10|11)\.)'
  UNION ALL
  SELECT e.id AS environment_id, e.name AS environment_name,
         e.shop_id, s.name AS shop_name,
         e.organization_id, o.name AS organization_name,
         'mariadb'::text AS package_name,
         regexp_replace(e.database_version, '^MariaDB\s+', '', 'i') AS version
  FROM environment e
  JOIN shop s ON s.id = e.shop_id
  JOIN organization o ON o.id = e.organization_id
  WHERE 'mariadb' = ANY(sqlc.arg('package_names')::text[])
    AND e.database_version IS NOT NULL AND e.database_version <> ''
    AND e.database_version ~* '(mariadb|^(10|11)\.)'
) inv
ORDER BY organization_name, organization_id, shop_name, shop_id, environment_name, environment_id, package_name;
