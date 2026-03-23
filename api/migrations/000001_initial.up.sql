CREATE TABLE IF NOT EXISTS "user" (
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

CREATE TABLE IF NOT EXISTS "account" (
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

CREATE TABLE IF NOT EXISTS "session" (
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

CREATE TABLE IF NOT EXISTS "passkey" (
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

CREATE TABLE IF NOT EXISTS "verification" (
  "id" text PRIMARY KEY NOT NULL,
  "identifier" text NOT NULL,
  "value" text NOT NULL,
  "expires_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "organization" (
  "id" text PRIMARY KEY NOT NULL,
  "name" text NOT NULL,
  "slug" text NOT NULL UNIQUE,
  "logo" text,
  "created_at" timestamp NOT NULL,
  "metadata" text
);

CREATE TABLE IF NOT EXISTS "member" (
  "id" text PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "role" text NOT NULL DEFAULT 'member',
  "created_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "invitation" (
  "id" text PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "email" text NOT NULL,
  "role" text,
  "status" text NOT NULL DEFAULT 'pending',
  "expires_at" timestamp NOT NULL,
  "created_at" timestamp NOT NULL,
  "inviter_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade
);

CREATE TABLE IF NOT EXISTS "sso_provider" (
  "id" text PRIMARY KEY NOT NULL,
  "issuer" text NOT NULL,
  "oidc_config" text,
  "saml_config" text,
  "user_id" text REFERENCES "user"("id") ON DELETE cascade,
  "provider_id" text NOT NULL UNIQUE,
  "organization_id" text,
  "domain" text NOT NULL
);

CREATE TABLE IF NOT EXISTS "project" (
  "id" serial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "description" text,
  "git_url" text,
  "created_at" timestamp NOT NULL,
  "updated_at" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "project_api_key" (
  "id" text PRIMARY KEY NOT NULL,
  "project_id" integer NOT NULL REFERENCES "project"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "token" text NOT NULL UNIQUE,
  "scopes" jsonb NOT NULL,
  "created_at" timestamp NOT NULL,
  "last_used_at" timestamp
);

CREATE TABLE IF NOT EXISTS "deployment" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL,
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

CREATE TABLE IF NOT EXISTS "shop" (
  "id" serial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id"),
  "project_id" integer NOT NULL REFERENCES "project"("id"),
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
  "shop_image" text,
  "last_changelog" jsonb DEFAULT '{}'::jsonb,
  "active_deployment_id" integer REFERENCES "deployment"("id") ON DELETE set null,
  "connection_issue_count" integer NOT NULL DEFAULT 0,
  "sitespeed_enabled" boolean NOT NULL DEFAULT false,
  "sitespeed_urls" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "shop_token" text NOT NULL,
  "created_at" timestamp NOT NULL
);

DO $$ BEGIN
  ALTER TABLE "deployment" ADD CONSTRAINT fk_deployment_shop
    FOREIGN KEY ("shop_id") REFERENCES "shop"("id") ON DELETE cascade;
EXCEPTION WHEN duplicate_object THEN NULL;
END $$;

CREATE TABLE IF NOT EXISTS "shop_sitespeed" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer REFERENCES "shop"("id"),
  "deployment_id" integer REFERENCES "deployment"("id") ON DELETE set null,
  "created_at" timestamp NOT NULL,
  "ttfb" integer,
  "fully_loaded" integer,
  "largest_contentful_paint" integer,
  "first_contentful_paint" integer,
  "cumulative_layout_shift" real,
  "transfer_size" integer
);

CREATE TABLE IF NOT EXISTS "shop_changelog" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer REFERENCES "shop"("id"),
  "extensions" jsonb NOT NULL,
  "old_shopware_version" text,
  "new_shopware_version" text,
  "date" timestamp NOT NULL
);

CREATE TABLE IF NOT EXISTS "shop_cache" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL UNIQUE REFERENCES "shop"("id") ON DELETE cascade,
  "environment" text NOT NULL,
  "http_cache" boolean NOT NULL,
  "cache_adapter" text NOT NULL
);

CREATE TABLE IF NOT EXISTS "shop_check" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  "check_id" text NOT NULL,
  "level" text NOT NULL,
  "message" text NOT NULL,
  "source" text NOT NULL,
  "link" text,
  UNIQUE ("shop_id", "check_id")
);

CREATE TABLE IF NOT EXISTS "shop_extension" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
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
  UNIQUE ("shop_id", "name")
);

CREATE TABLE IF NOT EXISTS "shop_queue" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  "name" text NOT NULL,
  "size" integer NOT NULL,
  UNIQUE ("shop_id", "name")
);

CREATE TABLE IF NOT EXISTS "shop_scheduled_task" (
  "id" serial PRIMARY KEY NOT NULL,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  "task_id" text NOT NULL,
  "name" text NOT NULL,
  "status" text NOT NULL,
  "interval" integer NOT NULL,
  "overdue" boolean NOT NULL,
  "last_execution_time" text,
  "next_execution_time" text,
  UNIQUE ("shop_id", "task_id")
);

CREATE TABLE IF NOT EXISTS "user_notification" (
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

CREATE TABLE IF NOT EXISTS "lock" (
  "key" text PRIMARY KEY NOT NULL,
  "expires" timestamp NOT NULL,
  "created_at" timestamp NOT NULL
);
