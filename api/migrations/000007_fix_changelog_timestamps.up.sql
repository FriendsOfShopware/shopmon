-- Fix creationDate timestamps in environment_changelog.extensions JSONB
-- Old format: "2026-01-05 12:44:57.000000"
-- New format: "2026-01-05T12:44:57.000000Z"
UPDATE environment_changelog
SET extensions = (
    SELECT jsonb_agg(
        CASE
            WHEN ext->'changelog' IS NOT NULL AND jsonb_typeof(ext->'changelog') = 'array' THEN
                jsonb_set(
                    ext,
                    '{changelog}',
                    (
                        SELECT COALESCE(jsonb_agg(
                            jsonb_set(
                                entry,
                                '{creationDate}',
                                to_jsonb(to_char(
                                    (entry->>'creationDate')::timestamp,
                                    'YYYY-MM-DD"T"HH24:MI:SS"Z"'
                                ))
                            )
                        ), '[]'::jsonb)
                        FROM jsonb_array_elements(ext->'changelog') AS entry
                    )
                )
            ELSE ext
        END
    )
    FROM jsonb_array_elements(extensions) AS ext
)
WHERE extensions IS NOT NULL
  AND extensions::text LIKE '%"creationDate"%'
  AND extensions::text NOT LIKE '%T%:%Z%';
