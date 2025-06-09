CREATE TABLE `shop_sitespeed` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer,
	`created_at` integer NOT NULL,
	`ttfb` integer,
	`fully_loaded` integer,
	`largest_contentful_paint` integer,
	`first_contentful_paint` integer,
	`cumulative_layout_shift` integer,
	`speed_index` integer,
	`transfer_size` integer,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE no action
);
