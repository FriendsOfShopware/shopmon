-- Shop-owner acknowledgement that a known advisory is accepted or mitigated
-- outside Shopmon (WAF rule, unused component, compensating control).
--
-- Deliberately separate from environment.ignores: that column is a flat array
-- of check ids with no room for a reason, an actor, an expiry, or a shop-wide
-- scope, and it is rewritten wholesale by the generic environment PATCH. Both
-- mechanisms coexist; they are unioned at the point of use.
CREATE TABLE "advisory_suppression" (
  "id" bigserial PRIMARY KEY NOT NULL,
  "organization_id" text NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "shop_id" integer NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  -- NULL means every environment of the shop; set narrows to one environment
  -- (e.g. "accepted on staging, still alert on production").
  "environment_id" integer REFERENCES "environment"("id") ON DELETE cascade,
  "advisory_id" text NOT NULL REFERENCES "composer_advisory"("advisory_id") ON DELETE cascade,
  "reason" text NOT NULL,
  "expires_at" timestamp,
  "created_by" text REFERENCES "user"("id") ON DELETE SET NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  -- Revoking is a soft delete: created_by/created_at/reason/revoked_* are the
  -- audit trail, so rows are never removed.
  "revoked_at" timestamp,
  "revoked_by" text REFERENCES "user"("id") ON DELETE SET NULL
);

-- Two partial uniques rather than one: NULLs are never equal in a unique
-- index, so a single constraint would let duplicate shop-wide rows through.
CREATE UNIQUE INDEX idx_advisory_suppression_shop
  ON advisory_suppression (shop_id, advisory_id)
  WHERE revoked_at IS NULL AND environment_id IS NULL;
CREATE UNIQUE INDEX idx_advisory_suppression_env
  ON advisory_suppression (environment_id, advisory_id)
  WHERE revoked_at IS NULL AND environment_id IS NOT NULL;

CREATE INDEX idx_advisory_suppression_advisory
  ON advisory_suppression (advisory_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_advisory_suppression_lookup
  ON advisory_suppression (shop_id, environment_id) WHERE revoked_at IS NULL;
CREATE INDEX idx_advisory_suppression_organization
  ON advisory_suppression (organization_id) WHERE revoked_at IS NULL;
