-- name: CreateStoreExtensionReport :one
INSERT INTO store_extension_report (extension_name, user_id, category, comment)
VALUES ($1, $2, $3, $4)
RETURNING id;

-- name: StoreExtensionExists :one
SELECT COUNT(*) > 0 AS exists FROM store_extension WHERE name = $1;

-- name: CountUserPendingReportsForExtension :one
SELECT COUNT(*)::int FROM store_extension_report
WHERE extension_name = $1 AND user_id = $2 AND status = 'pending';

-- GetApprovedExtensionReportSummary returns, per extension, the number of
-- approved reports grouped by category. Used to surface community warnings on
-- the extension lists.
-- name: GetApprovedExtensionReportSummary :many
SELECT extension_name, category, COUNT(*)::int AS count
FROM store_extension_report
WHERE status = 'approved'
GROUP BY extension_name, category
ORDER BY extension_name, category;

-- name: AdminListExtensionReports :many
SELECT r.id, r.extension_name, set_en.label AS extension_label, r.category, r.comment,
       r.status, r.created_at, r.reviewed_at,
       r.user_id AS reporter_id, reporter.name AS reporter_name, reporter.email AS reporter_email,
       r.reviewed_by, reviewer.name AS reviewer_name
FROM store_extension_report r
JOIN store_extension se ON se.name = r.extension_name
LEFT JOIN store_extension_translation set_en ON set_en.extension_name = r.extension_name AND set_en.language = 'en'
LEFT JOIN "user" reporter ON reporter.id = r.user_id
LEFT JOIN "user" reviewer ON reviewer.id = r.reviewed_by
WHERE (sqlc.narg('status')::text IS NULL OR r.status = sqlc.narg('status'))
ORDER BY r.created_at DESC
LIMIT $1 OFFSET $2;

-- name: AdminCountExtensionReports :one
SELECT COUNT(*)::int FROM store_extension_report r
WHERE (sqlc.narg('status')::text IS NULL OR r.status = sqlc.narg('status'));

-- name: GetStoreExtensionReportByID :one
SELECT id, extension_name, user_id, category, comment, status, reviewed_by, reviewed_at, created_at
FROM store_extension_report WHERE id = $1;

-- name: SetStoreExtensionReportStatus :exec
UPDATE store_extension_report
SET status = $2, reviewed_by = $3, reviewed_at = NOW()
WHERE id = $1;
