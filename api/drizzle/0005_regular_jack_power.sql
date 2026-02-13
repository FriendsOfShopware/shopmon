DELETE FROM "deployment";--> statement-breakpoint
ALTER TABLE "deployment" ALTER COLUMN "execution_time" SET DATA TYPE real USING execution_time::real;