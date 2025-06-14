PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_shop` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`organization_id` text NOT NULL,
	`project_id` integer NOT NULL,
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
	`connection_issue_count` integer DEFAULT 0 NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE no action,
	FOREIGN KEY (`project_id`) REFERENCES `project`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_shop`("id", "organization_id", "project_id", "name", "status", "url", "favicon", "client_id", "client_secret", "shopware_version", "last_scraped_at", "last_scraped_error", "ignores", "shop_image", "last_changelog", "connection_issue_count", "created_at") SELECT "id", "organization_id", "project_id", "name", "status", "url", "favicon", "client_id", "client_secret", "shopware_version", "last_scraped_at", "last_scraped_error", "ignores", "shop_image", "last_changelog", "connection_issue_count", "created_at" FROM `shop`;--> statement-breakpoint
DROP TABLE `shop`;--> statement-breakpoint
ALTER TABLE `__new_shop` RENAME TO `shop`;--> statement-breakpoint
PRAGMA foreign_keys=ON;