-- Create the store_extension table for global extension metadata
CREATE TABLE IF NOT EXISTS "store_extension" (
	"id" serial PRIMARY KEY NOT NULL,
	"name" text NOT NULL UNIQUE,
	"rating_average" real,
	"store_link" text,
	"updated_at" timestamp NOT NULL DEFAULT now()
);--> statement-breakpoint

-- Create the store_extension_version table for changelog entries
CREATE TABLE IF NOT EXISTS "store_extension_version" (
	"id" serial PRIMARY KEY NOT NULL,
	"store_extension_id" integer NOT NULL REFERENCES "store_extension"("id") ON DELETE CASCADE,
	"version" text NOT NULL,
	"changelog" text,
	"release_date" text,
	CONSTRAINT "store_extension_version_store_extension_id_version_unique" UNIQUE("store_extension_id","version")
);--> statement-breakpoint

-- Migrate existing store data from shop_extension into store_extension
INSERT INTO "store_extension" ("name", "rating_average", "store_link", "updated_at")
SELECT DISTINCT ON (se."name") se."name", se."rating_average", se."store_link", now()
FROM "shop_extension" se
WHERE se."store_link" IS NOT NULL
ON CONFLICT ("name") DO NOTHING;--> statement-breakpoint

-- Add store_extension_id column to shop_extension
ALTER TABLE "shop_extension" ADD COLUMN "store_extension_id" integer REFERENCES "store_extension"("id") ON DELETE SET NULL;--> statement-breakpoint

-- Backfill the store_extension_id references
UPDATE "shop_extension" se
SET "store_extension_id" = ste."id"
FROM "store_extension" ste
WHERE se."name" = ste."name";--> statement-breakpoint

-- Drop the columns that are now in store_extension
ALTER TABLE "shop_extension" DROP COLUMN IF EXISTS "rating_average";--> statement-breakpoint
ALTER TABLE "shop_extension" DROP COLUMN IF EXISTS "store_link";--> statement-breakpoint
ALTER TABLE "shop_extension" DROP COLUMN IF EXISTS "changelog";
