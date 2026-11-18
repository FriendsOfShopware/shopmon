CREATE TABLE security_plugin_verified_fix (
  ghsa_id text NOT NULL,
  plugin_branch text NOT NULL,
  plugin_version text NOT NULL,
  shopware_branch text NOT NULL,
  release_url text NOT NULL,
  PRIMARY KEY (ghsa_id, plugin_branch)
);

CREATE VIEW security_plugin_effective_fix AS
SELECT ghsa_id, plugin_branch, plugin_version, shopware_branch
FROM security_plugin_verified_fix
UNION ALL
SELECT f.ghsa_id, f.plugin_branch, f.plugin_version, f.shopware_branch
FROM security_plugin_fix f
WHERE NOT EXISTS (
  SELECT 1 FROM security_plugin_verified_fix v
  WHERE v.ghsa_id = f.ghsa_id AND v.plugin_branch = f.plugin_branch
);

-- These releases contain fixes omitted from the Store changelog. See
-- CONTRIBUTING.md for how to verify the mappings in a Store installation.
WITH batch(ghsa_id, only_shopware_67) AS (
  VALUES
    ('GHSA-23G6-M65M-5QCG', true),
    ('GHSA-28FX-G7QC-7C24', false),
    ('GHSA-3Q8X-6FC9-5428', false),
    ('GHSA-4FRP-8844-M589', false),
    ('GHSA-4V2W-JPMG-WP82', false),
    ('GHSA-4V6J-JH9J-PMCH', false),
    ('GHSA-59J5-8MV2-2FC5', false),
    ('GHSA-7FCF-WRCG-GCGF', false),
    ('GHSA-CXVC-PQR3-G5CQ', false),
    ('GHSA-FMVH-847R-3HMQ', false),
    ('GHSA-G8R6-CXW8-F75W', true),
    ('GHSA-PHG9-8V2J-G4C4', false),
    ('GHSA-X6MG-FRX2-6P8J', false)
), releases(plugin_branch, plugin_version, shopware_branch) AS (
  VALUES
    ('2', '2.0.27', '6.5'),
    ('3', '3.0.21', '6.6'),
    ('4', '4.0.17', '6.7')
)
INSERT INTO security_plugin_verified_fix (ghsa_id, plugin_branch, plugin_version, shopware_branch, release_url)
SELECT b.ghsa_id, r.plugin_branch, r.plugin_version, r.shopware_branch,
       'https://packages.shopware.com/feeds/package/store.shopware.com/swagplatformsecurity.json'
FROM batch b CROSS JOIN releases r
WHERE r.plugin_branch = '4' OR NOT b.only_shopware_67;
