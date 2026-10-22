-- Shopmon-authored security advisories: admins can publish an advisory before
-- an official CVE/GHSA exists (registries are slow), and link it to the
-- official identifier later. Such rows live in the same composer_advisory
-- catalog so SBOM matching, suppressions, checks, and alerts apply unchanged;
-- "origin" is what marks them as Shopmon reports in the UI.

-- Numbering for Shopmon advisory ids ("SHOPMON-2026-0001"). Global on purpose:
-- the year prefix is the creation year, the number never restarts, so an id
-- can never be reissued for a different vulnerability.
CREATE SEQUENCE shopmon_advisory_seq;

ALTER TABLE "composer_advisory"
  ADD COLUMN "origin" text NOT NULL DEFAULT 'upstream',
  ADD COLUMN "created_by" text REFERENCES "user"("id") ON DELETE SET NULL;

ALTER TABLE "composer_advisory"
  ADD CONSTRAINT composer_advisory_origin_check CHECK (origin IN ('upstream', 'shopmon'));

-- Backs the admin advisory list's origin filter. The user-facing lists filter
-- on is_visible first and are already served by the visible/reported index.
CREATE INDEX idx_composer_advisory_origin ON composer_advisory (origin);

-- Shopmon-authored package rows have no Packagist PKSA id. The existing unique
-- index tolerates any number of NULLs, so only the NOT NULL has to go.
ALTER TABLE "composer_advisory_package"
  ALTER COLUMN "packagist_advisory_id" DROP NOT NULL;
