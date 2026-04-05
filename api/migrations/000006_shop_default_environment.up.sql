ALTER TABLE shop ADD COLUMN default_environment_id integer REFERENCES environment(id) ON DELETE RESTRICT;

-- Backfill: pick one environment per shop (lowest id)
UPDATE shop SET default_environment_id = (
    SELECT e.id FROM environment e WHERE e.shop_id = shop.id ORDER BY e.id LIMIT 1
);

-- Delete shops that have no environments (they are unusable without a default)
DELETE FROM shop WHERE default_environment_id IS NULL;
