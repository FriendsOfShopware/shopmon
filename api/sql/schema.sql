CREATE TABLE "user" (
  "id" text PRIMARY KEY NOT NULL,
  "name" text NOT NULL,
  "email" text NOT NULL UNIQUE,
  "email_verified" boolean NOT NULL DEFAULT false,
  "image" text,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "role" text NOT NULL DEFAULT 'user',
  "banned" boolean DEFAULT false,
  "ban_reason" text,
  "ban_expires" timestamp,
  "notifications" jsonb DEFAULT '[]'::jsonb,
  "locale" text NOT NULL DEFAULT 'en'
);

CREATE TABLE "account" (
  "id" text PRIMARY KEY NOT NULL,
  "account_id" text NOT NULL,
  "provider_id" text NOT NULL,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "access_token" text,
  "refresh_token" text,
  "id_token" text,
  "access_token_expires_at" timestamp,
  "refresh_token_expires_at" timestamp,
  "scope" text,
  "password" text,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL
);

CREATE TABLE "session" (
  "id" text PRIMARY KEY NOT NULL,
  "expires_at" timestamp NOT NULL,
  "token" text NOT NULL UNIQUE,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL,
  "ip_address" text,
  "user_agent" text,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "impersonated_by" text,
  "active_organization_id" text
);

CREATE TABLE "passkey" (
  "id" text PRIMARY KEY NOT NULL,
  "name" text,
  "public_key" text NOT NULL,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "credential_id" text NOT NULL,
  "counter" integer NOT NULL,
  "device_type" text NOT NULL,
  "backed_up" boolean NOT NULL,
  "transports" text,
  "created_at" timestamp,
  "aaguid" text
);

CREATE TABLE "verification" (
  "id" text PRIMARY KEY NOT NULL,
  "identifier" text NOT NULL,
  "value" text NOT NULL,
  "expires_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL
);

CREATE TABLE "organization" (
  "id" text PRIMARY KEY NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL UNIQUE,
  "logo" text,
  "created_at" timestamp NOT NULL
);

CREATE TABLE "member" (
  "id" text PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "role" text NOT NULL DEFAULT 'member',
  "created_at" timestamp NOT NULL,
  UNIQUE ("organization_id", "user_id")
);

CREATE TABLE "invitation" (
  "id" text PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "email" text NOT NULL,
  "role" text,
  "status" text NOT NULL DEFAULT 'pending',
  "expires_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,
  "inviter_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade
);

CREATE TABLE "sso_provider" (
  "id" text PRIMARY KEY NOT NULL,
  "issuer" text NOT NULL,
  "oidc_config" text,
  "provider_id" text NOT NULL UNIQUE,
  "organization_id" text REFERENCES "organization"("id") ON DELETE CASCADE,
  "domain" text NOT NULL
);

CREATE TABLE "shop" (
  "id" serial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "description" text,
  "git_url" text,
  "default_environment_id" integer,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL
);

CREATE TABLE "shop_api_key" (
  "id" text PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "token" text NOT NULL UNIQUE,
  "scopes" jsonb NOT NULL,
  "created_at" timestamp NOT NULL,
  "last_used_at" timestamp
);

-- deployment is created before environment to resolve the circular reference:
-- environment.active_deployment_id -> deployment and deployment.environment_id -> environment.
-- We create deployment first with environment_id as a plain integer (no FK),
-- then create environment with its FK to deployment, then ALTER deployment to add the FK to environment.

CREATE TABLE "deployment" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL,
  "name" text NOT NULL,
  "command" text NOT NULL,
  "return_code" integer NOT NULL,
  "start_date" timestamp NOT NULL,
  "end_date" timestamp NOT NULL,
  "execution_time" real NOT NULL,
  "composer" jsonb DEFAULT '{}'::jsonb,
  "reference" text,
  "created_at" timestamp NOT NULL
);

