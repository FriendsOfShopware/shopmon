-- Restore the wide per-language columns and copy translation data back.
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS label_en text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS label_de text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS short_description_en text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS short_description_de text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS description_en text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS description_de text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS installation_manual_en text;
ALTER TABLE store_extension ADD COLUMN IF NOT EXISTS installation_manual_de text;

ALTER TABLE store_extension_version ADD COLUMN IF NOT EXISTS changelog_en text;
ALTER TABLE store_extension_version ADD COLUMN IF NOT EXISTS changelog_de text;

UPDATE store_extension se SET
  label_en = t.label, short_description_en = t.short_description,
  description_en = t.description, installation_manual_en = t.installation_manual
FROM store_extension_translation t
WHERE t.extension_name = se.name AND t.language = 'en';

UPDATE store_extension se SET
  label_de = t.label, short_description_de = t.short_description,
  description_de = t.description, installation_manual_de = t.installation_manual
FROM store_extension_translation t
WHERE t.extension_name = se.name AND t.language = 'de';

UPDATE store_extension_version sev SET changelog_en = t.changelog
FROM store_extension_version_translation t
WHERE t.extension_version_id = sev.id AND t.language = 'en';

UPDATE store_extension_version sev SET changelog_de = t.changelog
FROM store_extension_version_translation t
WHERE t.extension_version_id = sev.id AND t.language = 'de';

DROP TABLE IF EXISTS store_extension_version_translation;
DROP TABLE IF EXISTS store_extension_translation;
