-- Owner-visible record of why an environment's background refresh is failing
-- (fully or partially): missing integration privileges, rejected credentials,
-- unreachable shop, and similar. One row per (environment, issue); rows are
-- deleted when the issue resolves, so presence means "still broken".
CREATE TABLE "environment_scrape_issue" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "issue_key" text NOT NULL,
  "kind" text NOT NULL,
  "dataset" text NOT NULL,
  "message_key" text NOT NULL,
  "params" jsonb NOT NULL DEFAULT '{}'::jsonb,
  "consecutive_count" integer NOT NULL DEFAULT 1,
  "first_seen_at" timestamp NOT NULL DEFAULT NOW(),
  "last_seen_at" timestamp NOT NULL DEFAULT NOW(),
  "notified_at" timestamp,
  PRIMARY KEY ("environment_id", "issue_key")
);
