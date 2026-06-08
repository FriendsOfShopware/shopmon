ALTER TABLE member ADD CONSTRAINT member_org_user_unique UNIQUE (organization_id, user_id);
