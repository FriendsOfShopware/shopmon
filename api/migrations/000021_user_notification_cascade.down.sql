ALTER TABLE "user_notification" DROP CONSTRAINT IF EXISTS "user_notification_user_id_fkey";
ALTER TABLE "user_notification"
  ADD CONSTRAINT "user_notification_user_id_fkey"
  FOREIGN KEY ("user_id") REFERENCES "user"("id");
