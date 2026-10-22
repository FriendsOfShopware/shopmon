-- GitHub-imported advisories have no representation before this migration;
-- deleting them cascades to packages, matches, notifications, suppressions.
DELETE FROM "composer_advisory" WHERE origin = 'github';

ALTER TABLE "composer_advisory"
  DROP CONSTRAINT composer_advisory_origin_check,
  ADD CONSTRAINT composer_advisory_origin_check CHECK (origin IN ('upstream', 'shopmon'));
