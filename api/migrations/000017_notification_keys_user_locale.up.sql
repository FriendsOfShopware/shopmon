-- User locale drives server-side rendering of localized emails.
ALTER TABLE "user" ADD COLUMN "locale" text NOT NULL DEFAULT 'en';

-- Notifications move from pre-rendered English strings to translation keys plus
-- structured params, rendered per-viewer in the UI and per-recipient in emails.
-- The legacy title/message columns are retained as a rendered fallback so older
-- clients keep working until they switch to key-based rendering.
ALTER TABLE "user_notification" ADD COLUMN "title_key" text;
ALTER TABLE "user_notification" ADD COLUMN "message_key" text;
ALTER TABLE "user_notification" ADD COLUMN "params" jsonb NOT NULL DEFAULT '{}'::jsonb;
