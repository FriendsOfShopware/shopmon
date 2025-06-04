CREATE TABLE `lock` (
	`key` text PRIMARY KEY NOT NULL,
	`expires` integer NOT NULL,
	`created_at` integer NOT NULL
);
