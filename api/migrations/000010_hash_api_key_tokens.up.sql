-- API key tokens are now stored as a sha256 hex digest at rest instead of in
-- plaintext (see internal/handler/api_key.go HashApiKeyToken). Existing rows
-- still hold the plaintext token, so the application's hashed lookup would no
-- longer match them. Hash every existing token in place exactly once so all
-- previously-issued keys keep working transparently after the change ships.
--
-- This migration runs a single time (tracked by the migration version), so the
-- unconditional UPDATE will not double-hash on re-run.

CREATE EXTENSION IF NOT EXISTS pgcrypto;

UPDATE "shop_api_key"
SET "token" = encode(digest("token", 'sha256'), 'hex');
