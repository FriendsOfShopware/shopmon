-- Extended disclosure text (and CVSS/CWE metadata) fetched from GitHub
-- Security Advisories so users can read the full write-up in Shopmon.
ALTER TABLE composer_advisory
  ADD COLUMN "summary" text,
  ADD COLUMN "description" text,
  ADD COLUMN "cvss_score" double precision,
  ADD COLUMN "cvss_vector" text,
  ADD COLUMN "cwes" jsonb NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN "external_references" jsonb NOT NULL DEFAULT '[]'::jsonb,
  ADD COLUMN "details_source" text,
  ADD COLUMN "details_synced_at" timestamp;

CREATE INDEX idx_composer_advisory_details_pending
  ON composer_advisory (ghsa_id)
  WHERE ghsa_id IS NOT NULL AND ghsa_id <> '' AND details_synced_at IS NULL;
