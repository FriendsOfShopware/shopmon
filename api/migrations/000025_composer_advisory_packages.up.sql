-- Collapse Packagist PKSA rows that share a CVE/GHSA into one logical advisory
-- with one-to-many package rows (affected version ranges per package).

CREATE TABLE "composer_advisory_package" (
  "advisory_id" text NOT NULL,
  "package_name" text NOT NULL,
  "packagist_advisory_id" text NOT NULL,
  "affected_versions" text NOT NULL,
  "synced_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("advisory_id", "package_name")
);

CREATE UNIQUE INDEX idx_composer_advisory_package_packagist_id
  ON composer_advisory_package (packagist_advisory_id);
CREATE INDEX idx_composer_advisory_package_name
  ON composer_advisory_package (package_name);

-- Package rows keyed by canonical advisory id: CVE > GHSA > PKSA.
INSERT INTO composer_advisory_package (
  advisory_id, package_name, packagist_advisory_id, affected_versions, synced_at
)
SELECT
  COALESCE(NULLIF(upper(trim(cve)), ''), NULLIF(trim(ghsa_id), ''), advisory_id),
  package_name,
  advisory_id,
  affected_versions,
  synced_at
FROM composer_advisory
ON CONFLICT (advisory_id, package_name) DO UPDATE SET
  packagist_advisory_id = EXCLUDED.packagist_advisory_id,
  affected_versions = EXCLUDED.affected_versions,
  synced_at = GREATEST(composer_advisory_package.synced_at, EXCLUDED.synced_at);

-- Rebuild parent as one row per canonical id (prefer detailed/enriched survivors).
CREATE TABLE "composer_advisory_new" (
  "advisory_id" text PRIMARY KEY NOT NULL,
  "title" text NOT NULL,
  "link" text,
  "cve" text,
  "ghsa_id" text,
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

-- Disclosure details come from the survivor row (most detailed wins), but
-- admin enrichment is MERGED across every row collapsing into one canonical
-- advisory: when two enriched PKSA rows share a CVE, picking one survivor
-- would silently discard the other's notes, tags, and overrides.
WITH src AS (
  SELECT
    *,
    COALESCE(NULLIF(upper(trim(cve)), ''), NULLIF(trim(ghsa_id), ''), advisory_id) AS canonical_id
  FROM composer_advisory
),
survivor AS (
  SELECT DISTINCT ON (canonical_id) *
  FROM src
  ORDER BY
    canonical_id,
    (details_synced_at IS NOT NULL) DESC,
    details_synced_at DESC NULLS LAST,
    (enriched_at IS NOT NULL) DESC,
    enriched_at DESC NULLS LAST,
    (notes_public IS NOT NULL OR notes_internal IS NOT NULL) DESC,
    synced_at DESC
),
-- Scalar enrichment: the value from the most recently enriched row that has
-- one (array_remove drops the NULLs of unenriched rows, keeping order).
-- Visibility: hidden wins — if an admin hid any collapsing row, the merged
-- advisory stays hidden rather than resurfacing.
enrichment AS (
  SELECT
    canonical_id,
    (array_remove(array_agg(severity_override ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS severity_override,
    bool_and(is_visible) AS is_visible,
    (array_remove(array_agg(notes_public ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS notes_public,
    (array_remove(array_agg(notes_internal ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS notes_internal,
    (array_remove(array_agg(remediation_summary ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS remediation_summary,
    (array_remove(array_agg(remediation_url ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS remediation_url,
    (array_remove(array_agg(recommended_upgrade ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS recommended_upgrade,
    (array_remove(array_agg(shopware_impact_summary ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS shopware_impact_summary,
    MAX(enriched_at) AS enriched_at,
    (array_remove(array_agg(enriched_by ORDER BY enriched_at DESC NULLS LAST, synced_at DESC), NULL))[1] AS enriched_by
  FROM src
  GROUP BY canonical_id
),
tag_union AS (
  SELECT canonical_id, array_agg(DISTINCT tag) AS tags
  FROM src, unnest(tags) AS tag
  GROUP BY canonical_id
),
component_union AS (
  SELECT canonical_id, array_agg(DISTINCT comp) AS affected_components
  FROM src, unnest(affected_components) AS comp
  GROUP BY canonical_id
)
INSERT INTO composer_advisory_new (
  advisory_id, title, link, cve, ghsa_id, severity, sources, reported_at,
  composer_repository, synced_at, created_at, updated_at,
  summary, description, cvss_score, cvss_vector, cwes, external_references,
  details_source, details_synced_at,
  severity_override, is_visible, notes_public, notes_internal,
  remediation_summary, remediation_url, recommended_upgrade,
  shopware_impact_summary, affected_components, tags, enriched_at, enriched_by
)
SELECT
  s.canonical_id,
  s.title,
  s.link,
  CASE WHEN s.cve IS NOT NULL AND trim(s.cve) <> '' THEN upper(trim(s.cve)) ELSE NULL END,
  NULLIF(trim(s.ghsa_id), ''),
  s.severity,
  s.sources,
  s.reported_at,
  s.composer_repository,
  s.synced_at,
  s.created_at,
  s.updated_at,
  s.summary,
  s.description,
  s.cvss_score,
  s.cvss_vector,
  s.cwes,
  s.external_references,
  s.details_source,
  s.details_synced_at,
  e.severity_override,
  e.is_visible,
  e.notes_public,
  e.notes_internal,
  e.remediation_summary,
  e.remediation_url,
  e.recommended_upgrade,
  e.shopware_impact_summary,
  COALESCE(c.affected_components, '{}'::text[]),
  COALESCE(t.tags, '{}'::text[]),
  e.enriched_at,
  e.enriched_by
FROM survivor s
JOIN enrichment e USING (canonical_id)
LEFT JOIN tag_union t USING (canonical_id)
LEFT JOIN component_union c USING (canonical_id);

DROP TABLE composer_advisory;
ALTER TABLE composer_advisory_new RENAME TO composer_advisory;

ALTER TABLE composer_advisory_package
  ADD CONSTRAINT composer_advisory_package_advisory_id_fkey
  FOREIGN KEY (advisory_id) REFERENCES composer_advisory (advisory_id) ON DELETE CASCADE;

CREATE INDEX idx_composer_advisory_cve
  ON composer_advisory (cve) WHERE cve IS NOT NULL AND cve <> '';
CREATE INDEX idx_composer_advisory_ghsa
  ON composer_advisory (ghsa_id) WHERE ghsa_id IS NOT NULL AND ghsa_id <> '';
CREATE INDEX idx_composer_advisory_visible_reported
  ON composer_advisory (is_visible, reported_at DESC NULLS LAST);
CREATE INDEX idx_composer_advisory_tags
  ON composer_advisory USING GIN (tags);
CREATE INDEX idx_composer_advisory_reported
  ON composer_advisory (reported_at DESC NULLS LAST);
