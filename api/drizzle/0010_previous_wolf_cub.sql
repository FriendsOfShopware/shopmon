PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_user_to_organization` (
	`user_id` text NOT NULL,
	`organization_id` integer NOT NULL,
	PRIMARY KEY(`user_id`, `organization_id`),
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action,
	FOREIGN KEY (`organization_id`) REFERENCES `organization`(`id`) ON UPDATE no action ON DELETE no action
);
--> statement-breakpoint
INSERT INTO `__new_user_to_organization`("user_id", "organization_id") SELECT substr(concat('00000000000000000000000000000000'||"user_id"), -32, 32), "organization_id" FROM `user_to_organization`;--> statement-breakpoint
DROP TABLE `user_to_organization`;--> statement-breakpoint
ALTER TABLE `__new_user_to_organization` RENAME TO `user_to_organization`;--> statement-breakpoint
PRAGMA foreign_keys=ON;