CREATE TABLE "environment" (
  "id" serial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id"),
  "shop_id" integer NOT NULL REFERENCES "shop"("id"),
  "name" text NOT NULL,
  "status" text NOT NULL DEFAULT 'green',
  "url" text NOT NULL,
  "favicon" text,
  "client_id" text NOT NULL,
  "client_secret" text NOT NULL,
  "shopware_version" text NOT NULL,
  "last_scraped_at" timestamp,
  "last_scraped_error" text,
  "ignores" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "environment_image" text,
  "last_changelog" jsonb DEFAULT '{}'::jsonb,
  "active_deployment_id" integer REFERENCES "deployment"("id") ON DELETE set null,
  "connection_issue_count" integer NOT NULL DEFAULT 0,
  "sitespeed_enabled" boolean NOT NULL DEFAULT false,
  "sitespeed_urls" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "environment_token" text NOT NULL,
  "created_at" timestamp NOT NULL
);

-- Now add the FK from deployment.environment_id -> environment.id
ALTER TABLE "deployment" ADD CONSTRAINT fk_deployment_environment
  FOREIGN KEY ("environment_id") REFERENCES "environment"("id") ON DELETE cascade;

-- shop.default_environment_id references environment with ON DELETE RESTRICT, so an
-- environment that is still a shop's default cannot be deleted. The application moves
-- the default to another environment of the shop (or NULL) before deleting. Defined
-- here via ALTER because shop is created before environment to resolve the cycle.
ALTER TABLE "shop" ADD CONSTRAINT shop_default_environment_id_fkey
  FOREIGN KEY ("default_environment_id") REFERENCES "environment"("id") ON DELETE RESTRICT;

CREATE TABLE "environment_sitespeed" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer REFERENCES "environment"("id") ON DELETE cascade,
  "deployment_id" integer REFERENCES "deployment"("id") ON DELETE set null,
  "created_at" timestamp NOT NULL,
  "ttfb" integer,
  "fully_loaded" integer,
  "largest_contentful_paint" integer,
  "first_contentful_paint" integer,
  "cumulative_layout_shift" real,
  "transfer_size" integer
);

CREATE TABLE "environment_changelog" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer REFERENCES "environment"("id") ON DELETE cascade,
  "extensions" jsonb NOT NULL,
  "old_shopware_version" text,
  "new_shopware_version" text,
  "date" timestamp NOT NULL
);

CREATE TABLE "environment_cache" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL UNIQUE REFERENCES "environment"("id") ON DELETE cascade,
  "environment" text NOT NULL,
  "http_cache" boolean NOT NULL,
  "cache_adapter" text NOT NULL
);

CREATE TABLE "environment_check" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "check_id" text NOT NULL,
  "level" text NOT NULL,
  "message" text NOT NULL,
  "message_key" text,
  "params" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "source" text NOT NULL,
  "link" text,
  UNIQUE ("environment_id", "check_id")
);

-- environment_extension holds only extensions that are NOT known to the Shopware
-- store (api.shopware.com). Store-known extensions live in the normalized
-- store_extension* tables and are linked per environment via
-- environment_store_extension.
CREATE TABLE "environment_extension" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "label" text NOT NULL,
  "active" boolean NOT NULL,
  "version" text NOT NULL,
  "latest_version" text,
  "installed" boolean NOT NULL,
  "installed_at" text,
  UNIQUE ("environment_id", "name")
);

-- store_extension is the deduplicated catalog of extensions available on the
-- Shopware store, one row per technical name. The compatibility-capped "latest
-- version" is environment-specific and lives on environment_store_extension, not here.
CREATE TABLE "store_extension" (
  "name" text PRIMARY KEY NOT NULL,
  "store_id" integer,
  "icon_url" text,
  "producer_name" text,
  "producer_website" text,
  "rating_average" integer,
  "store_link" text,
  "release_date" text,
  "latest_version" text,
  "last_refreshed_at" timestamp NOT NULL DEFAULT NOW()
);

-- store_extension_translation holds the per-language store metadata for a store
-- extension. The API joins on the requested language and falls back to English.
CREATE TABLE "store_extension_translation" (
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "language" text NOT NULL,
  "label" text,
  "short_description" text,
  "description" text,
  "installation_manual" text,
  PRIMARY KEY ("extension_name", "language")
);

-- store_extension_version is the per-version changelog catalog for a store
-- extension; the changelog text itself is stored per language in
-- store_extension_version_translation.
CREATE TABLE "store_extension_version" (
  "id" serial PRIMARY KEY NOT NULL,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "version" text NOT NULL,
  "released_at" text,
  UNIQUE ("extension_name", "version")
);

