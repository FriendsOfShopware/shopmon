CREATE TABLE `shop` (
                        `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                        `team_id` INTEGER NOT NULL,
                        `name` TEXT NOT NULL,
                        `status` TEXT NOT NULL DEFAULT 'green',
                        `url` TEXT NOT NULL,
                        `favicon` TEXT,
                        `client_id` TEXT NOT NULL,
                        `client_secret` TEXT NOT NULL,
                        `shopware_version` TEXT,
                        `last_scraped_at` TEXT,
                        `last_scraped_error` text,
                        `ignores` text default '[]' null,
                        `shop_image` TEXT NULL DEFAULT NULL,
                        `created_at` TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TEXT,
                        UNIQUE (`url`)
);

CREATE TABLE `shop_pagespeed` (
                                  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                                  `shop_id` INTEGER NOT NULL DEFAULT '0',
                                  `created_at` TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                                  `performance` INTEGER NOT NULL DEFAULT '0',
                                  `accessibility` INTEGER NOT NULL DEFAULT '0',
                                  `bestpractices` INTEGER NOT NULL DEFAULT '0',
                                  `seo` INTEGER NOT NULL DEFAULT '0'
);

CREATE TABLE `shop_scrape_info` (
                                    `shop_id` INTEGER NOT NULL,
                                    `extensions` text NOT NULL,
                                    `scheduled_task` text NOT NULL,
                                    `queue_info` text NOT NULL,
                                    `cache_info` text NOT NULL,
                                    `checks` text,
                                    `created_at` TEXT NOT NULL
);

CREATE TABLE `team` (
                        `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                        `name` TEXT,
                        `owner_id` INTEGER NOT NULL,
                        `created_at` TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TEXT,
                        UNIQUE (`name`)
);

CREATE TABLE `user` (
                        `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                        `username` TEXT,
                        `email` TEXT NOT NULL,
                        `password` TEXT NOT NULL,
                        `verify_code` TEXT,
                        `created_at` TEXT DEFAULT CURRENT_TIMESTAMP,
                        `updated_at` TEXT,
                        UNIQUE (`email`),
                        UNIQUE (`verify_code`)
);

CREATE TABLE `user_notification` (
                                     `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                                     `user_id` INTEGER NOT NULL,
                                     `key` TEXT NOT NULL,
                                     `level` TEXT NOT NULL,
                                     `title` TEXT NOT NULL DEFAULT '',
                                     `message` TEXT NOT NULL,
                                     `link` TEXT NOT NULL,
                                     `read` INTEGER NOT NULL DEFAULT '0',
                                     `created_at` TEXT DEFAULT CURRENT_TIMESTAMP,
                                     UNIQUE (`user_id`, `key`)
);

CREATE TABLE `user_to_team` (
                                `user_id` INTEGER NOT NULL,
                                `team_id` INTEGER NOT NULL,
                                UNIQUE (`user_id`, `team_id`)
);

CREATE TABLE `shop_changelog` (
                                  `id` INTEGER PRIMARY KEY AUTOINCREMENT,
                                  `shop_id` INTEGER NOT NULL,
                                  `extensions` text NULL,
                                  `old_shopware_version` TEXT NULL,
                                  `new_shopware_version` TEXT NULL,
                                  `date` TEXT NULL
);
