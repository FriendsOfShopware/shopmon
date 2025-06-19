CREATE TABLE `account` (
	`id` text PRIMARY KEY NOT NULL,
	`account_id` text NOT NULL,
	`provider_id` text NOT NULL,
	`user_id` text NOT NULL,
	`access_token` text,
	`refresh_token` text,
	`id_token` text,
	`access_token_expires_at` integer,
	`refresh_token_expires_at` integer,
	`scope` text,
	`password` text,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `passkey` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text,
	`public_key` text NOT NULL,
	`user_id` text NOT NULL,
	`credential_i_d` text NOT NULL,
	`counter` integer NOT NULL,
	`device_type` text NOT NULL,
	`backed_up` integer NOT NULL,
	`transports` text,
	`created_at` integer,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE TABLE `session` (
	`id` text PRIMARY KEY NOT NULL,
	`expires_at` integer NOT NULL,
	`token` text NOT NULL,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	`ip_address` text,
	`user_agent` text,
	`user_id` text NOT NULL,
	FOREIGN KEY (`user_id`) REFERENCES `user`(`id`) ON UPDATE no action ON DELETE cascade
);
--> statement-breakpoint
CREATE UNIQUE INDEX `session_token_unique` ON `session` (`token`);--> statement-breakpoint
CREATE TABLE `verification` (
	`id` text PRIMARY KEY NOT NULL,
	`identifier` text NOT NULL,
	`value` text NOT NULL,
	`expires_at` integer NOT NULL,
	`created_at` integer,
	`updated_at` integer
);
--> statement-breakpoint
DROP TABLE `password_reset_tokens`;--> statement-breakpoint
DROP TABLE `sessions`;--> statement-breakpoint
INSERT INTO `passkey` ("id", "name", "public_key", "user_id", "credential_i_d", "counter", "device_type", "backed_up", "transports", "created_at") 
SELECT 
    substr(concat('00000000000000000000000000000000'||"id"), -32, 32), 
    "name", 
    JSON_EXTRACT("key", '$.credential.publicKey'), 
    substr(concat('00000000000000000000000000000000'||"userId"), -32, 32), 
    JSON_EXTRACT("key", '$.credential.id'), 
    JSON_EXTRACT("key", '$.authenticator.counter'), 
    'multiDevice', 
    IFNULL(JSON_EXTRACT("key", '$.authenticator.flags.backupState'), 0), 
    'internal,hybrid', 
    "created_at"  
FROM user_passkeys;
--> statement-breakpoint

DROP TABLE `user_passkeys`;--> statement-breakpoint
PRAGMA foreign_keys=OFF;--> statement-breakpoint
CREATE TABLE `__new_user` (
	`id` text PRIMARY KEY NOT NULL,
	`name` text NOT NULL,
	`email` text NOT NULL,
	`email_verified` integer NOT NULL,
	`image` text,
	`created_at` integer NOT NULL,
	`updated_at` integer NOT NULL,
	`display_name` text
);
--> statement-breakpoint
INSERT INTO `__new_user`("id", "name", "email", "email_verified", "image", "created_at", "updated_at", "display_name") SELECT substr(concat('00000000000000000000000000000000'||"id"), -32, 32), "displayName", "email", 1, NULL, "created_at", "created_at", "displayName" FROM `user`;--> statement-breakpoint
INSERT INTO `account` ("id", "account_id", "provider_id", "user_id", "password", "created_at", "updated_at") SELECT substr(concat('00000000000000000000000000000000'||"id"), -32, 32), substr(concat('00000000000000000000000000000000'||"id"), -32, 32), 'credential', substr(concat('00000000000000000000000000000000'||"id"), -32, 32), "password", "created_at", "created_at" FROM `user`;--> statement-breakpoint
DROP TABLE `user`;--> statement-breakpoint
ALTER TABLE `__new_user` RENAME TO `user`;--> statement-breakpoint
PRAGMA foreign_keys=ON;--> statement-breakpoint
CREATE UNIQUE INDEX `user_email_unique` ON `user` (`email`);
