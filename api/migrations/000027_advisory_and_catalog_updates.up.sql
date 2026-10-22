-- Retain the Store's current release snapshot separately from historical
-- changelogs so withdrawn releases are excluded from update suggestions.
ALTER TABLE store_extension ADD COLUMN available_versions text[];

-- Manually tracked advisories can be authored in Shopmon or imported from
-- GitHub before Packagist picks them up. They share the existing catalog so
-- inventory matching, suppressions, checks, and alerts apply unchanged.
-- Shopmon reports receive SHOPMON-YYYY-NNNN ids; GitHub imports keep GHSA ids.
-- The sequence is global and never resets, so ids cannot be reissued.
CREATE SEQUENCE shopmon_advisory_seq;

ALTER TABLE composer_advisory
  ADD COLUMN origin text NOT NULL DEFAULT 'upstream',
  ADD COLUMN created_by text REFERENCES "user"(id) ON DELETE SET NULL,
  ADD CONSTRAINT composer_advisory_origin_check
    CHECK (origin IN ('upstream', 'shopmon', 'github'));

CREATE INDEX idx_composer_advisory_origin ON composer_advisory (origin);

-- Manual package rows have no Packagist PKSA id; the existing unique index
-- already permits multiple NULLs.
ALTER TABLE composer_advisory_package
  ALTER COLUMN packagist_advisory_id DROP NOT NULL;
