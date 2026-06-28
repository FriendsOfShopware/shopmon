-- Checks move from pre-rendered English strings to translation keys plus
-- structured params, rendered per-viewer in the UI. The legacy message column
-- is retained as an English fallback.
ALTER TABLE "environment_check" ADD COLUMN "message_key" text;
ALTER TABLE "environment_check" ADD COLUMN "params" jsonb NOT NULL DEFAULT '{}'::jsonb;
