-- Revert user notification keys: "environment-123" -> "shop-123"
UPDATE "user" SET notifications = (
  SELECT COALESCE(jsonb_agg(
    CASE WHEN elem #>> '{}' LIKE 'environment-%'
      THEN to_jsonb('shop-' || substring(elem #>> '{}' FROM 13))
      ELSE elem
    END
  ), '[]'::jsonb) FROM jsonb_array_elements(notifications) elem
) WHERE notifications::text LIKE '%environment-%';

-- Rename unique constraints back
ALTER TABLE environment_cache RENAME CONSTRAINT environment_cache_environment_id_unique TO shop_cache_shop_id_unique;
ALTER TABLE environment_check RENAME CONSTRAINT environment_check_environment_id_check_id_unique TO shop_check_shop_id_check_id_unique;
ALTER TABLE environment_scheduled_task RENAME CONSTRAINT environment_scheduled_task_environment_id_task_id_unique TO shop_scheduled_task_shop_id_task_id_unique;
ALTER TABLE environment_queue RENAME CONSTRAINT environment_queue_environment_id_name_unique TO shop_queue_shop_id_name_unique;
ALTER TABLE environment_extension RENAME CONSTRAINT environment_extension_environment_id_name_unique TO shop_extension_shop_id_name_unique;

-- Rename indexes back
ALTER INDEX idx_environment_extension_environment RENAME TO idx_shop_extension_shop;
ALTER INDEX idx_deployment_environment RENAME TO idx_deployment_shop;
ALTER INDEX idx_environment_organization RENAME TO idx_shop_organization;

-- Rename "shop_api_key" back to "project_api_key"
ALTER TABLE shop_api_key RENAME COLUMN shop_id TO project_id;
ALTER TABLE shop_api_key RENAME TO project_api_key;

-- Rename "shop" table back to "project"
ALTER TABLE shop RENAME TO project;

-- Drop FK before renaming environment back to shop
ALTER TABLE deployment DROP CONSTRAINT fk_deployment_environment;

-- Rename "environment" back to "shop"
ALTER TABLE environment RENAME COLUMN environment_image TO shop_image;
ALTER TABLE environment RENAME COLUMN environment_token TO shop_token;
ALTER TABLE environment RENAME COLUMN shop_id TO project_id;
ALTER TABLE environment RENAME TO shop;

-- Re-add both FKs from deployment -> shop
ALTER TABLE deployment RENAME COLUMN environment_id TO shop_id;
ALTER TABLE deployment ADD CONSTRAINT fk_deployment_shop
  FOREIGN KEY (shop_id) REFERENCES shop(id) ON DELETE CASCADE;
ALTER TABLE deployment ADD CONSTRAINT deployment_shop_id_shop_id_fk
  FOREIGN KEY (shop_id) REFERENCES shop(id) ON DELETE CASCADE;

-- Rename columns back in child tables
ALTER TABLE environment_cache RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_sitespeed RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_changelog RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_check RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_scheduled_task RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_queue RENAME COLUMN environment_id TO shop_id;
ALTER TABLE environment_extension RENAME COLUMN environment_id TO shop_id;

-- Rename child tables back
ALTER TABLE environment_cache RENAME TO shop_cache;
ALTER TABLE environment_sitespeed RENAME TO shop_sitespeed;
ALTER TABLE environment_changelog RENAME TO shop_changelog;
ALTER TABLE environment_check RENAME TO shop_check;
ALTER TABLE environment_scheduled_task RENAME TO shop_scheduled_task;
ALTER TABLE environment_queue RENAME TO shop_queue;
ALTER TABLE environment_extension RENAME TO shop_extension;
