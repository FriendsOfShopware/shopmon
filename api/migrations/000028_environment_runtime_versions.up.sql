-- Exact runtime versions reported by FroshTools health checks so Shopmon can
-- surface and alert on PHP and MySQL/MariaDB versions.
ALTER TABLE "environment"
  ADD COLUMN "php_version" text,
  ADD COLUMN "mysql_version" text;

UPDATE "environment" e
SET "php_version" = NULLIF(BTRIM(ec."params"->>'current'), '')
FROM "environment_check" ec
WHERE ec."environment_id" = e."id"
  AND ec."check_id" IN ('frosh.PHP Version', 'frosh.php-version', 'frosh.phpGood', 'frosh.phpOutdated')
  AND ec."params" ? 'current'
  AND BTRIM(ec."params"->>'current') <> ''
  AND LOWER(BTRIM(ec."params"->>'current')) <> 'unknown';

UPDATE "environment" e
SET "mysql_version" = NULLIF(BTRIM(ec."params"->>'current'), '')
FROM "environment_check" ec
WHERE ec."environment_id" = e."id"
  AND ec."check_id" IN ('frosh.MySQL Version', 'frosh.MariaDB Version', 'frosh.mysql', 'frosh.mysqlGood', 'frosh.mysqlOutdated')
  AND ec."params" ? 'current'
  AND BTRIM(ec."params"->>'current') <> ''
  AND LOWER(BTRIM(ec."params"->>'current')) <> 'unknown';