CREATE TABLE "store_extension_version_translation" (
  "extension_version_id" integer NOT NULL REFERENCES "store_extension_version"("id") ON DELETE cascade,
  "language" text NOT NULL,
  "changelog" text,
  PRIMARY KEY ("extension_version_id", "language")
);

-- store_extension_sync records when the store API was last asked about an
-- extension name, including names the store does not know (so custom plugins
-- are not re-checked on every scrape). It gates the shared catalog refresh.
CREATE TABLE "store_extension_sync" (
  "extension_name" text PRIMARY KEY NOT NULL,
  "last_synced_at" timestamp NOT NULL DEFAULT NOW()
);

-- store_extension_compatibility stores, per Shopware version in use by any
-- environment, the latest extension release the store reports as compatible.
-- Environment scrapes read their compatible latest version from here instead
-- of asking the store API per environment.
CREATE TABLE "store_extension_compatibility" (
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "shopware_version" text NOT NULL,
  "latest_version" text,
  PRIMARY KEY ("extension_name", "shopware_version")
);

-- Lets CleanupUnusedStoreExtensionCompatibility, which filters by
-- shopware_version alone, seek instead of scanning (the PK leads with extension_name).
CREATE INDEX "store_extension_compatibility_shopware_version_idx"
  ON "store_extension_compatibility" ("shopware_version");

-- store_extension_image holds the store listing pictures (screenshots) for a
-- store extension, used to build a richer extension listing in the UI.
CREATE TABLE "store_extension_image" (
  "id" serial PRIMARY KEY NOT NULL,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "url" text NOT NULL,
  "preview" boolean NOT NULL DEFAULT false,
  "priority" integer NOT NULL DEFAULT 0,
  UNIQUE ("extension_name", "url")
);

-- environment_store_extension links an environment to a store_extension and
-- records the per-environment install state. latest_version is the latest
-- release the store reports as compatible with this environment's Shopware
-- version, so it is stored here rather than on the shared catalog row.
CREATE TABLE "environment_store_extension" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "label" text NOT NULL,
  "version" text NOT NULL,
  "latest_version" text,
  "active" boolean NOT NULL,
  "installed" boolean NOT NULL,
  "installed_at" text,
  UNIQUE ("environment_id", "extension_name")
);

CREATE INDEX IF NOT EXISTS idx_store_extension_version_name ON store_extension_version (extension_name);
CREATE INDEX IF NOT EXISTS idx_store_extension_image_name ON store_extension_image (extension_name);
CREATE INDEX IF NOT EXISTS idx_environment_store_extension_env ON environment_store_extension (environment_id);
CREATE INDEX IF NOT EXISTS idx_environment_store_extension_name ON environment_store_extension (extension_name);

CREATE TABLE "environment_queue" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "size" integer NOT NULL,
  UNIQUE ("environment_id", "name")
);

CREATE TABLE "environment_scheduled_task" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "task_id" text NOT NULL,
  "name" text NOT NULL,
  "status" text NOT NULL,
  "interval" integer NOT NULL,
  "overdue" boolean NOT NULL,
  "last_execution_time" text,
  "next_execution_time" text,
  UNIQUE ("environment_id", "task_id")
);

CREATE TABLE "user_notification" (
  "id" serial PRIMARY KEY NOT NULL,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "key" text NOT NULL,
  "level" text NOT NULL,
  "title" text NOT NULL,
  "message" text NOT NULL,
  "title_key" text,
  "message_key" text,
  "params" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "link" jsonb NOT NULL,
  "read" boolean NOT NULL DEFAULT false,
  "created_at" timestamp NOT NULL,
  UNIQUE ("user_id", "key")
);

CREATE TABLE "environment_status_event" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "old_status" text NOT NULL,
  "new_status" text NOT NULL,
  "reasons" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX "idx_environment_status_event_env" ON "environment_status_event" ("environment_id", "created_at" DESC);

CREATE TABLE "notification_preference" (
  "id" serial PRIMARY KEY NOT NULL,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "scope_type" text NOT NULL,
  "scope_id" text NOT NULL DEFAULT '',
  "event_type" text NOT NULL DEFAULT '',
  "channel" text NOT NULL DEFAULT '',
  "enabled" boolean NOT NULL DEFAULT true,
  "config" jsonb NOT NULL DEFAULT '{}'::jsonb,
  UNIQUE ("user_id", "scope_type", "scope_id", "event_type", "channel")
);

