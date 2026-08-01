-- name: UpsertComposerAdvisory :exec
INSERT INTO composer_advisory (
  advisory_id, title, link, cve, ghsa_id,
  severity, sources, reported_at, composer_repository,
  synced_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5,
  $6, $7, $8, $9,
  NOW(), NOW()
)
ON CONFLICT (advisory_id) DO UPDATE SET
  title = EXCLUDED.title,
  link = COALESCE(EXCLUDED.link, composer_advisory.link),
  cve = COALESCE(EXCLUDED.cve, composer_advisory.cve),
  ghsa_id = COALESCE(EXCLUDED.ghsa_id, composer_advisory.ghsa_id),
  severity = COALESCE(EXCLUDED.severity, composer_advisory.severity),
  sources = EXCLUDED.sources,
  reported_at = COALESCE(EXCLUDED.reported_at, composer_advisory.reported_at),
  composer_repository = COALESCE(EXCLUDED.composer_repository, composer_advisory.composer_repository),
  synced_at = NOW(),
  updated_at = NOW();

-- name: UpsertComposerAdvisoryPackage :exec
INSERT INTO composer_advisory_package (
  advisory_id, package_name, packagist_advisory_id, affected_versions, synced_at
) VALUES (
  $1, $2, $3, $4, NOW()
)
ON CONFLICT (advisory_id, package_name) DO UPDATE SET
  packagist_advisory_id = EXCLUDED.packagist_advisory_id,
  affected_versions = EXCLUDED.affected_versions,
  synced_at = NOW();

-- name: GetComposerAdvisory :one
SELECT * FROM composer_advisory WHERE advisory_id = $1;

-- name: GetComposerAdvisoryByPackagistID :one
SELECT a.*
FROM composer_advisory a
JOIN composer_advisory_package p ON p.advisory_id = a.advisory_id
WHERE p.packagist_advisory_id = $1
LIMIT 1;

-- name: ListComposerAdvisoryPackagesForAdvisory :many
SELECT package_name, packagist_advisory_id, affected_versions, synced_at
FROM composer_advisory_package
WHERE advisory_id = $1
ORDER BY package_name;

-- name: ListComposerAdvisoryPackagesForAdvisories :many
SELECT advisory_id, package_name, packagist_advisory_id, affected_versions, synced_at
FROM composer_advisory_package
WHERE advisory_id = ANY($1::text[])
ORDER BY advisory_id, package_name;

-- name: ListVisibleComposerAdvisories :many
SELECT *
FROM composer_advisory
WHERE is_visible = true
ORDER BY reported_at DESC NULLS LAST, advisory_id DESC;

-- Least-recently-attempted first: a fetch that keeps failing drifts to the
-- back of the queue instead of occupying the whole batch every run.
-- name: ListComposerAdvisoryCvePendingDetails :many
SELECT cve
FROM composer_advisory
WHERE cve IS NOT NULL AND cve <> ''
  AND details_synced_at IS NULL
GROUP BY cve
ORDER BY MIN(details_attempted_at) ASC NULLS FIRST, cve
LIMIT $1;

-- The GitHub pass serves two needs: details for rows nothing else enriched,
-- and first_patched_versions for every GHSA-bearing row — including those NVD
-- already enriched, hence the separate github_synced_at marker.
-- name: ListComposerAdvisoryGhsaPendingDetails :many
SELECT ghsa_id, BOOL_OR(details_synced_at IS NULL) AS needs_details
FROM composer_advisory
WHERE ghsa_id IS NOT NULL AND ghsa_id <> ''
  AND (details_synced_at IS NULL OR github_synced_at IS NULL)
GROUP BY ghsa_id
ORDER BY MIN(details_attempted_at) ASC NULLS FIRST, ghsa_id
LIMIT $1;

-- name: MarkComposerAdvisoryDetailsAttemptedByCve :exec
UPDATE composer_advisory SET details_attempted_at = NOW()
WHERE cve = $1;

-- name: MarkComposerAdvisoryDetailsAttemptedByGhsa :exec
UPDATE composer_advisory SET details_attempted_at = NOW()
WHERE ghsa_id = $1;

-- name: MarkComposerAdvisoryGithubSyncedByGhsa :exec
UPDATE composer_advisory SET github_synced_at = NOW()
WHERE ghsa_id = $1;

-- The WHERE guards stop one identifier's fetch from clobbering sibling rows
-- that share the cve/ghsa but were already enriched from another source.
-- name: UpdateComposerAdvisoryDetailsByCve :exec
UPDATE composer_advisory SET
  summary = $2,
  description = $3,
  cvss_score = $4,
  cvss_vector = $5,
  cwes = $6,
  external_references = $7,
  details_source = $8,
  details_synced_at = NOW(),
  updated_at = NOW()
WHERE cve = $1
  AND (details_synced_at IS NULL OR details_source = $8);

-- name: UpdateComposerAdvisoryDetailsByGhsa :exec
UPDATE composer_advisory SET
  summary = $2,
  description = $3,
  cvss_score = $4,
  cvss_vector = $5,
  cwes = $6,
  external_references = $7,
  details_source = $8,
  details_synced_at = NOW(),
  updated_at = NOW()
WHERE ghsa_id = $1
  AND (details_synced_at IS NULL OR details_source = $8);

