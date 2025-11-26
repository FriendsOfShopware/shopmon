CREATE TABLE `project_api_key` (
	`id` text PRIMARY KEY NOT NULL,
	`project_id` integer NOT NULL,
	`name` text NOT NULL,
	`token` text NOT NULL,
	`scopes` text NOT NULL,
	`created_at` integer NOT NULL,
	`last_used_at` integer,
	FOREIGN KEY (`project_id`) REFERENCES `project`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `project_api_key_token_unique` ON `project_api_key` (`token`);
--> statement-breakpoint
DROP TABLE IF EXISTS `deployment_token`;
