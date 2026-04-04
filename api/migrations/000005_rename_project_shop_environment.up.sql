-- Rename shop child tables to environment_*
ALTER TABLE shop_extension RENAME TO environment_extension;
ALTER TABLE shop_queue RENAME TO environment_queue;
ALTER TABLE shop_scheduled_task RENAME TO environment_scheduled_task;
ALTER TABLE shop_check RENAME TO environment_check;
ALTER TABLE shop_changelog RENAME TO environment_changelog;
ALTER TABLE shop_sitespeed RENAME TO environment_sitespeed;
ALTER TABLE shop_cache RENAME TO environment_cache;

-- Rename columns in child tables: shop_id -> environment_id
ALTER TABLE environment_extension RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_queue RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_scheduled_task RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_check RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_changelog RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_sitespeed RENAME COLUMN shop_id TO environment_id;
ALTER TABLE environment_cache RENAME COLUMN shop_id TO environment_id;

-- Rename deployment.shop_id -> deployment.environment_id
ALTER TABLE deployment RENAME COLUMN shop_id TO environment_id;

-- Drop FK constraints on deployment before renaming shop table
ALTER TABLE deployment DROP CONSTRAINT fk_deployment_shop;
ALTER TABLE deployment DROP CONSTRAINT deployment_shop_id_shop_id_fk;

-- Rename the main "shop" table to "environment"
ALTER TABLE shop RENAME TO environment;
ALTER TABLE environment RENAME COLUMN project_id TO shop_id;
ALTER TABLE environment RENAME COLUMN shop_token TO environment_token;
ALTER TABLE environment RENAME COLUMN shop_image TO environment_image;

-- Re-add FK from deployment.environment_id -> environment.id
ALTER TABLE deployment ADD CONSTRAINT fk_deployment_environment
  FOREIGN KEY (environment_id) REFERENCES environment(id) ON DELETE CASCADE;

-- Rename "project" table to "shop"
ALTER TABLE project RENAME TO shop;

-- Rename "project_api_key" table to "shop_api_key"
ALTER TABLE project_api_key RENAME TO shop_api_key;
ALTER TABLE shop_api_key RENAME COLUMN project_id TO shop_id;

-- Rename indexes
ALTER INDEX idx_shop_organization RENAME TO idx_environment_organization;
ALTER INDEX idx_deployment_shop RENAME TO idx_deployment_environment;
ALTER INDEX idx_shop_extension_shop RENAME TO idx_environment_extension_environment;

-- Rename unique constraints on child tables
ALTER TABLE environment_extension RENAME CONSTRAINT shop_extension_shop_id_name_unique TO environment_extension_environment_id_name_unique;
ALTER TABLE environment_queue RENAME CONSTRAINT shop_queue_shop_id_name_unique TO environment_queue_environment_id_name_unique;
ALTER TABLE environment_scheduled_task RENAME CONSTRAINT shop_scheduled_task_shop_id_task_id_unique TO environment_scheduled_task_environment_id_task_id_unique;
ALTER TABLE environment_check RENAME CONSTRAINT shop_check_shop_id_check_id_unique TO environment_check_environment_id_check_id_unique;
ALTER TABLE environment_cache RENAME CONSTRAINT shop_cache_shop_id_unique TO environment_cache_environment_id_unique;

-- Update user notification keys: "shop-123" -> "environment-123"
UPDATE "user" SET notifications = (
  SELECT COALESCE(jsonb_agg(
    CASE WHEN elem #>> '{}' LIKE 'shop-%'
      THEN to_jsonb('environment-' || substring(elem #>> '{}' FROM 6))
      ELSE elem
    END
  ), '[]'::jsonb) FROM jsonb_array_elements(notifications) elem
) WHERE notifications::text LIKE '%shop-%';
