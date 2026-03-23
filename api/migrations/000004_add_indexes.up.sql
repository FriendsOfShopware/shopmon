CREATE INDEX IF NOT EXISTS idx_member_org_user ON member(organization_id, user_id);
CREATE INDEX IF NOT EXISTS idx_shop_organization ON shop(organization_id);
CREATE INDEX IF NOT EXISTS idx_deployment_shop ON deployment(shop_id);
CREATE INDEX IF NOT EXISTS idx_user_notification_user ON user_notification(user_id);
CREATE INDEX IF NOT EXISTS idx_shop_extension_shop ON shop_extension(shop_id);
CREATE INDEX IF NOT EXISTS idx_session_user ON session(user_id);
CREATE INDEX IF NOT EXISTS idx_account_user ON account(user_id);
