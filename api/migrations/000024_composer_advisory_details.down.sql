DROP INDEX IF EXISTS idx_composer_advisory_details_pending;
ALTER TABLE composer_advisory
  DROP COLUMN IF EXISTS "summary",
  DROP COLUMN IF EXISTS "description",
  DROP COLUMN IF EXISTS "cvss_score",
  DROP COLUMN IF EXISTS "cvss_vector",
  DROP COLUMN IF EXISTS "cwes",
  DROP COLUMN IF EXISTS "external_references",
  DROP COLUMN IF EXISTS "details_source",
  DROP COLUMN IF EXISTS "details_synced_at";
