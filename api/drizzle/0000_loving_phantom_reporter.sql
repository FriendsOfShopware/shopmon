CREATE TABLE `shop` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`team_id` integer NOT NULL,
	`name` text NOT NULL,
	`status` text DEFAULT 'green' NOT NULL,
	`url` text NOT NULL,
	`favicon` text,
	`client_id` text NOT NULL,
	`client_secret` text NOT NULL,
	`shopware_version` text NOT NULL,
	`last_scraped_at` text,
	`last_scraped_error` text,
	`ignores` text DEFAULT '[]',
	`shop_image` text,
	`last_updated` text DEFAULT '{}',
	`created_at` text NOT NULL,
	`updated_at` text,
	FOREIGN KEY (`team_id`) REFERENCES `team`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop_changelog` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`extensions` text NOT NULL,
	`old_shopware_version` text,
	`new_shopware_version` text,
	`date` text NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop_pagespeed` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`created_at` text NOT NULL,
	`performance` integer,
	`accessibility` integer,
	`best_practices` integer,
	`seo` integer,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop_scrape_info` (
	`shop_id` integer PRIMARY KEY NOT NULL,
	`extensions` text NOT NULL,
	`scheduled_task` text,
	`queue_info` text,
	`cache_info` text,
	`checks` text,
	`created_at` text NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `team` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`owner_id` integer NOT NULL,
	`created_at` text NOT NULL,
	`updated_at` text,
	FOREIGN KEY (`owner_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `user` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`username` text NOT NULL,
	`email` text NOT NULL,
	`password` text NOT NULL,
	`verify_code` text,
	`created_at` text NOT NULL,
	`updated_at` text
);
--> statement-breakpoint
CREATE TABLE `user_notification` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`user_id` integer NOT NULL,
	`key` text NOT NULL,
	`level` text NOT NULL,
	`title` text NOT NULL,
	`message` text NOT NULL,
	`link` text NOT NULL,
	`read` integer DEFAULT 0 NOT NULL,
	`created_at` text NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `user_to_team` (
	`user_id` integer NOT NULL,
	`team_id` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action,
	FOREIGN KEY (`team_id`) REFERENCES `team`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE UNIQUE INDEX `user_email_unique` ON `user` (`email`);