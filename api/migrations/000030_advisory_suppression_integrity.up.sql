-- advisory_suppression carried independent FKs on shop_id and environment_id,
-- so nothing in the database tied the two together: a writer bypassing the
-- service could store an environment belonging to a different shop. Such a row
-- is invisible to every read query (they join on s.shop_id = e.shop_id AND
-- s.environment_id = e.id) yet still occupies the partial unique index,
-- blocking the correct suppression. The composite FK enforces the pairing;
-- MATCH SIMPLE skips rows whose environment_id is NULL (shop-wide scope).
ALTER TABLE environment
  ADD CONSTRAINT environment_shop_id_id_key UNIQUE (shop_id, id);

ALTER TABLE advisory_suppression
  ADD CONSTRAINT advisory_suppression_environment_shop_fkey
  FOREIGN KEY (shop_id, environment_id)
  REFERENCES environment (shop_id, id) ON DELETE CASCADE;

-- Serves the ON DELETE CASCADE from environment; the existing lookup index
-- leads with shop_id and cannot.
CREATE INDEX idx_advisory_suppression_environment
  ON advisory_suppression (environment_id) WHERE environment_id IS NOT NULL;
