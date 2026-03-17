CREATE TABLE "shop_uptime" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"checked_at" timestamp NOT NULL,
	"status_code" integer,
	"ttfb" integer,
	"response_time" integer,
	"is_up" boolean NOT NULL,
	"error" text
);
--> statement-breakpoint
CREATE TABLE "shop_uptime_daily" (
	"id" serial PRIMARY KEY NOT NULL,
	"shop_id" integer NOT NULL,
	"date" timestamp NOT NULL,
	"total_checks" integer NOT NULL,
	"up_checks" integer NOT NULL,
	"avg_ttfb" integer,
	"min_ttfb" integer,
	"max_ttfb" integer,
	"avg_response_time" integer,
	CONSTRAINT "shop_uptime_daily_shop_date_unique" UNIQUE("shop_id","date")
);
--> statement-breakpoint
ALTER TABLE "shop" ADD COLUMN "uptime_enabled" boolean DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE "shop" ADD COLUMN "uptime_status" text DEFAULT 'unknown' NOT NULL;--> statement-breakpoint
ALTER TABLE "shop" ADD COLUMN "uptime_last_checked_at" timestamp;--> statement-breakpoint
ALTER TABLE "shop" ADD COLUMN "uptime_down_since" timestamp;--> statement-breakpoint
ALTER TABLE "shop" ADD COLUMN "uptime_consecutive_fails" integer DEFAULT 0 NOT NULL;--> statement-breakpoint
ALTER TABLE "shop_uptime" ADD CONSTRAINT "shop_uptime_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
ALTER TABLE "shop_uptime_daily" ADD CONSTRAINT "shop_uptime_daily_shop_id_shop_id_fk" FOREIGN KEY ("shop_id") REFERENCES "public"."shop"("id") ON DELETE cascade ON UPDATE no action;--> statement-breakpoint
CREATE INDEX "shop_uptime_shop_checked_idx" ON "shop_uptime" USING btree ("shop_id","checked_at");--> statement-breakpoint
CREATE INDEX "shop_uptime_daily_shop_date_idx" ON "shop_uptime_daily" USING btree ("shop_id","date");