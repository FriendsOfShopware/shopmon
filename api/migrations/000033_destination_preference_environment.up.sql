-- A real foreign key coordinates preference writes with environment deletion,
-- including shop/organization cascades and concurrent requests.
DELETE FROM destination_notification_preference p
WHERE scope_type = 'environment' AND NOT EXISTS (
  SELECT 1 FROM environment e WHERE e.id::text = p.scope_id
);
ALTER TABLE destination_notification_preference
ADD COLUMN environment_id integer GENERATED ALWAYS AS
  (CASE WHEN scope_type = 'environment' THEN scope_id::integer END) STORED
  REFERENCES environment(id) ON DELETE CASCADE;
