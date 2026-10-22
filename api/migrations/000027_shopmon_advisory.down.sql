-- Shopmon-authored rows cannot survive the rollback: their package rows carry
-- a NULL packagist_advisory_id, which the restored NOT NULL forbids. Deleting
-- them cascades to packages, matches, notifications, and suppressions.
DELETE FROM "composer_advisory" WHERE origin = 'shopmon';

ALTER TABLE "composer_advisory_package"
  ALTER COLUMN "packagist_advisory_id" SET NOT NULL;

DROP INDEX IF EXISTS idx_composer_advisory_origin;

ALTER TABLE "composer_advisory"
  DROP CONSTRAINT IF EXISTS composer_advisory_origin_check,
  DROP COLUMN IF EXISTS "created_by",
  DROP COLUMN IF EXISTS "origin";

DROP SEQUENCE IF EXISTS shopmon_advisory_seq;
