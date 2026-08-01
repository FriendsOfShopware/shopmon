-- Per-environment Composer SBOM inventory and the advisory matches derived
-- from it. Like the extension upserts, the component upsert carries a
-- WHERE ... IS DISTINCT FROM guard so unchanged rows produce no WAL churn on
-- every scrape (a full composer.lock is ~200 rows per environment).

-- name: UpsertEnvironmentSbomComponent :exec
INSERT INTO environment_sbom_component (
  environment_id, package_name, version, package_type, purl, is_dev, synced_at
) VALUES ($1, $2, $3, $4, $5, $6, NOW())
ON CONFLICT (environment_id, package_name) DO UPDATE SET
  version = $3, package_type = $4, purl = $5, is_dev = $6, synced_at = NOW()
WHERE (environment_sbom_component.version, environment_sbom_component.package_type,
       environment_sbom_component.purl, environment_sbom_component.is_dev)
  IS DISTINCT FROM ($3, $4, $5, $6);

-- name: DeleteEnvironmentSbomComponentsNotIn :exec
DELETE FROM environment_sbom_component
WHERE environment_id = $1 AND package_name != ALL($2::text[]);

-- name: GetEnvironmentSbomComponents :many
SELECT package_name, version, package_type, purl, is_dev, synced_at
FROM environment_sbom_component
WHERE environment_id = $1
ORDER BY package_name;

-- name: UpsertEnvironmentSbomState :exec
INSERT INTO environment_sbom_state (
  environment_id, supported, component_count, spec_version, serial_number,
  generated_at, last_synced_at, last_error
) VALUES ($1, $2, $3, $4, $5, $6, NOW(), $7)
ON CONFLICT (environment_id) DO UPDATE SET
  supported = $2, component_count = $3, spec_version = $4, serial_number = $5,
  generated_at = $6, last_synced_at = NOW(), last_error = $7;

-- Status-only upsert for scrapes that collected nothing (unsupported endpoint
-- or fetch failure): the previously collected component_count and provenance
-- (spec_version, serial_number, generated_at) are preserved, matching the
-- component rows which are also kept.
-- name: UpsertEnvironmentSbomStateStatus :exec
INSERT INTO environment_sbom_state (
  environment_id, supported, component_count, last_synced_at, last_error
) VALUES ($1, $2, 0, NOW(), $3)
ON CONFLICT (environment_id) DO UPDATE SET
  supported = $2, last_synced_at = NOW(), last_error = $3;

-- name: GetEnvironmentSbomState :one
SELECT * FROM environment_sbom_state WHERE environment_id = $1;

