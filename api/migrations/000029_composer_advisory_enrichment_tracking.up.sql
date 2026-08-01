-- Enrichment bookkeeping fixes.
--
-- github_synced_at: the GitHub pass is the only source of first_patched_versions,
-- but the shared details_synced_at marker meant an advisory enriched by NVD first
-- (any advisory with both a CVE and a GHSA) dropped out of the GitHub queue and
-- never received patched versions. GitHub sync is now tracked separately.
--
-- details_attempted_at: a failed detail fetch used to leave the row looking
-- untouched, so the fixed-order queue (ORDER BY cve LIMIT n) refetched the same
-- unresolvable identifiers every run and starved everything behind them. The
-- queues now order by least-recently-attempted.
ALTER TABLE composer_advisory ADD COLUMN "github_synced_at" timestamp;
ALTER TABLE composer_advisory ADD COLUMN "details_attempted_at" timestamp;

-- Rows already enriched from GitHub have, by definition, been through the
-- GitHub pass (including its patched-versions write).
UPDATE composer_advisory
SET github_synced_at = details_synced_at
WHERE details_source = 'github';
