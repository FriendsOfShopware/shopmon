-- Shared destinations use the same scope/event/channel preference model as
-- personal notifications, with a destination owner instead of a user owner.
CREATE TABLE destination_notification_preference (
  destination_id text NOT NULL REFERENCES notification_destination(id) ON DELETE CASCADE,
  scope_type text NOT NULL CHECK (scope_type IN ('global', 'environment')),
  scope_id text NOT NULL DEFAULT '',
  event_type text NOT NULL DEFAULT '',
  channel text NOT NULL DEFAULT 'destination' CHECK (channel = 'destination'),
  enabled boolean NOT NULL,
  PRIMARY KEY (destination_id, scope_type, scope_id, event_type, channel)
);
