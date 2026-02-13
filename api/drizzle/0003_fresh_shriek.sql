CREATE TABLE "deployment" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"name" text NOT NULL,
	"command" text NOT NULL,
	"output" text NOT NULL,
	"return_code" integer NOT NULL,
	"start_date" timestamp NOT NULL,
	"end_date" timestamp NOT NULL,
	"execution_time" text NOT NULL,
	"composer" jsonb DEFAULT '{}'::jsonb,
	"reference" text,
	"created_at" timestamp NOT NULL
);
--> statement-breakpoint
CREATE TABLE "deployment_token" (
	"id" text PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"token" text NOT NULL,
	"name" text NOT NULL,
	"created_at" timestamp NOT NULL,
	"last_used_at" timestamp,
	CONSTRAINT "deployment_token_token_unique" UNIQUE("token")
);
--> statement-breakpoint
ALTER TABLE "deployment" ADD CONSTRAINT "deployment_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "deployment_token" ADD CONSTRAINT "deployment_token_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;