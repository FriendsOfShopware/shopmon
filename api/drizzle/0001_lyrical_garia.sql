CREATE TABLE `user_passkeys` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text NOT NULL,
	`userId` integer NOT NULL,
	`key` text NOT NULL,
	`created_at` integer NOT NULL,
	FOREIGN KEY (`userId`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE no action
);
