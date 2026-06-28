-- Central, browsable catalog of security advisories. Imported from the Packagist
-- Security Advisories API for the first-party Shopware packages. The schema is
-- shaped so future work slots in without a migration: `origin` distinguishes
-- imported advisories from user-submitted plugin disclosures, and
-- `affected_versions` (a composer constraint) enables matching against tracked
-- environments later.
CREATE TABLE IF NOT EXISTS "security_advisory" (
  "advisory_id"       text PRIMARY KEY NOT NULL,
  "origin"            text NOT NULL DEFAULT 'packagist',
  "package_name"      text NOT NULL,
  "title"             text NOT NULL,
  "link"              text,
  "cve"               text,
  "affected_versions" text NOT NULL,
  "source_name"       text,
  "source_remote_id"  text,
  "severity"          text,
  "reported_at"       timestamp,
  "created_at"        timestamp NOT NULL DEFAULT NOW(),
  "updated_at"        timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_advisory_package ON security_advisory (package_name);
CREATE INDEX IF NOT EXISTS idx_security_advisory_reported_at ON security_advisory (reported_at DESC);
