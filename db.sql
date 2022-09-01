CREATE TABLE `shop` (
	`id` int NOT NULL AUTO_INCREMENT,
	`team_id` int NOT NULL,
	`name` varchar(255) NOT NULL,
	`url` varchar(255) NOT NULL,
	`client_id` varchar(255) NOT NULL,
	`client_secret` varchar(255) NOT NULL,
	`shopware_version` varchar(255),
	`last_scraped_at` datetime,
	`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	`updated_at` datetime,
	PRIMARY KEY (`id`),
	UNIQUE KEY `url` (`url`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;

  CREATE TABLE `shop_scrape_info` (
	`shop_id` int NOT NULL,
	`extensions` text NOT NULL,
	`scheduled_task` text NOT NULL,
	`queue_info` text,
	`cache_info` text,
	`created_at` datetime NOT NULL,
	PRIMARY KEY (`shop_id`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;

  CREATE TABLE `team` (
	`id` int NOT NULL AUTO_INCREMENT,
	`name` varchar(255),
	`owner_id` int NOT NULL,
	`created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP(),
	`updated_at` datetime,
	PRIMARY KEY (`id`),
	UNIQUE KEY `name` (`name`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;

  CREATE TABLE `user` (
	`id` int NOT NULL AUTO_INCREMENT,
	`username` varchar(50),
	`email` varchar(255) NOT NULL,
	`password` varchar(255) NOT NULL,
	`verify_code` varchar(32),
	`created_at` datetime DEFAULT CURRENT_TIMESTAMP(),
	`updated_at` datetime,
	PRIMARY KEY (`id`),
	UNIQUE KEY `email` (`email`),
	UNIQUE KEY `verify_code` (`verify_code`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;

  CREATE TABLE `user_to_team` (
	`user_id` int NOT NULL,
	`team_id` int NOT NULL,
	UNIQUE KEY `user_id_team_id` (`user_id`, `team_id`)
) ENGINE InnoDB,
  CHARSET utf8mb4,
  COLLATE utf8mb4_0900_ai_ci;