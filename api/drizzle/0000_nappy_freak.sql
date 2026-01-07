CREATE TABLE "account" (
	"id" text PRIMARY KEY NOT NULL,
	"account_id" text NOT NULL,
	"provider_id" text NOT NULL,
	"user_id" text NOT NULL,
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
--> statement-breakpoint
CREATE TABLE "invitation" (
	"id" text PRIMARY KEY NOT NULL,
	"organization_id" text NOT NULL,
	"email" text NOT NULL,
	"role" text,
	"status" text DEFAULT 'pending' NOT NULL,
	"expires_at" timestamp NOT NULL,
	"inviter_id" text NOT NULL
);
--> statement-breakpoint
CREATE TABLE "lock" (
	"key" text PRIMARY KEY NOT NULL,
	"expires" timestamp NOT NULL,
	"created_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "member" (
	"id" text PRIMARY KEY NOT NULL,
	"organization_id" text NOT NULL,
	"user_id" text NOT NULL,
	"role" text DEFAULT 'member' NOT NULL,
	"created_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "organization" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"slug" text,
	"logo" text,
	"created_at" timestamp NOT NULL,
	"metadata" text,
	CONSTRAINT "organization_slug_unique" UNIQUE("slug")
);
--> statement-breakpoint
CREATE TABLE "passkey" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text,
	"public_key" text NOT NULL,
	"user_id" text NOT NULL,
	"credential_id" text NOT NULL,
	"counter" integer NOT NULL,
	"device_type" text NOT NULL,
	"backed_up" boolean NOT NULL,
	"transports" text,
	"created_at" timestamp,
	"aaguid" text
);
--> statement-breakpoint
CREATE TABLE "project" (
	"id" serial PRIMARY KEY NOT NULL,
	"organization_id" text NOT NULL,
	"name" text NOT NULL,
	"description" text,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "project_api_key" (
	"id" text PRIMARY KEY NOT NULL,
	"project_id" integer NOT NULL,
	"name" text NOT NULL,
	"token" text NOT NULL,
	"scopes" jsonb NOT NULL,
	"created_at" timestamp NOT NULL,
	"last_used_at" timestamp,
	CONSTRAINT "project_api_key_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "session" (
	"id" text PRIMARY KEY NOT NULL,
	"expires_at" timestamp NOT NULL,
	"token" text NOT NULL,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp NOT NULL,
	"ip_address" text,
	"user_agent" text,
	"user_id" text NOT NULL,
	"impersonated_by" text,
	"active_organization_id" text,
	CONSTRAINT "session_token_unique" UNIQUE("token")
);
--> statement-breakpoint
CREATE TABLE "shop" (
	"id" serial PRIMARY KEY NOT NULL,
	"organization_id" text NOT NULL,
	"project_id" integer NOT NULL,
	"name" text NOT NULL,
	"status" text DEFAULT 'green' NOT NULL,
	"url" text NOT NULL,
	"favicon" text,
	"client_id" text NOT NULL,
	"client_secret" text NOT NULL,
	"shopware_version" text NOT NULL,
	"last_scraped_at" timestamp,
	"last_scraped_error" text,
	"ignores" jsonb DEFAULT '[]'::jsonb NOT NULL,
	"shop_image" text,
	"last_changelog" jsonb DEFAULT '{}'::jsonb,
	"connection_issue_count" integer DEFAULT 0 NOT NULL,
	"sitespeed_enabled" boolean DEFAULT false NOT NULL,
	"sitespeed_urls" jsonb DEFAULT '[]'::jsonb NOT NULL,
	"created_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "shop_cache" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"environment" text NOT NULL,
	"http_cache" boolean NOT NULL,
	"cache_adapter" text NOT NULL,
	CONSTRAINT "shop_cache_shop_id_unique" UNIQUE("shop_id")
);
--> statement-breakpoint
CREATE TABLE "shop_changelog" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer,
	"extensions" jsonb NOT NULL,
	"old_shopware_version" text,
	"new_shopware_version" text,
	"date" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "shop_check" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"check_id" text NOT NULL,
	"level" text NOT NULL,
	"message" text NOT NULL,
	"source" text NOT NULL,
	"link" text,
	CONSTRAINT "shop_check_shop_id_check_id_unique" UNIQUE("shop_id","check_id")
);
--> statement-breakpoint
CREATE TABLE "shop_extension" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
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
	CONSTRAINT "shop_extension_shop_id_name_unique" UNIQUE("shop_id","name")
);
--> statement-breakpoint
CREATE TABLE "shop_queue" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"name" text NOT NULL,
	"size" integer NOT NULL,
	CONSTRAINT "shop_queue_shop_id_name_unique" UNIQUE("shop_id","name")
);
--> statement-breakpoint
CREATE TABLE "shop_scheduled_task" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"task_id" text NOT NULL,
	"name" text NOT NULL,
	"status" text NOT NULL,
	"interval" integer NOT NULL,
	"overdue" boolean NOT NULL,
	"last_execution_time" text,
	"next_execution_time" text,
	CONSTRAINT "shop_scheduled_task_shop_id_task_id_unique" UNIQUE("shop_id","task_id")
);
--> statement-breakpoint
CREATE TABLE "shop_sitespeed" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer,
	"created_at" timestamp NOT NULL,
	"ttfb" integer,
	"fully_loaded" integer,
	"largest_contentful_paint" integer,
	"first_contentful_paint" integer,
	"cumulative_layout_shift" integer,
	"transfer_size" integer
);
--> statement-breakpoint
CREATE TABLE "sso_provider" (
	"id" text PRIMARY KEY NOT NULL,
	"issuer" text NOT NULL,
	"oidc_config" text,
	"saml_config" text,
	"user_id" text,
	"provider_id" text NOT NULL,
	"organization_id" text,
	"domain" text NOT NULL,
	CONSTRAINT "sso_provider_provider_id_unique" UNIQUE("provider_id")
);
--> statement-breakpoint
CREATE TABLE "user" (
	"id" text PRIMARY KEY NOT NULL,
	"name" text NOT NULL,
	"email" text NOT NULL,
	"email_verified" boolean NOT NULL,
	"image" text,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp NOT NULL,
	"role" text DEFAULT 'user' NOT NULL,
	"banned" boolean DEFAULT false,
	"ban_reason" text,
	"ban_expires" timestamp,
	"notifications" jsonb DEFAULT '[]'::jsonb,
	CONSTRAINT "user_email_unique" UNIQUE("email")
);
--> statement-breakpoint
CREATE TABLE "user_notification" (
	"id" serial PRIMARY KEY NOT NULL,
	"user_id" text NOT NULL,
	"key" text NOT NULL,
	"level" text NOT NULL,
	"title" text NOT NULL,
	"message" text NOT NULL,
	"link" jsonb NOT NULL,
	"read" boolean DEFAULT false NOT NULL,
	"created_at" timestamp NOT NULL,
	CONSTRAINT "user_notification_user_id_key_unique" UNIQUE("user_id","key")
);
--> statement-breakpoint
CREATE TABLE "verification" (
	"id" text PRIMARY KEY NOT NULL,
	"identifier" text NOT NULL,
	"value" text NOT NULL,
	"expires_at" timestamp NOT NULL,
	"created_at" timestamp NOT NULL,
	"updated_at" timestamp NOT NULL
);
--> statement-breakpoint
ALTER TABLE "account" ADD CONSTRAINT "account_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "invitation" ADD CONSTRAINT "invitation_organization_id_organization_id_fk" FOREIGN KEY ("organization_id") REFERENCES "public"."organization"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "invitation" ADD CONSTRAINT "invitation_inviter_id_user_id_fk" FOREIGN KEY ("inviter_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member" ADD CONSTRAINT "member_organization_id_organization_id_fk" FOREIGN KEY ("organization_id") REFERENCES "public"."organization"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "member" ADD CONSTRAINT "member_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "passkey" ADD CONSTRAINT "passkey_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "project" ADD CONSTRAINT "project_organization_id_organization_id_fk" FOREIGN KEY ("organization_id") REFERENCES "public"."organization"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "project_api_key" ADD CONSTRAINT "project_api_key_project_id_project_id_fk" FOREIGN KEY ("project_id") REFERENCES "public"."project"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "session" ADD CONSTRAINT "session_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop" ADD CONSTRAINT "shop_organization_id_organization_id_fk" FOREIGN KEY ("organization_id") REFERENCES "public"."organization"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop" ADD CONSTRAINT "shop_project_id_project_id_fk" FOREIGN KEY ("project_id") REFERENCES "public"."project"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_cache" ADD CONSTRAINT "shop_cache_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_changelog" ADD CONSTRAINT "shop_changelog_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_check" ADD CONSTRAINT "shop_check_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_extension" ADD CONSTRAINT "shop_extension_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_queue" ADD CONSTRAINT "shop_queue_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_scheduled_task" ADD CONSTRAINT "shop_scheduled_task_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_sitespeed" ADD CONSTRAINT "shop_sitespeed_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE no action ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "sso_provider" ADD CONSTRAINT "sso_provider_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "user_notification" ADD CONSTRAINT "user_notification_user_id_user_id_fk" FOREIGN KEY ("user_id") REFERENCES "public"."user"("id") ON DELETE no action ON UPDATE no action;