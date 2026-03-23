ALTER TABLE sso_provider ADD CONSTRAINT fk_sso_provider_organization
  FOREIGN KEY (organization_id) REFERENCES organization(id) ON DELETE CASCADE;
