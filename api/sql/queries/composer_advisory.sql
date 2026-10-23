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
-- link is carried along so a GHSA missing from the global advisory database can
-- still be fetched from the repository that published it. Only a
-- repository-scoped link serves that purpose, so siblings sharing the GHSA are
-- filtered to that shape rather than picked by sort order: a global
-- github.com/advisories link sorts first but names no repository, and would
-- leave the advisory unenriched forever.
-- name: ListComposerAdvisoryGhsaPendingDetails :many
SELECT ghsa_id,
       BOOL_OR(details_synced_at IS NULL) AS needs_details,
       COALESCE(MIN(link) FILTER (
         WHERE link ~ '^https://github\.com/[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/security/advisories/GHSA-'
       ), '')::text AS link
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
    sqlc.narg('origin')::text IS NULL
    OR a.origin = sqlc.narg('origin')
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
    sqlc.narg('origin')::text IS NULL
    OR a.origin = sqlc.narg('origin')
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

-- Shopmon-authored advisory numbering ("SHOPMON-2026-0001"); the sequence is
-- global so an id is never reissued for a different vulnerability.
-- name: NextShopmonAdvisoryNumber :one
SELECT nextval('shopmon_advisory_seq')::bigint AS number;

-- Serialize catalog identity changes before resolving aliases or moving
-- packages. Held only during database work, never while fetching upstream.
-- name: LockComposerAdvisoryCatalog :exec
SELECT pg_advisory_xact_lock(hashtextextended('shopmon:composer-advisory-catalog', 0));

-- name: InsertManualAdvisory :one
INSERT INTO composer_advisory (
  advisory_id, title, link, cve, ghsa_id, severity, sources, reported_at,
  summary, description, cvss_score, cvss_vector, cwes, external_references,
  details_source, details_synced_at, github_synced_at,
  is_visible, notes_public, notes_internal,
  remediation_summary, remediation_url, recommended_upgrade,
  shopware_impact_summary, affected_components, tags, origin, created_by,
  synced_at, created_at, updated_at
) VALUES (
  $1, $2, $3, $4, $5, $6, $7, $8,
  $9, $10, $11, $12, $13, $14,
  $15, $16, $17,
  $18, $19, $20,
  $21, $22, $23,
  $24, $25, $26, $27, $28,
  NOW(), NOW(), NOW()
)
RETURNING *;

-- Full content rewrite for Shopmon-authored rows (admin edit) and for the
-- link/merge path, where the service computes the surviving row's content
-- from the Shopmon row and the official row being merged into it. Enrichment
-- columns are deliberately absent; UpdateComposerAdvisoryEnrichment owns them.
-- name: UpdateComposerAdvisoryContent :one
UPDATE composer_advisory SET
  title = $2,
  link = $3,
  cve = $4,
  ghsa_id = $5,
  severity = $6,
  sources = $7,
  reported_at = $8,
  composer_repository = $9,
  summary = $10,
  description = $11,
  cvss_score = $12,
  cvss_vector = $13,
  cwes = $14,
  external_references = $15,
  details_source = $16,
  details_synced_at = $17,
  github_synced_at = $18,
  first_patched_versions = $19,
  updated_at = NOW()
WHERE advisory_id = $1
RETURNING *;

-- name: DeleteComposerAdvisoryPackages :exec
DELETE FROM composer_advisory_package WHERE advisory_id = $1;

-- Official data replaces manual ranges on first takeover. Preserve already
-- synced packages, including those outside the current package-scoped fetch.
-- name: DeleteManualComposerAdvisoryPackages :exec
DELETE FROM composer_advisory_package
WHERE advisory_id = $1 AND packagist_advisory_id IS NULL;

-- name: GetComposerAdvisoryForUpdate :one
SELECT * FROM composer_advisory WHERE advisory_id = $1 FOR UPDATE;

-- Resolution fallbacks so a linked Shopmon advisory answers to its official
-- identifiers, mirroring GetComposerAdvisoryByPackagistID. A Shopmon row wins
-- over theoretical siblings sharing the identifier because it is the row
-- suppressions and alerts were recorded against.
-- name: GetComposerAdvisoryByCve :one
SELECT * FROM composer_advisory
WHERE cve = $1
ORDER BY (origin = 'shopmon') DESC, advisory_id
LIMIT 1;

