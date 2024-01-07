import { sqliteTable, text, integer } from "drizzle-orm/sqlite-core";
import { drizzle, DrizzleD1Database } from 'drizzle-orm/d1';

export const shop = sqliteTable('shop', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    team_id: integer('team_id').notNull().references(() => team.id),
    name: text('name').notNull(),
    status: text('status').notNull().default('green'),
    url: text('url').notNull(),
    favicon: text('favicon'),
    client_id: text('client_id').notNull(),
    client_secret: text('client_secret').notNull(),
    shopware_version: text('shopware_version').notNull(),
    last_scraped_at: text('last_scraped_at'),
    last_scraped_error: text('last_scraped_error'),
    ignores: text('ignores').default('[]'),
    shop_image: text('shop_image'),
    last_updated: text('last_updated').default('{}'),
    created_at: text('created_at').notNull(),
    updated_at: text('updated_at'),
});

export const shopPageSpeed = sqliteTable('shop_pagespeed', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    shop_id: integer('shop_id').references(() => shop.id),
    created_at: text('created_at').notNull(),
    performance: integer('performance'),
    accessibility: integer('accessibility'),
    best_practices: integer('best_practices'),
    seo: integer('seo'),
});

export const shopScrapeInfo = sqliteTable('shop_scrape_info', {
    shop: integer('shop_id').notNull().primaryKey().references(() => shop.id),
    extensions: text('extensions').notNull(),
    scheduled_task: text('scheduled_task'),
    queue_info: text('queue_info'),
    cache_info: text('cache_info'),
    checks: text('checks'),
    created_at: text('created_at').notNull(),
});

export const shopChangelog = sqliteTable('shop_changelog', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    shop_id: integer('shop_id').references(() => shop.id),
    extensions: text('extensions').notNull(),
    old_shopware_version: text('old_shopware_version'),
    new_shopware_version: text('new_shopware_version'),
    date: text('date').notNull(),
});

export const team = sqliteTable('team', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    name: text('name').notNull(),
    owner_id: integer('owner_id').notNull().references(() => user.id),
    created_at: text('created_at').notNull(),
    updated_at: text('updated_at'),
});

export const user = sqliteTable('user', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    username: text('username').notNull(),
    email: text('email').notNull().unique(),
    password: text('password').notNull(),
    verify_code: text('verify_code'),
    created_at: text('created_at').notNull(),
    updated_at: text('updated_at'),
});

export const userNotification = sqliteTable('user_notification', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    user_id: integer('user_id').notNull().references(() => user.id),
    key: text('key').notNull(),
    level: text('level').notNull(),
    title: text('title').notNull(),
    message: text('message').notNull(),
    link: text('link').notNull(),
    read: integer('read').notNull().default(0),
    created_at: text('created_at').notNull(),
});

export const userToTeam = sqliteTable('user_to_team', {
    user_id: integer('user_id').notNull().references(() => user.id),
    team_id: integer('team_id').notNull().references(() => team.id),
});

export const schema = { user, shop, shopPageSpeed, shopScrapeInfo, shopChangelog, team, userNotification, userToTeam }

export type Drizzle = DrizzleD1Database<typeof schema>

export function getConnection(env: Env) {
    return drizzle(env.shopmonDB, { schema })
}
