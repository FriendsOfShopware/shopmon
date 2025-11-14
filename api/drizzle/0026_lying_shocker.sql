CREATE TABLE `deployment` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`command` text NOT NULL,
	`output` text NOT NULL,
	`return_code` integer NOT NULL,
	`start_date` integer NOT NULL,
	`end_date` integer NOT NULL,
	`execution_time` text NOT NULL,
	`composer` text DEFAULT '{}',
	`created_at` integer NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `deployment_token` (
	`id` text PRIMARY KEY NOT NULL,
	`shop_id` integer NOT NULL,
	`token` text NOT NULL,
	`name` text NOT NULL,
	`created_at` integer NOT NULL,
	`last_used_at` integer,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `deployment_token_token_unique` ON `deployment_token` (`token`);