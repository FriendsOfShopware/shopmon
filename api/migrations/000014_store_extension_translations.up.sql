-- Normalize the per-language store metadata columns (label_en/label_de, …,
-- changelog_en/changelog_de) into dedicated translation tables keyed by
-- language. This lets the API resolve a single language per request (with an
-- English fallback) by joining on the requested language, instead of shipping
-- every language to the client. Languages are still a fixed set today (en/de),
-- but adding one is now a data change, not a schema change.

CREATE TABLE IF NOT EXISTS "store_extension_translation" (
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "language" text NOT NULL,
  "label" text,
  "short_description" text,
  "description" text,
  "installation_manual" text,
  PRIMARY KEY ("extension_name", "language")
);

CREATE TABLE IF NOT EXISTS "store_extension_version_translation" (
  "extension_version_id" integer NOT NULL REFERENCES "store_extension_version"("id") ON DELETE cascade,
  "language" text NOT NULL,
  "changelog" text,
  PRIMARY KEY ("extension_version_id", "language")
);

-- Backfill the extension translations from the wide columns. Only insert a
-- language row when at least one field for that language is present.
INSERT INTO store_extension_translation (extension_name, language, label, short_description, description, installation_manual)
SELECT name, 'en', label_en, short_description_en, description_en, installation_manual_en
FROM store_extension
WHERE label_en IS NOT NULL OR short_description_en IS NOT NULL
   OR description_en IS NOT NULL OR installation_manual_en IS NOT NULL
ON CONFLICT (extension_name, language) DO NOTHING;

INSERT INTO store_extension_translation (extension_name, language, label, short_description, description, installation_manual)
SELECT name, 'de', label_de, short_description_de, description_de, installation_manual_de
FROM store_extension
WHERE label_de IS NOT NULL OR short_description_de IS NOT NULL
   OR description_de IS NOT NULL OR installation_manual_de IS NOT NULL
ON CONFLICT (extension_name, language) DO NOTHING;

-- Backfill the per-version changelog translations.
INSERT INTO store_extension_version_translation (extension_version_id, language, changelog)
SELECT id, 'en', changelog_en FROM store_extension_version WHERE changelog_en IS NOT NULL
ON CONFLICT (extension_version_id, language) DO NOTHING;

INSERT INTO store_extension_version_translation (extension_version_id, language, changelog)
SELECT id, 'de', changelog_de FROM store_extension_version WHERE changelog_de IS NOT NULL
ON CONFLICT (extension_version_id, language) DO NOTHING;

-- Drop the now-normalized wide columns.
ALTER TABLE store_extension DROP COLUMN IF EXISTS label_en;
ALTER TABLE store_extension DROP COLUMN IF EXISTS label_de;
ALTER TABLE store_extension DROP COLUMN IF EXISTS short_description_en;
ALTER TABLE store_extension DROP COLUMN IF EXISTS short_description_de;
ALTER TABLE store_extension DROP COLUMN IF EXISTS description_en;
ALTER TABLE store_extension DROP COLUMN IF EXISTS description_de;
ALTER TABLE store_extension DROP COLUMN IF EXISTS installation_manual_en;
ALTER TABLE store_extension DROP COLUMN IF EXISTS installation_manual_de;

ALTER TABLE store_extension_version DROP COLUMN IF EXISTS changelog_en;
ALTER TABLE store_extension_version DROP COLUMN IF EXISTS changelog_de;
