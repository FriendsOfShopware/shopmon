-- name: GetEnvironmentScrapeIssues :many
SELECT environment_id, issue_key, kind, dataset, message_key, params,
       consecutive_count, first_seen_at, last_seen_at, notified_at
FROM environment_scrape_issue WHERE environment_id = $1 ORDER BY issue_key;

-- name: UpsertEnvironmentScrapeIssue :one
-- Records one more consecutive sighting of an issue and returns the running
-- count plus whether owners were already told, which together decide whether
-- this sighting notifies (immediately for deterministic failures, after a
-- threshold for transient ones, never twice for the same occurrence).
INSERT INTO environment_scrape_issue (environment_id, issue_key, kind, dataset, message_key, params)
VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (environment_id, issue_key) DO UPDATE SET
  kind = EXCLUDED.kind,
  dataset = EXCLUDED.dataset,
  message_key = EXCLUDED.message_key,
  params = EXCLUDED.params,
  consecutive_count = environment_scrape_issue.consecutive_count + 1,
  last_seen_at = NOW()
RETURNING consecutive_count, notified_at;

-- name: MarkEnvironmentScrapeIssueNotified :exec
UPDATE environment_scrape_issue SET notified_at = NOW()
WHERE environment_id = $1 AND issue_key = $2 AND notified_at IS NULL;

-- name: DeleteEnvironmentScrapeIssuesByKeys :exec
DELETE FROM environment_scrape_issue WHERE environment_id = $1 AND issue_key = ANY($2::text[]);

