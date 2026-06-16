CREATE TABLE IF NOT EXISTS audit_log (
  "id" bigserial PRIMARY KEY,
  "actor_user_id" text REFERENCES "user"("id") ON DELETE SET NULL,
  "action" text NOT NULL,
  "target_user_id" text REFERENCES "user"("id") ON DELETE SET NULL,
  "detail" text,
  "ip_address" text,
  "created_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_audit_log_actor ON audit_log (actor_user_id);
CREATE INDEX IF NOT EXISTS idx_audit_log_created_at ON audit_log (created_at);
