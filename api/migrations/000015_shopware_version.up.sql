-- Local cache of Shopware releases crawled hourly from
-- https://releases.shopware.com/changelog so the API never has to call an
-- external service to answer "which Shopware versions exist" (previously fetched
-- live from the shopware-static-data GitHub repo).
CREATE TABLE "shopware_version" (
  "version" text PRIMARY KEY NOT NULL,
  "release_date" timestamp NOT NULL,
  "php_versions" jsonb NOT NULL DEFAULT '[]'::jsonb,
  "title" text NOT NULL DEFAULT '',
  "body" text NOT NULL DEFAULT '',
  "created_at" timestamp NOT NULL DEFAULT NOW(),
  "updated_at" timestamp NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_shopware_version_release_date ON shopware_version (release_date);
