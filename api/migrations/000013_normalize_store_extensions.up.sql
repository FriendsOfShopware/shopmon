-- Normalize Shopware-store extension data out of the per-environment
-- environment_extension table into a shared catalog (store_extension), a
-- per-version changelog catalog (store_extension_version), a listing-image
-- table (store_extension_image), and a per-environment link table
-- (environment_store_extension). After this migration environment_extension
-- holds only extensions that are unknown to the store.

CREATE TABLE IF NOT EXISTS "store_extension" (
  "name" text PRIMARY KEY NOT NULL,
  "store_id" integer,
  "icon_url" text,
  "producer_name" text,
  "producer_website" text,
  "rating_average" integer,
  "store_link" text,
  "release_date" text,
  "label_en" text,
  "label_de" text,
  "short_description_en" text,
  "short_description_de" text,
  "description_en" text,
  "description_de" text,
  "installation_manual_en" text,
  "installation_manual_de" text,
  "latest_version" text,
  "last_refreshed_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS "store_extension_version" (
  "id" serial PRIMARY KEY NOT NULL,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "version" text NOT NULL,
  "changelog_en" text,
  "changelog_de" text,
  "released_at" text,
  UNIQUE ("extension_name", "version")
);

CREATE TABLE IF NOT EXISTS "store_extension_image" (
  "id" serial PRIMARY KEY NOT NULL,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "url" text NOT NULL,
  "preview" boolean NOT NULL DEFAULT false,
  "priority" integer NOT NULL DEFAULT 0,
  UNIQUE ("extension_name", "url")
);

CREATE TABLE IF NOT EXISTS "environment_store_extension" (
  "id" serial PRIMARY KEY NOT NULL,
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "label" text NOT NULL,
  "version" text NOT NULL,
  "latest_version" text,
  "active" boolean NOT NULL,
  "installed" boolean NOT NULL,
  "installed_at" text,
  UNIQUE ("environment_id", "extension_name")
);

CREATE INDEX IF NOT EXISTS idx_store_extension_version_name ON store_extension_version (extension_name);
CREATE INDEX IF NOT EXISTS idx_store_extension_image_name ON store_extension_image (extension_name);
CREATE INDEX IF NOT EXISTS idx_environment_store_extension_env ON environment_store_extension (environment_id);
CREATE INDEX IF NOT EXISTS idx_environment_store_extension_name ON environment_store_extension (extension_name);

-- Backfill the catalog. An extension is considered store-known if it was
-- enriched with a store_link during a previous scrape. Rich metadata (icon,
-- producer, descriptions, pictures, German changelog) is not available from the
-- old denormalized rows and will be populated on the next scrape; the existing
-- label is kept as the English label as a best-effort fallback.
INSERT INTO store_extension (name, rating_average, store_link, label_en, latest_version, last_refreshed_at)
SELECT DISTINCT ON (ee.name)
  ee.name,
  ee.rating_average,
  ee.store_link,
  ee.label,
  ee.latest_version,
  NOW()
FROM environment_extension ee
WHERE ee.store_link IS NOT NULL
ORDER BY ee.name, ee.id
ON CONFLICT (name) DO NOTHING;

-- Backfill the per-version changelog catalog from the stored (English) changelog
-- JSON. Different environments may carry overlapping version subsets, so collapse
-- duplicates by (name, version).
INSERT INTO store_extension_version (extension_name, version, changelog_en, released_at)
SELECT DISTINCT ON (ee.name, elem->>'version')
  ee.name,
  elem->>'version',
  elem->>'text',
  NULLIF(elem->>'creationDate', '')
FROM environment_extension ee
CROSS JOIN LATERAL jsonb_array_elements(ee.changelog) AS elem
WHERE ee.store_link IS NOT NULL
  AND ee.changelog IS NOT NULL
  AND jsonb_typeof(ee.changelog) = 'array'
  AND COALESCE(elem->>'version', '') <> ''
ORDER BY ee.name, elem->>'version', ee.id
ON CONFLICT (extension_name, version) DO NOTHING;

-- Backfill the per-environment links.
INSERT INTO environment_store_extension (environment_id, extension_name, label, version, latest_version, active, installed, installed_at)
SELECT ee.environment_id, ee.name, ee.label, ee.version, ee.latest_version, ee.active, ee.installed, ee.installed_at
FROM environment_extension ee
WHERE ee.store_link IS NOT NULL
ON CONFLICT (environment_id, extension_name) DO NOTHING;

-- Remove the migrated rows; environment_extension now holds unknown extensions only.
DELETE FROM environment_extension WHERE store_link IS NOT NULL;

ALTER TABLE environment_extension DROP COLUMN IF EXISTS rating_average;
ALTER TABLE environment_extension DROP COLUMN IF EXISTS store_link;
ALTER TABLE environment_extension DROP COLUMN IF EXISTS changelog;
