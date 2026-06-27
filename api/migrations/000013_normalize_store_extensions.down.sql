-- Reverse the store-extension normalization: re-add the denormalized columns to
-- environment_extension, move store-linked extensions back into it (rebuilding
-- the English changelog JSON from the version catalog), then drop the new tables.

ALTER TABLE environment_extension ADD COLUMN IF NOT EXISTS rating_average integer;
ALTER TABLE environment_extension ADD COLUMN IF NOT EXISTS store_link text;
ALTER TABLE environment_extension ADD COLUMN IF NOT EXISTS changelog jsonb;

INSERT INTO environment_extension
  (environment_id, name, label, active, version, latest_version, installed, rating_average, store_link, changelog, installed_at)
SELECT
  ese.environment_id,
  ese.extension_name,
  ese.label,
  ese.active,
  ese.version,
  ese.latest_version,
  ese.installed,
  se.rating_average,
  se.store_link,
  (
    SELECT jsonb_agg(jsonb_build_object(
      'version', v.version,
      'text', v.changelog_en,
      'creationDate', v.released_at
    ) ORDER BY v.version)
    FROM store_extension_version v
    WHERE v.extension_name = ese.extension_name
      AND v.changelog_en IS NOT NULL
  ),
  ese.installed_at
FROM environment_store_extension ese
JOIN store_extension se ON se.name = ese.extension_name
ON CONFLICT (environment_id, name) DO NOTHING;

DROP TABLE IF EXISTS environment_store_extension;
DROP TABLE IF EXISTS store_extension_image;
DROP TABLE IF EXISTS store_extension_version;
DROP TABLE IF EXISTS store_extension;
