ALTER TABLE organization ADD COLUMN IF NOT EXISTS metadata text;

ALTER TABLE sso_provider
  ADD COLUMN IF NOT EXISTS saml_config text,
  ADD COLUMN IF NOT EXISTS user_id text REFERENCES "user"("id") ON DELETE cascade;
