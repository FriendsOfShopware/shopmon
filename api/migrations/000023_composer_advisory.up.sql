-- Composer security advisory catalog, SBOM-based exposure matching, and the
-- suppression/notification bookkeeping around them.

-- Logical Packagist/NVD security advisory (one row per CVE/GHSA).
-- Package-specific PKSA ids and version ranges live in composer_advisory_package.
CREATE TABLE "composer_advisory" (
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

  -- Disclosure details from NVD / GitHub Advisory, filled after Packagist sync
  "summary" text,
  "description" text,
  "cvss_score" double precision,
  "cvss_vector" text,
  "cwes" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "external_references" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "details_source" text,
  "details_synced_at" timestamp,
  -- GitHub is the only source of first_patched_versions, so its pass is
  -- tracked separately from the shared details marker: an advisory enriched by
  -- NVD first must still be visited by the GitHub queue.
  "github_synced_at" timestamp,
  -- Last detail-fetch attempt, successful or not. The enrichment queues order
  -- by this so unresolvable identifiers drift to the back instead of starving
  -- the queue.
  "details_attempted_at" timestamp,

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
  "enriched_by" text,

  -- Per-branch first patched core version from GitHub, e.g.
  -- {"6.7":"6.7.10.1"}. Machine-derived; never written by admin enrichment.
  "first_patched_versions" jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE TABLE "composer_advisory_package" (
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE CASCADE,
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

CREATE TABLE "composer_advisory_sync_state" (
  "id" int PRIMARY KEY NOT NULL DEFAULT 1 CHECK (id = 1),
  "last_updated_since" bigint,
  "last_full_sync_at" timestamp,
  "last_incremental_sync_at" timestamp,
  "last_error" text,
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

INSERT INTO composer_advisory_sync_state (id) VALUES (1);

-- Per-environment Composer package inventory from the FroshTools CycloneDX SBOM
-- endpoint. Current state only: each scrape replaces the environment's set.
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

-- Materialized advisory x environment matches, recomputed on scrape and after
-- each advisory sync so new CVEs surface without waiting for the next scrape.
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

-- Which advisories an environment's subscribers were actually alerted about.
-- Separate from environment_advisory_match because that table is rebuilt on
-- every rematch: a marker there would be lost each pass, and a notification
-- that failed after the match rows committed could never be retried.
CREATE TABLE "environment_advisory_notified" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "notified_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("environment_id", "advisory_id")
);

-- Shop-owner acknowledgement that a known advisory is accepted or mitigated
-- outside Shopmon (WAF rule, unused component, compensating control).
--
-- Deliberately separate from environment.ignores: that column is a flat array
-- of check ids with no room for a reason, an actor, an expiry, or a shop-wide
-- scope, and it is rewritten wholesale by the generic environment PATCH. Both
-- mechanisms coexist; they are unioned at the point of use.
CREATE TABLE "advisory_suppression" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  -- NULL means every environment of the shop; set narrows to one environment
  -- (e.g. "accepted on staging, still alert on production").
  "environment_id" integer REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "reason" text NOT NULL,
  "expires_at" timestamp,
  "created_by" text REFERENCES "user"("id") ON DELETE SET NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  -- Revoking is a soft delete: created_by/created_at/reason/revoked_* are the
  -- audit trail, so rows are never removed.
  "revoked_at" timestamp,
  "revoked_by" text REFERENCES "user"("id") ON DELETE SET NULL
);

-- Two partial uniques rather than one: NULLs are never equal in a unique
-- index, so a single constraint would let duplicate shop-wide rows through.
CREATE UNIQUE INDEX idx_advisory_suppression_shop
  ON advisory_suppression (shop_id, advisory_id)
  WHERE revoked_at IS NULL AND environment_id IS NULL;
CREATE UNIQUE INDEX idx_advisory_suppression_env
  ON advisory_suppression (environment_id, advisory_id)
  WHERE revoked_at IS NULL AND environment_id IS NOT NULL;

CREATE INDEX idx_advisory_suppression_advisory
  ON advisory_suppression (advisory_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_advisory_suppression_lookup
  ON advisory_suppression (shop_id, environment_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_advisory_suppression_organization
  ON advisory_suppression (organization_id) WHERE revoked_at IS NULL;
-- Serves the ON DELETE CASCADE from environment; the lookup index above leads
-- with shop_id and cannot.
CREATE INDEX idx_advisory_suppression_environment
  ON advisory_suppression (environment_id) WHERE environment_id IS NOT NULL;

-- Ties environment_id to shop_id so a row can never point at an environment of
-- a different shop: such a row is invisible to every read query (they join on
-- s.shop_id = e.shop_id AND s.environment_id = e.id) yet still occupies the
-- partial unique index, blocking the correct suppression. MATCH SIMPLE skips
-- rows whose environment_id is NULL (shop-wide scope).
ALTER TABLE environment
  ADD CONSTRAINT environment_shop_id_id_key UNIQUE (shop_id, id);
ALTER TABLE advisory_suppression
  ADD CONSTRAINT advisory_suppression_environment_shop_fkey
  FOREIGN KEY (shop_id, environment_id)
  REFERENCES environment (shop_id, id) ON DELETE CASCADE;

-- Which SwagPlatformSecurity release backports which advisory.
--
-- The Shopware Security Plugin closes core vulnerabilities without a core
-- upgrade, and its docs state that "every fix in the plugin corresponds to a
-- published security advisory and is identified by its GHSA id". Those ids
-- appear in the store changelog, which Shopmon already syncs, so this table is
-- derived from data we hold rather than a new integration.
CREATE TABLE "security_plugin_fix" (
  "ghsa_id" text NOT NULL,
  -- Plugin major version. The store returns changelogs for every branch
  -- regardless of the Shopware version queried, so the branch must come from
  -- the plugin version itself, never from the probe parameter.
  "plugin_branch" text NOT NULL,
  -- The LOWEST version on the branch whose changelog names this GHSA. Later
  -- entries refine an existing fix ("Improved the SVG validator for
  -- GHSA-xvhc-gm7j-mhmc" in 4.0.12), so the minimum is the first fix and a
  -- maximum would overstate what a shop needs.
  "plugin_version" text NOT NULL,
  -- Shopware line the branch serves ("6.7"), or '' when the mapping is unknown.
  "shopware_branch" text NOT NULL DEFAULT '',
  "released_at" timestamp,
  "synced_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("ghsa_id", "plugin_branch")
);

CREATE INDEX idx_security_plugin_fix_branch ON security_plugin_fix ("plugin_branch");

-- Parse bookkeeping per branch. Without it, "no fix row" is ambiguous between
-- "the plugin does not backport this" and "we have never parsed this branch" --
-- and the checker must not treat those the same, because only the first is a
-- statement about a shop's exposure.
CREATE TABLE "security_plugin_coverage" (
  "plugin_branch" text PRIMARY KEY NOT NULL,
  "min_parsed_version" text NOT NULL DEFAULT '',
  "max_parsed_version" text NOT NULL DEFAULT '',
  "ghsa_count" integer NOT NULL DEFAULT 0,
  "parsed_at" timestamp NOT NULL DEFAULT NOW(),
  "last_error" text
);
