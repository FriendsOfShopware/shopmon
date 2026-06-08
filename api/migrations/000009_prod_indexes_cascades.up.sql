-- Production indexes for frequently filtered columns
CREATE INDEX IF NOT EXISTS idx_passkey_credential_id ON passkey(credential_id);
CREATE INDEX IF NOT EXISTS idx_verification_value ON verification(value);
CREATE INDEX IF NOT EXISTS idx_sso_provider_domain ON sso_provider(domain);
CREATE INDEX IF NOT EXISTS idx_account_provider_user ON account(provider_id, user_id);
CREATE INDEX IF NOT EXISTS idx_account_provider_account ON account(provider_id, account_id);

-- Add ON DELETE CASCADE to the FK from environment_changelog.environment_id -> environment(id).
-- The constraint keeps its original auto-generated name from the initial migration
-- ("shop_changelog_shop_id_fkey"); resolve it dynamically to be safe.
DO $$
DECLARE
  conname text;
BEGIN
  SELECT c.conname INTO conname
  FROM pg_constraint c
  JOIN pg_class t ON t.oid = c.conrelid
  WHERE t.relname = 'environment_changelog'
    AND c.contype = 'f'
    AND c.confrelid = 'environment'::regclass
  LIMIT 1;
  IF conname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE environment_changelog DROP CONSTRAINT %I', conname);
  END IF;
END $$;

ALTER TABLE environment_changelog
  ADD CONSTRAINT environment_changelog_environment_id_fkey
  FOREIGN KEY (environment_id) REFERENCES environment(id) ON DELETE CASCADE;

-- Add ON DELETE CASCADE to the FK from environment_sitespeed.environment_id -> environment(id).
-- (The deployment_id FK already uses ON DELETE SET NULL and is left untouched.)
DO $$
DECLARE
  conname text;
BEGIN
  SELECT c.conname INTO conname
  FROM pg_constraint c
  JOIN pg_class t ON t.oid = c.conrelid
  WHERE t.relname = 'environment_sitespeed'
    AND c.contype = 'f'
    AND c.confrelid = 'environment'::regclass
  LIMIT 1;
  IF conname IS NOT NULL THEN
    EXECUTE format('ALTER TABLE environment_sitespeed DROP CONSTRAINT %I', conname);
  END IF;
END $$;

ALTER TABLE environment_sitespeed
  ADD CONSTRAINT environment_sitespeed_environment_id_fkey
  FOREIGN KEY (environment_id) REFERENCES environment(id) ON DELETE CASCADE;
