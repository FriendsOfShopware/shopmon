-- Per-environment Composer package inventory, ingested from the FroshTools
-- CycloneDX SBOM endpoint (/api/_action/frosh-tools/security/sbom), plus the
-- materialized advisory matches derived from it.

-- Current-state only: every scrape replaces an environment's component set.
-- No history is kept; environment_changelog covers the "what changed" story.
CREATE TABLE "environment_sbom_component" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "package_name" text NOT NULL,
  "version" text NOT NULL,
  "package_type" text,
  "purl" text,
  "is_dev" boolean NOT NULL DEFAULT false,
  "synced_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("environment_id", "package_name")
);

-- Reverse lookup ("which environments ship this package") drives the affected
-- shops list on advisory detail, so package_name leads the index.
CREATE INDEX idx_environment_sbom_component_package
  ON environment_sbom_component (package_name);

-- One row per environment, tracking SBOM availability separately from the
-- component rows so we can distinguish "never fetched" from "fetched, empty"
-- and avoid re-probing shops whose FroshTools is too old to serve the endpoint.
CREATE TABLE "environment_sbom_state" (
  "environment_id" integer PRIMARY KEY NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "supported" boolean NOT NULL DEFAULT true,
  "component_count" integer NOT NULL DEFAULT 0,
  "spec_version" text,
  "serial_number" text,
  "generated_at" timestamp,
  "last_synced_at" timestamp,
  "last_error" text
);

-- Materialized advisory x environment matches. Recomputed both on scrape (new
-- package versions) and after an advisory sync (new CVEs against unchanged
-- shops), so a freshly published advisory surfaces without waiting for the
-- environment's next scrape.
CREATE TABLE "environment_advisory_match" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "package_name" text NOT NULL,
  "installed_version" text NOT NULL,
  "affected_versions" text NOT NULL,
  "matched_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("environment_id", "advisory_id", "package_name")
);

CREATE INDEX idx_environment_advisory_match_advisory
  ON environment_advisory_match (advisory_id);
CREATE INDEX idx_environment_advisory_match_environment
  ON environment_advisory_match (environment_id);
