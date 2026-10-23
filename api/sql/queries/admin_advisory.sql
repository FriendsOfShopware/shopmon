-- On-demand preview against the latest inventory, independent of saved advisories.
-- name: ListAdminAdvisoryInventory :many
SELECT c.environment_id, e.name AS environment_name,
       e.shop_id, s.name AS shop_name,
       e.organization_id, o.name AS organization_name,
       c.package_name, c.version
FROM environment_sbom_component c
JOIN environment e ON e.id = c.environment_id
JOIN shop s ON s.id = e.shop_id
JOIN organization o ON o.id = e.organization_id
WHERE c.package_name = ANY(sqlc.arg('package_names')::text[])
ORDER BY o.name, o.id, s.name, s.id, e.name, e.id, c.package_name;
