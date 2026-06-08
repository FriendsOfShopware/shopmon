-- Revert the version key rename: oldVersion/newVersion -> old_version/new_version
UPDATE environment_changelog
SET extensions = (
    SELECT jsonb_agg(
        CASE
            WHEN ext ? 'oldVersion' OR ext ? 'newVersion' THEN
                (
                    jsonb_set(
                        jsonb_set(
                            ext,
                            '{old_version}',
                            COALESCE(ext->'oldVersion', 'null'::jsonb)
                        ),
                        '{new_version}',
                        COALESCE(ext->'newVersion', 'null'::jsonb)
                    )
                    - 'oldVersion'
                    - 'newVersion'
                )
            ELSE ext
        END
    )
    FROM jsonb_array_elements(extensions) AS ext
)
WHERE extensions IS NOT NULL
  AND (extensions::text LIKE '%"oldVersion"%'
       OR extensions::text LIKE '%"newVersion"%');