CREATE INDEX "idx_notification_preference_user" ON "notification_preference" ("user_id");
CREATE INDEX "idx_notification_preference_scope" ON "notification_preference" ("scope_type", "scope_id");

CREATE TABLE "lock" (
  "key" text PRIMARY KEY NOT NULL,
  "expires" timestamp NOT NULL,
  "created_at" timestamp NOT NULL
);

CREATE TABLE "audit_log" (
  "id" bigserial PRIMARY KEY,
  "actor_user_id" text REFERENCES "user"("id") ON DELETE SET NULL,
  "action" text NOT NULL,
  "target_user_id" text REFERENCES "user"("id") ON DELETE SET NULL,
  "detail" text,
  "ip_address" text,
  "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_passkey_credential_id ON passkey(credential_id);
CREATE INDEX IF NOT EXISTS idx_verification_value ON verification(value);
CREATE INDEX IF NOT EXISTS idx_sso_provider_domain ON sso_provider(domain);
CREATE INDEX IF NOT EXISTS idx_account_provider_user ON account(provider_id, user_id);
CREATE INDEX IF NOT EXISTS idx_account_provider_account ON account(provider_id, account_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at);

-- shopware_version is a local cache of Shopware releases crawled hourly from
-- https://releases.shopware.com/changelog. It lets the API answer version
-- queries without calling any external service.
CREATE TABLE "shopware_version" (
  "version" text PRIMARY KEY NOT NULL,
  "release_date" timestamp NOT NULL,
  "title" text NOT NULL DEFAULT '',
  "body" text NOT NULL DEFAULT '',
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shopware_version_release_date ON shopware_version (release_date);

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

CREATE INDEX idx_environment_sbom_component_package
  ON environment_sbom_component (package_name);

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
-- outside Shopmon. Separate from environment.ignores, which cannot carry a
-- reason, actor, expiry, or shop-wide scope; both coexist and are unioned at
-- the point of use. Revoking is a soft delete so the row remains an audit trail.
CREATE TABLE "advisory_suppression" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  -- NULL means every environment of the shop; set narrows to one environment.
  "environment_id" integer REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "reason" text NOT NULL,
  "expires_at" timestamp,
  "created_by" text REFERENCES "user"("id") ON DELETE SET NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "revoked_at" timestamp,
  "revoked_by" text REFERENCES "user"("id") ON DELETE SET NULL
);

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
-- Serves the ON DELETE CASCADE from environment.
CREATE INDEX idx_advisory_suppression_environment
  ON advisory_suppression (environment_id) WHERE environment_id IS NOT NULL;

-- Ties environment_id to shop_id so a row can never point at an environment of
-- a different shop (MATCH SIMPLE skips shop-wide rows, whose environment_id is
-- NULL). Requires the UNIQUE (shop_id, id) on environment.
ALTER TABLE environment
  ADD CONSTRAINT environment_shop_id_id_key UNIQUE (shop_id, id);
ALTER TABLE advisory_suppression
  ADD CONSTRAINT advisory_suppression_environment_shop_fkey
  FOREIGN KEY (shop_id, environment_id)
  REFERENCES environment (shop_id, id) ON DELETE CASCADE;

-- Which SwagPlatformSecurity release backports which advisory. Derived from the
-- store changelog Shopmon already syncs; plugin_version is the LOWEST version
-- on the branch naming the GHSA, since later entries only refine an existing fix.
CREATE TABLE "security_plugin_fix" (
  "ghsa_id" text NOT NULL,
  "plugin_branch" text NOT NULL,
  "plugin_version" text NOT NULL,
  "shopware_branch" text NOT NULL DEFAULT '',
  "released_at" timestamp,
  "synced_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("ghsa_id", "plugin_branch")
);

CREATE INDEX idx_security_plugin_fix_branch ON security_plugin_fix ("plugin_branch");

-- Parse bookkeeping, so "no fix row" can be told apart from "never parsed".
CREATE TABLE "security_plugin_coverage" (
  "plugin_branch" text PRIMARY KEY NOT NULL,
  "min_parsed_version" text NOT NULL DEFAULT '',
  "max_parsed_version" text NOT NULL DEFAULT '',
  "ghsa_count" integer NOT NULL DEFAULT 0,
  "parsed_at" timestamp NOT NULL DEFAULT NOW(),
  "last_error" text
);