-- name: GetComposerAdvisoryByGhsa :one
SELECT * FROM composer_advisory
WHERE ghsa_id = $1
ORDER BY (origin = 'shopmon') DESC, advisory_id
LIMIT 1;

-- Sync-time merge target: a manually tracked row (Shopmon-authored or
-- GitHub-imported) already linked to the CVE/GHSA a Packagist advisory
-- carries. Upserting under its id keeps one row per vulnerability instead
-- of a manual row plus a canonical CVE row.
-- name: GetManualAdvisoryByIdentifiers :one
SELECT * FROM composer_advisory
WHERE origin IN ('shopmon', 'github')
  AND (
    (sqlc.narg('cve')::text IS NOT NULL AND cve = sqlc.narg('cve'))
    OR (sqlc.narg('ghsa_id')::text IS NOT NULL AND ghsa_id = sqlc.narg('ghsa_id'))
  )
ORDER BY advisory_id
LIMIT 1;

-- Every row claiming one of the identifiers, by column or by canonical id.
-- Drives duplicate detection on create and merge-candidate lookup on link.
-- name: ListComposerAdvisoriesByIdentifiers :many
SELECT * FROM composer_advisory
WHERE
  (sqlc.narg('cve')::text IS NOT NULL AND (cve = sqlc.narg('cve') OR advisory_id = sqlc.narg('cve')))
  OR (sqlc.narg('ghsa_id')::text IS NOT NULL AND (ghsa_id = sqlc.narg('ghsa_id') OR advisory_id = sqlc.narg('ghsa_id')))
ORDER BY advisory_id;

-- Merge of an already-synced official row ($2) into the surviving Shopmon row
-- ($1). Matches and notification markers move first so affected shops keep
-- their state and are never re-alerted; whatever cannot move without
-- colliding with the survivor's own rows is left behind and removed by the
-- cascade when the official row is deleted.
-- name: MergeEnvironmentAdvisoryMatches :exec
UPDATE environment_advisory_match m SET advisory_id = $1
WHERE m.advisory_id = $2
  AND NOT EXISTS (
    SELECT 1 FROM environment_advisory_match keep
    WHERE keep.environment_id = m.environment_id
      AND keep.advisory_id = $1
      AND keep.package_name = m.package_name
  );

-- name: MergeEnvironmentAdvisoryNotifications :exec
INSERT INTO environment_advisory_notified (environment_id, advisory_id, notified_at)
SELECT src.environment_id, $1, src.notified_at
FROM environment_advisory_notified src
WHERE src.advisory_id = $2
ON CONFLICT (environment_id, advisory_id) DO NOTHING;

-- Revoked suppressions are pure audit trail and move unconditionally; active
-- ones move only where the survivor has no active suppression for the same
-- shop/environment (the partial unique indexes would reject the rest, and
-- those are duplicates of the survivor's own acknowledgement).
-- name: MergeRevokedAdvisorySuppressions :exec
UPDATE advisory_suppression SET advisory_id = $1
WHERE advisory_id = $2 AND revoked_at IS NOT NULL;

-- Expired rows still occupy the partial unique indexes. Retire them on both
-- sides before moving suppressions so a valid acknowledgement can survive.
-- name: RevokeExpiredSuppressionsForAdvisoryMerge :exec
UPDATE advisory_suppression SET revoked_at = NOW()
WHERE advisory_id = ANY(sqlc.arg('advisory_ids')::text[])
  AND revoked_at IS NULL AND expires_at <= NOW();

-- name: MergeActiveAdvisorySuppressions :exec
UPDATE advisory_suppression s SET advisory_id = $1
WHERE s.advisory_id = $2
  AND s.revoked_at IS NULL
  AND NOT EXISTS (
    SELECT 1 FROM advisory_suppression keep
    WHERE keep.advisory_id = $1
      AND keep.revoked_at IS NULL
      AND keep.shop_id = s.shop_id
      AND keep.environment_id IS NOT DISTINCT FROM s.environment_id
  );

-- Re-pointing the official package rows (after the survivor's own rows are
-- deleted) adopts the upstream ranges and PKSA ids without ever holding two
-- rows with the same packagist_advisory_id.
-- name: MergeComposerAdvisoryPackages :exec
UPDATE composer_advisory_package SET advisory_id = $1
WHERE advisory_id = $2;

-- name: DeleteComposerAdvisory :exec
DELETE FROM composer_advisory WHERE advisory_id = $1;

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
