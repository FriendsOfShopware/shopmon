-- Slack incoming webhooks at two levels. An organization webhook receives every
-- alert of the organization; a shop webhook receives that shop's alerts. Both
-- fire when both are set, so a team can keep one channel per shop next to an
-- organization-wide one. The URL is a bearer secret, so it is stored encrypted
-- with APP_SECRET like the shop client secrets.
CREATE TABLE "organization_slack_webhook" (
  "organization_id" text PRIMARY KEY NOT NULL REFERENCES "organization"("id") ON DELETE cascade,
  "webhook_url" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE TABLE "shop_slack_webhook" (
  "shop_id" integer PRIMARY KEY NOT NULL REFERENCES "shop"("id") ON DELETE cascade,
  "webhook_url" text NOT NULL,
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);
