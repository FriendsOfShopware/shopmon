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
  "created_at" timestamp NOT NULL,
  "metadata" text
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
  "saml_config" text,
  "user_id" text REFERENCES "user"("id") ON DELETE cascade,
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

CREATE TABLE "environment_sitespeed" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer REFERENCES "environment"("id"),
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
  "environment_id" integer REFERENCES "environment"("id"),
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

CREATE TABLE "environment_extension" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "label" text NOT NULL,
  "active" boolean NOT NULL,
  "version" text NOT NULL,
  "latest_version" text,
  "installed" boolean NOT NULL,
  "rating_average" integer,
  "store_link" text,
  "changelog" jsonb,
  "installed_at" text,
  UNIQUE ("environment_id", "name")
);

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
