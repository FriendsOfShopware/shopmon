-- Shop extension table
CREATE TABLE `shop_extension` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`name` text NOT NULL,
	`label` text NOT NULL,
	`active` integer NOT NULL,
	`version` text NOT NULL,
	`latest_version` text,
	`installed` integer NOT NULL,
	`rating_average` integer,
	`store_link` text,
	`changelog` text,
	`installed_at` text,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `shop_extension_shop_id_name_unique` ON `shop_extension` (`shop_id`, `name`);
--> statement-breakpoint

-- Shop scheduled task table
CREATE TABLE `shop_scheduled_task` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`task_id` text NOT NULL,
	`name` text NOT NULL,
	`status` text NOT NULL,
	`interval` integer NOT NULL,
	`overdue` integer NOT NULL,
	`last_execution_time` text,
	`next_execution_time` text,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `shop_scheduled_task_shop_id_task_id_unique` ON `shop_scheduled_task` (`shop_id`, `task_id`);
--> statement-breakpoint

-- Shop queue table
CREATE TABLE `shop_queue` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`name` text NOT NULL,
	`size` integer NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `shop_queue_shop_id_name_unique` ON `shop_queue` (`shop_id`, `name`);
--> statement-breakpoint

-- Shop cache table
CREATE TABLE `shop_cache` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`environment` text NOT NULL,
	`http_cache` integer NOT NULL,
	`cache_adapter` text NOT NULL,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `shop_cache_shop_id_unique` ON `shop_cache` (`shop_id`);
--> statement-breakpoint

-- Shop check table
CREATE TABLE `shop_check` (
	`id` integer PRIMARY KEY AUTOINCREMENT NOT NULL,
	`shop_id` integer NOT NULL,
	`check_id` text NOT NULL,
	`level` text NOT NULL,
	`message` text NOT NULL,
	`source` text NOT NULL,
	`link` text,
	FOREIGN KEY (`shop_id`) REFERENCES `shop`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `shop_check_shop_id_check_id_unique` ON `shop_check` (`shop_id`, `check_id`);
