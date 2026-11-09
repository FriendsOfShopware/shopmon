CREATE TABLE notification_destination (
  id text PRIMARY KEY,
  organization_id text NOT NULL REFERENCES organization(id) ON DELETE CASCADE,
  shop_id integer REFERENCES shop(id) ON DELETE CASCADE,
  name text NOT NULL,
  dsn text NOT NULL,
  created_at timestamp NOT NULL DEFAULT NOW(),
  updated_at timestamp NOT NULL DEFAULT NOW()
);
CREATE INDEX notification_destination_scope ON notification_destination (organization_id, shop_id);
