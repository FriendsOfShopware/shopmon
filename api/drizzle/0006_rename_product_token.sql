ALTER TABLE "deployment_token" RENAME TO "product_token";--> statement-breakpoint
ALTER TABLE "product_token" ADD COLUMN "scope" text NOT NULL DEFAULT 'deployment';--> statement-breakpoint
ALTER TABLE "product_token" RENAME CONSTRAINT "deployment_token_token_unique" TO "product_token_token_unique";--> statement-breakpoint
ALTER TABLE "product_token" RENAME CONSTRAINT "deployment_token_shop_id_shop_id_fk" TO "product_token_shop_id_shop_id_fk";
