DROP INDEX idx_advisory_suppression_environment;
ALTER TABLE advisory_suppression
  DROP CONSTRAINT advisory_suppression_environment_shop_fkey;
ALTER TABLE environment
  DROP CONSTRAINT environment_shop_id_id_key;
