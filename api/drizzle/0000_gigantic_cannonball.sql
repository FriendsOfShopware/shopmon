CREATE TABLE `organization` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`name` text NOT NULL,
	`owner_id` integer NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`owner_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`organization_id` integer NOT NULL,
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
	`last_updated` integer,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop_changelog` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`extensions` text NOT NULL,
	`old_shopware_version` text,
	`new_shopware_version` text,
	`date` integer NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `shop_pagespeed` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`created_at` integer NOT NULL,
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
	`scheduled_task` text NOT NULL,
	`queue_info` text NOT NULL,
	`cache_info` text NOT NULL,
	`checks` text NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `user` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`displayName` text NOT NULL,
	`email` text NOT NULL,
	`password` text NOT NULL,
	`verify_code` text,
	`created_at` integer NOT NULL
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
	`read` integer DEFAULT false NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE TABLE `user_to_organization` (
	`user_id` integer NOT NULL,
	`organization_id` integer NOT NULL,
	PRIMARY KEY(`organization_id`, `user_id`),
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
CREATE UNIQUE INDEX `user_email_unique` ON `user` (`email`);
