-- Account monitoring: detect security-relevant changes to a shop's admin users,
-- integrations and ACL roles.
--
-- Shopmon deliberately stores no readable directory of who has access. What it
-- keeps is a baseline to diff against, so it can report *that* something
-- changed, rate how dangerous the change is, and let an operator acknowledge
-- it. An attacker who adds a second admin account survives the original hole
-- being closed; this is the signal that catches that.

-- Opt-out per environment. Default on: an unnoticed new admin is exactly the
-- risk this feature exists for.
ALTER TABLE "environment"
  ADD COLUMN "account_monitoring_enabled" boolean NOT NULL DEFAULT true;

-- Baseline for all three account kinds in one table. They differ only in which
-- columns are populated -- the same shape environment_advisory_match uses for
-- php, mysql and composer packages.
CREATE TABLE "environment_account" (
  "environment_id"            integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "account_type"              text NOT NULL,   -- user | integration | acl_role
  "account_id"                text NOT NULL,
  -- The masked email FroshTools already returns, or the integration/role name.
  -- Without any handle nobody can react, and after an incident the account may
  -- already be gone from the shop -- a bare UUID would then say nothing at all.
  "label"                     text NOT NULL,
  "active"                    boolean NOT NULL DEFAULT true,
  "admin"                     boolean NOT NULL DEFAULT false,
  "role_ids"                  jsonb   NOT NULL DEFAULT '[]'::jsonb,
  "roles_hash"                text    NOT NULL DEFAULT '',
  "access_key_count"          integer NOT NULL DEFAULT 0,
  "privilege_count"           integer NOT NULL DEFAULT 0,
  "privileges_hash"           text    NOT NULL DEFAULT '',
  "managed_by_app"            boolean NOT NULL DEFAULT false,
  "app_name"                  text,
  "password_recovery_pending" boolean NOT NULL DEFAULT false,
  "last_usage_at"             timestamp,
  "password_changed_at"       timestamp,
  "remote_created_at"         timestamp,
  "remote_updated_at"         timestamp,
  -- Shop-relative admin deep link, validated on ingest. Stored relative so a
  -- changed shop URL does not leave dead links behind; the absolute URL is
  -- composed when a finding is rendered.
  "admin_path"                text,
  "first_seen_at"             timestamp NOT NULL DEFAULT NOW(),
  "synced_at"                 timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("environment_id", "account_type", "account_id")
);

-- Capability and first-run marker, mirroring environment_sbom_state.
CREATE TABLE "environment_account_state" (
  "environment_id" integer PRIMARY KEY NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "supported"      boolean NOT NULL DEFAULT true,
  -- NULL = never initialised. The first run only seeds; without a prior state
  -- every account would read as "new" and every existing shop would turn red.
  "baseline_at"    timestamp,
  "account_count"  integer NOT NULL DEFAULT 0,
  "generated_at"   timestamp,
  "last_synced_at" timestamp,
  "last_error"     text
);

-- Append-only history. Findings are the rows with level <> 'info'; they stay
-- open -- and keep escalating the environment status -- until acknowledged.
-- No UNIQUE constraint needed: the diff runs against the baseline, which is
-- advanced in the same transaction, so every change is detected exactly once.
CREATE TABLE "environment_account_event" (
  "id"               bigserial PRIMARY KEY NOT NULL,
  "environment_id"   integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "account_type"     text NOT NULL,
  "account_id"       text NOT NULL,
  "label"            text NOT NULL,
  "change"           text NOT NULL,
  "level"            text NOT NULL,              -- red | yellow | info
  "params"           jsonb NOT NULL DEFAULT '{}'::jsonb,
  "admin_path"       text,
  "detected_at"      timestamp NOT NULL DEFAULT NOW(),
  -- Only set once the dispatch succeeded for every recipient; otherwise the
  -- next scrape retries the notification.
  "notified_at"      timestamp,
  "acknowledged_at"  timestamp,
  "acknowledged_by"  text REFERENCES "user"("id") ON DELETE set null,
  "acknowledge_note" text
);

CREATE INDEX "idx_environment_account_event_env"
  ON "environment_account_event" ("environment_id", "detected_at" DESC);
CREATE INDEX "idx_environment_account_event_open"
  ON "environment_account_event" ("environment_id")
  WHERE "acknowledged_at" IS NULL AND "level" <> 'info';
CREATE INDEX "idx_environment_account_event_unnotified"
  ON "environment_account_event" ("environment_id")
  WHERE "notified_at" IS NULL AND "level" <> 'info';
