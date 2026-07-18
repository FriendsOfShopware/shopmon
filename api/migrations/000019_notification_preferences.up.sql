-- Normalized notification preferences replace the user.notifications JSONB array.
--
-- Row semantics (empty-string sentinels keep the UNIQUE/ON CONFLICT working,
-- since NULLs would be treated as distinct):
--   channel = ''              -> subscription marker ("I watch this scope")
--   channel in (in_app|email) -> per-channel enable/disable at this scope
--   scope_id = ''             -> global scope
--   event_type = ''           -> applies to all event types (wildcard)
--
-- Resolution is most-specific-first: environment -> organization -> global, and
-- event-specific overrides the wildcard. config is an extensibility bucket for
-- future per-preference settings (thresholds, quiet hours, webhook URLs, ...).
CREATE TABLE "notification_preference" (
  "id" serial PRIMARY KEY NOT NULL,
  "user_id" text NOT NULL REFERENCES "user"("id") ON DELETE cascade,
  "scope_type" text NOT NULL,
  "scope_id" text NOT NULL DEFAULT '',
  "event_type" text NOT NULL DEFAULT '',
  "channel" text NOT NULL DEFAULT '',
  "enabled" boolean NOT NULL DEFAULT true,
  "config" jsonb NOT NULL DEFAULT '{}'::jsonb,
  UNIQUE ("user_id", "scope_type", "scope_id", "event_type", "channel")
);

CREATE INDEX "idx_notification_preference_user" ON "notification_preference" ("user_id");
CREATE INDEX "idx_notification_preference_scope" ON "notification_preference" ("scope_type", "scope_id");

-- Migrate existing watches: each "environment-{id}" entry becomes a subscription
-- marker. Channels default to enabled (no channel rows), preserving the prior
-- behaviour where a watcher received every default channel for an event.
INSERT INTO "notification_preference" (user_id, scope_type, scope_id, event_type, channel, enabled)
SELECT u.id, 'environment', substring(elem from 'environment-(.*)'), '', '', true
FROM "user" u, jsonb_array_elements_text(u.notifications) AS elem
WHERE elem LIKE 'environment-%'
ON CONFLICT DO NOTHING;
