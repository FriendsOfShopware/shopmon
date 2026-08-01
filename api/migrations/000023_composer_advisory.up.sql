-- Packagist security advisories for official shopware/* Composer packages,
-- with Shopmon-owned enrichment columns that sync never overwrites.
CREATE TABLE "composer_advisory" (
  "advisory_id" text PRIMARY KEY NOT NULL,
  "package_name" text NOT NULL,
  "title" text NOT NULL,
  "link" text,
  "cve" text,
  "ghsa_id" text,
  "affected_versions" text NOT NULL,
  "severity" text,
  "sources" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "reported_at" timestamp,
  "composer_repository" text,
  "synced_at" timestamp NOT NULL DEFAULT NOW(),
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW(),

  -- Admin enrichment (never written by Packagist sync)
  "severity_override" text,
  "is_visible" boolean NOT NULL DEFAULT true,
  "notes_public" text,
  "notes_internal" text,
  "remediation_summary" text,
  "remediation_url" text,
  "recommended_upgrade" text,
  "shopware_impact_summary" text,
  "affected_components" text[] NOT NULL DEFAULT '{}',
  "tags" text[] NOT NULL DEFAULT '{}',
  "enriched_at" timestamp,
  "enriched_by" text
);

CREATE INDEX idx_composer_advisory_package_reported
  ON composer_advisory (package_name, reported_at DESC NULLS LAST);
CREATE INDEX idx_composer_advisory_cve
  ON composer_advisory (cve) WHERE cve IS NOT NULL AND cve <> '';
CREATE INDEX idx_composer_advisory_ghsa
  ON composer_advisory (ghsa_id) WHERE ghsa_id IS NOT NULL AND ghsa_id <> '';
CREATE INDEX idx_composer_advisory_visible_reported
  ON composer_advisory (is_visible, reported_at DESC NULLS LAST);
CREATE INDEX idx_composer_advisory_tags
  ON composer_advisory USING GIN (tags);

CREATE TABLE "composer_advisory_sync_state" (
  "id" int PRIMARY KEY NOT NULL DEFAULT 1 CHECK (id = 1),
  "last_updated_since" bigint,
  "last_full_sync_at" timestamp,
  "last_incremental_sync_at" timestamp,
  "last_error" text,
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

INSERT INTO composer_advisory_sync_state (id) VALUES (1);
