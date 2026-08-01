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

-- Per-branch first patched core version from the GitHub Advisory API, e.g.
-- {"6.7": "6.7.10.1", "6.6": "6.6.10.18"}. Machine-derived and written only by
-- the advisory sync, so it never collides with the admin-owned
-- recommended_upgrade / remediation_* columns.
ALTER TABLE "composer_advisory"
  ADD COLUMN "first_patched_versions" jsonb NOT NULL DEFAULT '{}'::jsonb;
