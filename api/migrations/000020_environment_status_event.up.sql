-- Records every environment status transition with the checks that caused it,
-- giving users a timeline of *why* a shop's status changed.
CREATE TABLE "environment_status_event" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "old_status" text NOT NULL,
  "new_status" text NOT NULL,
  "reasons" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX "idx_environment_status_event_env" ON "environment_status_event" ("environment_id", "created_at" DESC);
