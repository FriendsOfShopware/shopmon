CREATE TABLE `invitation` (
	`id` text PRIMARY KEY NOT NULL,
	`organization_id` text NOT NULL,
	`email` text NOT NULL,
	`role` text,
	`status` text DEFAULT 'pending' NOT NULL,
	`expires_at` integer NOT NULL,
	`inviter_id` text NOT NULL,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`inviter_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `member` (
	`id` text PRIMARY KEY NOT NULL,
	`organization_id` text NOT NULL,
	`user_id` text NOT NULL,
	`role` text DEFAULT 'member' NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE cascade,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
INSERT INTO "member" ("id", "organization_id", "user_id", "role", "created_at") SELECT substr(concat('00000000000000000000000000000000'||"rowid"), -32, 32), "organization_id", "user_id", CASE WHEN((SELECT owner_id FROM organization WHERE id = organization_id) == user_id) THEN 'owner' ELSE 'member' END, date('now') FROM `user_to_organization`;
--> statement-breakpoint
DROP TABLE `user_to_organization`;--> statement-breakpoint
PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_organization` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text NOT NULL,
	`slug` text,
	`logo` text,
	`created_at` integer NOT NULL,
	`metadata` text
);
--> statement-breakpoint
WITH duplicates AS (
  SELECT 
    id,
    name,
    ROW_NUMBER() OVER (PARTITION BY name ORDER BY id) as row_num
  FROM organization
)
UPDATE organization
SET name = 
  CASE 
    WHEN (SELECT row_num FROM duplicates WHERE duplicates.id = organization.id) > 1
    THEN name || '-' || (SELECT row_num FROM duplicates WHERE duplicates.id = organization.id)
    ELSE name
  END
WHERE id IN (
  SELECT id 
  FROM duplicates 
  WHERE row_num > 1
);
--> statement-breakpoint
INSERT INTO `__new_organization`("id", "name", "slug", "logo", "created_at", "metadata") SELECT "id", "name", replace(replace(lower("name"), ' ', '-'), '!', ''), NULL, "created_at", NULL FROM `organization`;--> statement-breakpoint
DROP TABLE `organization`;--> statement-breakpoint
ALTER TABLE `__new_organization` RENAME TO `organization`;--> statement-breakpoint
UPDATE "member" SET "organization_id" = substr(concat('00000000000000000000000000000000'||"organization_id"), -32, 32);--> statement-breakpoint
UPDATE "organization" SET "id" = substr(concat('00000000000000000000000000000000'||"id"), -32, 32);--> statement-breakpoint
PRAGMA foreign_keys=ON;--> statement-breakpoint
CREATE UNIQUE INDEX `organization_slug_unique` ON `organization` (`slug`);--> statement-breakpoint
CREATE TABLE `__new_shop` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`organization_id` text NOT NULL,
	`name` text NOT NULL,
	`status` text DEFAULT 'green' NOT NULL,
	`url` text NOT NULL,
	`favicon` text,
	`client_id` text NOT NULL,
	`client_secret` text NOT NULL,
	`shopware_version` text NOT NULL,
	`last_scraped_at` integer,
	`last_scraped_error` text,
	`ignores` text DEFAULT '[]' NOT NULL,
	`shop_image` text,
	`last_changelog` text DEFAULT '{}',
	`created_at` integer NOT NULL,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_shop`("id", "organization_id", "name", "status", "url", "favicon", "client_id", "client_secret", "shopware_version", "last_scraped_at", "last_scraped_error", "ignores", "shop_image", "last_changelog", "created_at") SELECT "id", "organization_id", "name", "status", "url", "favicon", "client_id", "client_secret", "shopware_version", "last_scraped_at", "last_scraped_error", "ignores", "shop_image", "last_changelog", "created_at" FROM `shop`;--> statement-breakpoint
DROP TABLE `shop`;--> statement-breakpoint
ALTER TABLE `__new_shop` RENAME TO `shop`;--> statement-breakpoint
UPDATE "shop" SET "organization_id" = substr(concat('00000000000000000000000000000000'||"organization_id"), -32, 32);--> statement-breakpoint
ALTER TABLE `session` ADD `active_organization_id` text;
