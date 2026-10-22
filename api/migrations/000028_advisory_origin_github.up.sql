-- Third advisory origin: advisories imported by an admin directly from a
-- GitHub security advisory (GHSA) that Packagist has not picked up (yet) —
-- typically because the CVE is still pending. They keep the GHSA id as their
-- advisory_id and are official advisories, just not synced ones, so neither
-- 'upstream' (synced) nor 'shopmon' (authored by Shopmon) describes them.
ALTER TABLE "composer_advisory"
  DROP CONSTRAINT composer_advisory_origin_check,
  ADD CONSTRAINT composer_advisory_origin_check CHECK (origin IN ('upstream', 'shopmon', 'github'));
