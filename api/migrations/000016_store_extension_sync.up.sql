-- store_extension_sync records when the store API was last asked about an
-- extension name, including names the store does not know (so custom plugins
-- are not re-checked on every scrape). It gates the shared catalog refresh.
CREATE TABLE "store_extension_sync" (
  "extension_name" text PRIMARY KEY NOT NULL,
  "last_synced_at" timestamp NOT NULL DEFAULT NOW()
);

-- store_extension_compatibility stores, per Shopware version in use by any
-- environment, the latest extension release the store reports as compatible.
-- Environment scrapes read their compatible latest version from here instead
-- of asking the store API per environment.
CREATE TABLE "store_extension_compatibility" (
  "extension_name" text NOT NULL REFERENCES "store_extension"("name") ON DELETE cascade,
  "shopware_version" text NOT NULL,
  "latest_version" text,
  PRIMARY KEY ("extension_name", "shopware_version")
);
