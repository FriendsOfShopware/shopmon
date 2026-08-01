-- Flatten package rows back into one parent row per Packagist PKSA.
CREATE TABLE "composer_advisory_flat" (
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
  "summary" text,
  "description" text,
  "cvss_score" double precision,
  "cvss_vector" text,
  "cwes" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "external_references" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "details_source" text,
  "details_synced_at" timestamp,
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

INSERT INTO composer_advisory_flat (
  advisory_id, package_name, title, link, cve, ghsa_id, affected_versions,
  severity, sources, reported_at, composer_repository, synced_at, created_at, updated_at,
  summary, description, cvss_score, cvss_vector, cwes, external_references,
  details_source, details_synced_at,
  severity_override, is_visible, notes_public, notes_internal,
  remediation_summary, remediation_url, recommended_upgrade,
  shopware_impact_summary, affected_components, tags, enriched_at, enriched_by
)
SELECT
  p.packagist_advisory_id,
  p.package_name,
  a.title,
  a.link,
  a.cve,
  a.ghsa_id,
  p.affected_versions,
  a.severity,
  a.sources,
  a.reported_at,
  a.composer_repository,
  p.synced_at,
  a.created_at,
  a.updated_at,
  a.summary,
  a.description,
  a.cvss_score,
  a.cvss_vector,
  a.cwes,
  a.external_references,
  a.details_source,
  a.details_synced_at,
  a.severity_override,
  a.is_visible,
  a.notes_public,
  a.notes_internal,
  a.remediation_summary,
  a.remediation_url,
  a.recommended_upgrade,
  a.shopware_impact_summary,
  a.affected_components,
  a.tags,
  a.enriched_at,
  a.enriched_by
FROM composer_advisory_package p
JOIN composer_advisory a ON a.advisory_id = p.advisory_id;

DROP TABLE composer_advisory_package;
DROP TABLE composer_advisory;
ALTER TABLE composer_advisory_flat RENAME TO composer_advisory;

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
