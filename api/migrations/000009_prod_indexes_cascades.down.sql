-- Revert FK cascades back to the default (no cascade) behaviour.
ALTER TABLE environment_sitespeed DROP CONSTRAINT IF EXISTS environment_sitespeed_environment_id_fkey;
ALTER TABLE environment_sitespeed
  ADD CONSTRAINT environment_sitespeed_environment_id_fkey
  FOREIGN KEY (environment_id) REFERENCES environment(id);

ALTER TABLE environment_changelog DROP CONSTRAINT IF EXISTS environment_changelog_environment_id_fkey;
ALTER TABLE environment_changelog
  ADD CONSTRAINT environment_changelog_environment_id_fkey
  FOREIGN KEY (environment_id) REFERENCES environment(id);

-- Drop production indexes
DROP INDEX IF EXISTS idx_account_provider_account;
DROP INDEX IF EXISTS idx_account_provider_user;
DROP INDEX IF EXISTS idx_sso_provider_domain;
DROP INDEX IF EXISTS idx_verification_value;
DROP INDEX IF EXISTS idx_passkey_credential_id;
