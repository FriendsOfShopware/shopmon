-- Normalize environment_changelog.extensions JSONB to the API/frontend shape.
--
-- Two legacy issues are fixed for every extension diff element:
--   1. Version keys "old_version"/"new_version" (snake_case, written by the
--      scrape job) are renamed to "oldVersion"/"newVersion" (camelCase).
--   2. Nested changelog "creationDate" values in the store format
--      ("2022-10-18 11:59:38.000000") are converted to RFC3339
--      ("2022-10-18T11:59:38Z").
--
-- Elements that already use the new shape are left untouched. JSON null values
-- are preserved (e.g. new_version: null on "removed" diffs).
UPDATE environment_changelog
SET extensions = (
    SELECT jsonb_agg(
        (
            -- 1. rename version keys
            CASE
                WHEN ext ? 'old_version' OR ext ? 'new_version' THEN
                    jsonb_set(
                        jsonb_set(ext, '{oldVersion}', COALESCE(ext->'old_version', 'null'::jsonb)),
                        '{newVersion}', COALESCE(ext->'new_version', 'null'::jsonb)
                    ) - 'old_version' - 'new_version'
                ELSE ext
            END
        ) || (
            -- 2. fix nested changelog creationDate format (merge over the result above)
            CASE
                WHEN jsonb_typeof(ext->'changelog') = 'array' THEN
                    jsonb_build_object('changelog', (
                        SELECT COALESCE(jsonb_agg(
                            CASE
                                WHEN entry ? 'creationDate'
                                     AND (entry->>'creationDate') !~ 'T.*Z$'
                                     AND (entry->>'creationDate') <> '' THEN
                                    jsonb_set(
                                        entry,
                                        '{creationDate}',
                                        to_jsonb(to_char(
                                            (entry->>'creationDate')::timestamp,
                                            'YYYY-MM-DD"T"HH24:MI:SS"Z"'
                                        ))
                                    )
                                ELSE entry
                            END
                        ), '[]'::jsonb)
                        FROM jsonb_array_elements(ext->'changelog') AS entry
                    ))
                ELSE '{}'::jsonb
            END
        )
    )
    FROM jsonb_array_elements(extensions) AS ext
)
WHERE extensions IS NOT NULL
  AND (extensions::text LIKE '%"old_version"%'
       OR extensions::text LIKE '%"new_version"%'
       OR (extensions::text LIKE '%"creationDate"%'
           AND extensions::text NOT LIKE '%T%:%Z%'));