-- Written only by the GitHub enrichment, which is the sole source of patched
-- versions. Kept separate from the details updates because the NVD path shares
-- those and has no equivalent field.
-- name: UpdateComposerAdvisoryFirstPatchedByGhsa :exec
UPDATE composer_advisory SET
  first_patched_versions = $2,
  updated_at = NOW()
WHERE ghsa_id = $1 AND first_patched_versions IS DISTINCT FROM $2;

-- name: ListComposerAdvisories :many
SELECT a.*
FROM composer_advisory a
WHERE
  (sqlc.narg('is_visible')::boolean IS NULL OR a.is_visible = sqlc.narg('is_visible'))
  AND (
    sqlc.narg('package_name')::text IS NULL
    OR EXISTS (
      SELECT 1 FROM composer_advisory_package p
      WHERE p.advisory_id = a.advisory_id AND p.package_name = sqlc.narg('package_name')
    )
  )
  AND (
    sqlc.narg('severity')::text IS NULL
    OR lower(COALESCE(a.severity_override, a.severity, '')) = lower(sqlc.narg('severity'))
  )
  AND (
    sqlc.narg('tag')::text IS NULL
    OR sqlc.narg('tag') = ANY(a.tags)
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR a.title ILIKE '%' || sqlc.narg('search') || '%'
    OR COALESCE(a.cve, '') ILIKE '%' || sqlc.narg('search') || '%'
    OR COALESCE(a.ghsa_id, '') ILIKE '%' || sqlc.narg('search') || '%'
    OR a.advisory_id ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM composer_advisory_package p
      WHERE p.advisory_id = a.advisory_id
        AND (
          p.package_name ILIKE '%' || sqlc.narg('search') || '%'
          OR p.packagist_advisory_id ILIKE '%' || sqlc.narg('search') || '%'
        )
    )
  )
ORDER BY a.reported_at DESC NULLS LAST, a.advisory_id DESC
LIMIT $1 OFFSET $2;

-- name: CountComposerAdvisories :one
SELECT COUNT(*)::int
FROM composer_advisory a
WHERE
  (sqlc.narg('is_visible')::boolean IS NULL OR a.is_visible = sqlc.narg('is_visible'))
  AND (
    sqlc.narg('package_name')::text IS NULL
    OR EXISTS (
      SELECT 1 FROM composer_advisory_package p
      WHERE p.advisory_id = a.advisory_id AND p.package_name = sqlc.narg('package_name')
    )
  )
  AND (
    sqlc.narg('severity')::text IS NULL
    OR lower(COALESCE(a.severity_override, a.severity, '')) = lower(sqlc.narg('severity'))
  )
  AND (
    sqlc.narg('tag')::text IS NULL
    OR sqlc.narg('tag') = ANY(a.tags)
  )
  AND (
    sqlc.narg('search')::text IS NULL
    OR a.title ILIKE '%' || sqlc.narg('search') || '%'
    OR COALESCE(a.cve, '') ILIKE '%' || sqlc.narg('search') || '%'
    OR COALESCE(a.ghsa_id, '') ILIKE '%' || sqlc.narg('search') || '%'
    OR a.advisory_id ILIKE '%' || sqlc.narg('search') || '%'
    OR EXISTS (
      SELECT 1 FROM composer_advisory_package p
      WHERE p.advisory_id = a.advisory_id
        AND (
          p.package_name ILIKE '%' || sqlc.narg('search') || '%'
          OR p.packagist_advisory_id ILIKE '%' || sqlc.narg('search') || '%'
        )
    )
  );

-- name: ListComposerAdvisoryPackages :many
SELECT DISTINCT p.package_name
FROM composer_advisory_package p
JOIN composer_advisory a ON a.advisory_id = p.advisory_id
WHERE (sqlc.narg('is_visible')::boolean IS NULL OR a.is_visible = sqlc.narg('is_visible'))
ORDER BY p.package_name;

-- name: UpdateComposerAdvisoryEnrichment :one
UPDATE composer_advisory SET
  severity_override = $2,
  is_visible = $3,
  notes_public = $4,
  notes_internal = $5,
  remediation_summary = $6,
  remediation_url = $7,
  recommended_upgrade = $8,
  shopware_impact_summary = $9,
  affected_components = $10,
  tags = $11,
  enriched_at = NOW(),
  enriched_by = $12,
  updated_at = NOW()
WHERE advisory_id = $1
RETURNING *;

-- name: GetComposerAdvisorySyncState :one
SELECT * FROM composer_advisory_sync_state WHERE id = 1;

-- name: UpsertComposerAdvisorySyncState :exec
INSERT INTO composer_advisory_sync_state (
  id, last_updated_since, last_full_sync_at, last_incremental_sync_at, last_error, updated_at
) VALUES (
  1, $1, $2, $3, $4, NOW()
)
ON CONFLICT (id) DO UPDATE SET
  last_updated_since = EXCLUDED.last_updated_since,
  last_full_sync_at = COALESCE(EXCLUDED.last_full_sync_at, composer_advisory_sync_state.last_full_sync_at),
  last_incremental_sync_at = COALESCE(EXCLUDED.last_incremental_sync_at, composer_advisory_sync_state.last_incremental_sync_at),
  last_error = EXCLUDED.last_error,
  updated_at = NOW();
