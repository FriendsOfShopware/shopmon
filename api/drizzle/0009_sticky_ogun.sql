PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_organization` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`owner_id` text NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`owner_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_organization`("id", "name", "owner_id", "created_at") SELECT "id", "name", substr(concat('00000000000000000000000000000000'||"owner_id"), -32, 32), "created_at" FROM `organization`;--> statement-breakpoint
DROP TABLE `organization`;--> statement-breakpoint
ALTER TABLE `__new_organization` RENAME TO `organization`;--> statement-breakpoint
PRAGMA foreign_keys=ON;--> statement-breakpoint
CREATE TABLE `__new_user_notification` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` text NOT NULL,
	`key` text NOT NULL,
	`level` text NOT NULL,
	`title` text NOT NULL,
	`message` text NOT NULL,
	`link` text NOT NULL,
	`read` integer DEFAULT false NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_user_notification`("id", "user_id", "key", "level", "title", "message", "link", "read", "created_at") SELECT "id", substr(concat('00000000000000000000000000000000'||"user_id"), -32, 32), "key", "level", "title", "message", "link", "read", "created_at" FROM `user_notification`;--> statement-breakpoint
DROP TABLE `user_notification`;--> statement-breakpoint
ALTER TABLE `__new_user_notification` RENAME TO `user_notification`;--> statement-breakpoint
CREATE UNIQUE INDEX `user_notification_user_id_key_unique` ON `user_notification` (`user_id`,`key`);
