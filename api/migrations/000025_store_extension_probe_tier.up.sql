-- Split store extension sync bookkeeping into two freshness tiers:
-- last_synced_at keeps gating the cheap packages.shopware.com feed pass
-- (compatibility + new-release detection, hourly), while the new
-- last_store_probe_at gates the rate-limited api.shopware.com probes
-- (metadata, translations, pictures, bilingual changelogs, store membership),
-- which only need to happen once per day per extension name.
ALTER TABLE "store_extension_sync"
  ADD COLUMN "last_store_probe_at" timestamp;

-- Existing rows were fully synced (store probe included) at last_synced_at, so
-- start the store-probe tier from there instead of re-probing every known
-- extension right after deploy.
UPDATE "store_extension_sync" SET last_store_probe_at = last_synced_at;
