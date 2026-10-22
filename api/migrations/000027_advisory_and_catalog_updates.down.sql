-- Manual advisories cannot survive rollback: their packages may have NULL
-- Packagist ids, which the restored NOT NULL constraint forbids. Deletion
-- cascades to packages, matches, notification markers, and suppressions.
DELETE FROM composer_advisory WHERE origin IN ('shopmon', 'github');

ALTER TABLE composer_advisory_package
  ALTER COLUMN packagist_advisory_id SET NOT NULL;

DROP INDEX IF EXISTS idx_composer_advisory_origin;

ALTER TABLE composer_advisory
  DROP CONSTRAINT IF EXISTS composer_advisory_origin_check,
  DROP COLUMN IF EXISTS created_by,
  DROP COLUMN IF EXISTS origin;

DROP SEQUENCE IF EXISTS shopmon_advisory_seq;

ALTER TABLE store_extension DROP COLUMN available_versions;
