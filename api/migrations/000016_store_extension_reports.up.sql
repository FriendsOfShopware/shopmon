-- store_extension_report lets users flag a store extension (e.g. for performance,
-- security or compatibility issues). Reports start in the 'pending' state and only
-- become community-visible once an admin moves them to 'approved'. Rejected reports
-- are kept for the audit trail. reviewed_by / reviewed_at record the moderation.
CREATE TABLE "store_extension_report" (
  "id" serial PRIMARY KEY NOT NULL,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "user_id" text REFERENCES "user"("id") ON DELETE SET NULL,
  "category" text NOT NULL,
  "comment" text NOT NULL,
  "status" text NOT NULL DEFAULT 'pending',
  "reviewed_by" text REFERENCES "user"("id") ON DELETE SET NULL,
  "reviewed_at" timestamp,
  "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_store_extension_report_name ON store_extension_report (extension_name);
CREATE INDEX IF NOT EXISTS idx_store_extension_report_status ON store_extension_report (status);
