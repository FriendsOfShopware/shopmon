ALTER TABLE advisory_suppression
  DROP CONSTRAINT IF EXISTS advisory_suppression_environment_shop_fkey;
ALTER TABLE environment
  DROP CONSTRAINT IF EXISTS environment_shop_id_id_key;

DROP TABLE IF EXISTS "security_plugin_coverage";
DROP TABLE IF EXISTS "security_plugin_fix";
DROP TABLE IF EXISTS "advisory_suppression";
DROP TABLE IF EXISTS "environment_advisory_notified";
DROP TABLE IF EXISTS "environment_advisory_match";
DROP TABLE IF EXISTS "environment_sbom_state";
DROP TABLE IF EXISTS "environment_sbom_component";
DROP TABLE IF EXISTS "composer_advisory_sync_state";
DROP TABLE IF EXISTS "composer_advisory_package";
DROP TABLE IF EXISTS "composer_advisory";