-- Distinct package names across all environments: the input set for widening
-- the Packagist advisory sync beyond shopware/*.
-- name: ListDistinctSbomPackageNames :many
SELECT DISTINCT package_name FROM environment_sbom_component ORDER BY package_name;

-- name: UpsertEnvironmentAdvisoryMatch :exec
INSERT INTO environment_advisory_match (
  environment_id, advisory_id, package_name, installed_version, affected_versions, matched_at
) VALUES ($1, $2, $3, $4, $5, NOW())
ON CONFLICT (environment_id, advisory_id, package_name) DO UPDATE SET
  installed_version = $4, affected_versions = $5
WHERE (environment_advisory_match.installed_version, environment_advisory_match.affected_versions)
  IS DISTINCT FROM ($4, $5);

-- name: DeleteEnvironmentAdvisoryMatches :exec
DELETE FROM environment_advisory_match WHERE environment_id = $1;

-- Consumed by the rematch pass as its "already known" snapshot for
-- notification dedup. Deliberately ignores suppressions (notify filters those
-- itself); anything user-facing must add the suppression predicate the
-- affected-environment queries below carry.
-- name: GetEnvironmentAdvisoryMatches :many
SELECT m.advisory_id, m.package_name, m.installed_version, m.affected_versions, m.matched_at,
       a.title, a.cve, a.ghsa_id, a.severity, a.severity_override, a.link, a.remediation_url,
       a.cvss_score, a.recommended_upgrade
FROM environment_advisory_match m
JOIN composer_advisory a ON a.advisory_id = m.advisory_id
WHERE m.environment_id = $1 AND a.is_visible = true
ORDER BY a.cvss_score DESC NULLS LAST, m.advisory_id, m.package_name;

-- Environments affected by one advisory, restricted to the caller's
-- organizations. Advisories are global but shop exposure is tenant-scoped.
-- name: ListAffectedEnvironmentsForAdvisory :many
SELECT m.environment_id, m.package_name, m.installed_version, m.affected_versions, m.matched_at,
       e.name AS environment_name, e.status AS environment_status,
       e.shopware_version, e.organization_id,
       s.id AS shop_id, s.name AS shop_name,
       o.name AS organization_name,
       EXISTS (
         SELECT 1 FROM advisory_suppression sup
         WHERE sup.advisory_id = m.advisory_id
           AND sup.shop_id = e.shop_id
           AND (sup.environment_id IS NULL OR sup.environment_id = e.id)
           AND sup.revoked_at IS NULL
           AND (sup.expires_at IS NULL OR sup.expires_at > NOW())
       ) AS suppressed
FROM environment_advisory_match m
JOIN environment e ON e.id = m.environment_id
JOIN shop s ON s.id = e.shop_id
JOIN organization o ON o.id = e.organization_id
WHERE m.advisory_id = $1 AND e.organization_id = ANY($2::text[])
ORDER BY o.name, s.name, e.name, m.package_name;

-- Global impact counts for admins (all tenants). Deliberately counts
-- suppressed environments too: a suppression acknowledges exposure, it does
-- not remove it, and the fleet-wide blast radius must not shrink because a
-- tenant acknowledged an advisory.
-- name: CountAffectedEnvironmentsForAdvisories :many
SELECT advisory_id, COUNT(DISTINCT environment_id)::bigint AS environment_count
FROM environment_advisory_match
WHERE advisory_id = ANY($1::text[])
GROUP BY advisory_id;

-- Per-advisory counts limited to the caller's organizations. Excludes
-- suppressed environments with the same predicate as the scoped list queries,
-- so the detail page's count agrees with the list badge it was opened from.
-- name: CountAffectedEnvironmentsForAdvisoriesInOrgs :many
SELECT m.advisory_id, COUNT(DISTINCT m.environment_id)::bigint AS environment_count
FROM environment_advisory_match m
JOIN environment e ON e.id = m.environment_id
WHERE m.advisory_id = ANY($1::text[]) AND e.organization_id = ANY($2::text[])
  AND NOT EXISTS (
    SELECT 1 FROM advisory_suppression s
    WHERE s.advisory_id = m.advisory_id
      AND s.shop_id = e.shop_id
      AND (s.environment_id IS NULL OR s.environment_id = e.id)
      AND s.revoked_at IS NULL
      AND (s.expires_at IS NULL OR s.expires_at > NOW())
  )
GROUP BY m.advisory_id;

-- Every environment that has SBOM components, for the post-advisory-sync
-- rematch pass.
-- name: ListEnvironmentIDsWithSbom :many
SELECT DISTINCT environment_id FROM environment_sbom_component ORDER BY environment_id;

-- User-facing advisory listing with an "affects my shops" scope, sorting, and
-- a per-row affected-environment count.
--
-- affected_count is computed once here (LEFT JOIN LATERAL) so the list can both
-- sort by blast radius and render the count per row without an N+1. When
-- only_affected is true the count doubles as the scope filter.
--
-- Suppressed environments are excluded from affected_count, and suppressed_count
-- counts them separately, so an acknowledged advisory leaves the "affecting my
-- shops" to-do list without disappearing. The suppression predicate here MUST
-- stay identical to the one in CountComposerAdvisoriesScoped, otherwise the
-- scope tab badge disagrees with the list it labels.
-- name: ListComposerAdvisoriesScoped :many
SELECT a.*,
       COALESCE(m.affected_count, 0)::int AS affected_count,
       COALESCE(m.suppressed_count, 0)::int AS suppressed_count
FROM composer_advisory a
LEFT JOIN LATERAL (
  SELECT
    COUNT(DISTINCT em.environment_id) FILTER (WHERE NOT em.is_suppressed) AS affected_count,
    COUNT(DISTINCT em.environment_id) FILTER (WHERE em.is_suppressed) AS suppressed_count
  FROM (
    SELECT
      em.environment_id,
      EXISTS (
        SELECT 1 FROM advisory_suppression s
        WHERE s.advisory_id = em.advisory_id
          AND s.shop_id = e.shop_id
          AND (s.environment_id IS NULL OR s.environment_id = e.id)
          AND s.revoked_at IS NULL
          AND (s.expires_at IS NULL OR s.expires_at > NOW())
      ) AS is_suppressed
    FROM environment_advisory_match em
    JOIN environment e ON e.id = em.environment_id
    WHERE em.advisory_id = a.advisory_id
      AND e.organization_id = ANY(sqlc.arg('organization_ids')::text[])
  ) em
) m ON true
WHERE
  a.is_visible = true
  AND (NOT sqlc.arg('only_affected')::boolean OR COALESCE(m.affected_count, 0) > 0)
  AND (NOT sqlc.arg('only_suppressed')::boolean OR COALESCE(m.suppressed_count, 0) > 0)
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
ORDER BY
  -- Severity is stored as text, so map it to a rank rather than sorting
  -- alphabetically (which would put "critical" after "high").
  CASE WHEN sqlc.arg('sort')::text = 'severity' THEN
    CASE lower(COALESCE(a.severity_override, a.severity, ''))
      WHEN 'critical' THEN 5
      WHEN 'high' THEN 4
      WHEN 'medium' THEN 3
      WHEN 'low' THEN 2
      WHEN 'none' THEN 1
      ELSE 0
    END
  END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort')::text = 'affected' THEN COALESCE(m.affected_count, 0) END DESC NULLS LAST,
  CASE WHEN sqlc.arg('sort')::text = 'cvss' THEN a.cvss_score END DESC NULLS LAST,
  a.reported_at DESC NULLS LAST,
  a.advisory_id DESC
LIMIT sqlc.arg('limit_') OFFSET sqlc.arg('offset_');

-- Scope tab badges: total matching the non-scope filters, how many of those
-- affect the caller's environments, and how many are suppressed. All from one
-- scan. The LATERAL below is intentionally identical to the one in
-- ListComposerAdvisoriesScoped -- if they drift, the badge lies about the list.
-- name: CountComposerAdvisoriesScoped :one
SELECT
  COUNT(*)::int AS all_count,
  COUNT(*) FILTER (WHERE COALESCE(m.affected_count, 0) > 0)::int AS affected_count,
  COUNT(*) FILTER (WHERE COALESCE(m.suppressed_count, 0) > 0)::int AS suppressed_count
FROM composer_advisory a
LEFT JOIN LATERAL (
  SELECT
    COUNT(DISTINCT em.environment_id) FILTER (WHERE NOT em.is_suppressed) AS affected_count,
    COUNT(DISTINCT em.environment_id) FILTER (WHERE em.is_suppressed) AS suppressed_count
  FROM (
    SELECT
      em.environment_id,
      EXISTS (
        SELECT 1 FROM advisory_suppression s
        WHERE s.advisory_id = em.advisory_id
          AND s.shop_id = e.shop_id
          AND (s.environment_id IS NULL OR s.environment_id = e.id)
          AND s.revoked_at IS NULL
          AND (s.expires_at IS NULL OR s.expires_at > NOW())
      ) AS is_suppressed
    FROM environment_advisory_match em
    JOIN environment e ON e.id = em.environment_id
    WHERE em.advisory_id = a.advisory_id
      AND e.organization_id = ANY(sqlc.arg('organization_ids')::text[])
  ) em
) m ON true
WHERE
  a.is_visible = true
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
