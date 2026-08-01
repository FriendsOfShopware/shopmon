-- Which advisories an environment's subscribers have actually been alerted
-- about.
--
-- Deliberately NOT a column on environment_advisory_match: that table is
-- deleted and rebuilt wholesale on every rematch, so a marker living there
-- would be lost on each pass. Keeping it separate also decouples "this
-- advisory currently matches" from "we have already told someone", which is
-- what makes a failed dispatch retryable: the rematch commits its match rows
-- before notifying (they must persist even if the alert fails), so without
-- this table a transient failure in the notify path would leave the row
-- looking known forever and the alert would never be sent.
CREATE TABLE "environment_advisory_notified" (
  "environment_id" integer NOT NULL REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "notified_at" timestamp NOT NULL DEFAULT NOW(),
  PRIMARY KEY ("environment_id", "advisory_id")
);

-- Existing matches predate this table. Seed them as already notified so the
-- deployment that introduces it does not re-alert every subscriber about
-- every advisory their shops already match.
INSERT INTO environment_advisory_notified (environment_id, advisory_id, notified_at)
SELECT DISTINCT environment_id, advisory_id, NOW()
FROM environment_advisory_match
ON CONFLICT DO NOTHING;
