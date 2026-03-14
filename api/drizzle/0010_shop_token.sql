ALTER TABLE "shop" ADD COLUMN "shop_token" text;--> statement-breakpoint
UPDATE "shop" SET "shop_token" = md5(random()::text || clock_timestamp()::text) || md5(random()::text || clock_timestamp()::text) WHERE "shop_token" IS NULL;--> statement-breakpoint
ALTER TABLE "shop" ALTER COLUMN "shop_token" SET NOT NULL;