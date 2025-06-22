CREATE TABLE `shop_sitespeed` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`created_at` integer NOT NULL,
	`ttfb` integer,
	`fully_loaded` integer,
	`largest_contentful_paint` integer,
	`first_contentful_paint` integer,
	`cumulative_layout_shift` integer,
	`transfer_size` integer,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
DROP TABLE `shop_pagespeed`;--> statement-breakpoint
ALTER TABLE `shop` ADD `sitespeed_enabled` integer DEFAULT false NOT NULL;--> statement-breakpoint
ALTER TABLE `shop` ADD `sitespeed_urls` text DEFAULT '[]' NOT NULL;