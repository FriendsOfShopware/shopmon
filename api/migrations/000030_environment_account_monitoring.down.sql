DROP TABLE IF EXISTS "environment_account_event";
DROP TABLE IF EXISTS "environment_account_state";
DROP TABLE IF EXISTS "environment_account";
ALTER TABLE "environment" DROP COLUMN IF EXISTS "account_monitoring_enabled";
