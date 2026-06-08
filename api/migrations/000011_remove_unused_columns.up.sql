ALTER TABLE organization DROP COLUMN IF EXISTS metadata;

ALTER TABLE sso_provider
  DROP COLUMN IF EXISTS saml_config,
  DROP COLUMN IF EXISTS user_id;
