-- user_notification was the only user-referencing table without ON DELETE
-- cascade, so deleting a user who had any notification failed on the foreign
-- key. Align it with account/session/member/notification_preference.
ALTER TABLE "user_notification" DROP CONSTRAINT IF EXISTS "user_notification_user_id_fkey";
ALTER TABLE "user_notification"
  ADD CONSTRAINT "user_notification_user_id_fkey"
  FOREIGN KEY ("user_id") REFERENCES "user"("id") ON DELETE CASCADE;
