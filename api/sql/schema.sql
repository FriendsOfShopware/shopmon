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
  "notifications" jsonb DEFAULT '[]'::jsonb
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
  "user_id" text NOT NULL REFERENCES "user"("id"),
  "key" text NOT NULL,
  "level" text NOT NULL,
  "title" text NOT NULL,
  "message" text NOT NULL,
  "link" jsonb NOT NULL,
  "read" boolean NOT NULL DEFAULT false,
  "created_at" timestamp NOT NULL,
  UNIQUE ("user_id", "key")
);

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

-- security_advisory is the central, browsable catalog of security advisories,
-- imported from the Packagist Security Advisories API for the first-party
-- Shopware packages. `origin` distinguishes imported advisories from future
-- user-submitted plugin disclosures; `affected_versions` (a composer
-- constraint) enables matching against tracked environments later.
CREATE TABLE "security_advisory" (
  "advisory_id" text PRIMARY KEY NOT NULL,
  "origin" text NOT NULL DEFAULT 'packagist',
  "package_name" text NOT NULL,
  "title" text NOT NULL,
  "link" text,
  "cve" text,
  "affected_versions" text NOT NULL,
  "source_name" text,
  "source_remote_id" text,
  "severity" text,
  "reported_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_security_advisory_package ON security_advisory (package_name);
CREATE INDEX IF NOT EXISTS idx_security_advisory_reported_at ON security_advisory (reported_at DESC);
