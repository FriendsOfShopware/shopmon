import {
    sqliteTable,
    text,
    integer,
    primaryKey,
} from 'drizzle-orm/sqlite-core';
import { drizzle, DrizzleD1Database } from 'drizzle-orm/d1';
import type { Bindings } from './router';
import {
    CacheInfo,
    CheckerChecks,
    Extension,
    ExtensionDiff,
    NotificationLink,
    QueueInfo,
    ScheduledTask,
} from './types';
import type { RegistrationParsed } from '@passwordless-id/webauthn/src/types'

export const shop = sqliteTable('shop', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    organizationId: integer('organization_id')
        .notNull()
        .references(() => organization.id),
    name: text('name').notNull(),
    status: text('status').notNull().default('green'),
    url: text('url').notNull(),
    favicon: text('favicon'),
    clientId: text('client_id').notNull(),
    clientSecret: text('client_secret').notNull(),
    shopwareVersion: text('shopware_version').notNull(),
    lastScrapedAt: integer('last_scraped_at', { mode: 'timestamp' }),
    lastScrapedError: text('last_scraped_error'),
    ignores: text('ignores', { mode: 'json' })
        .default('[]')
        .$type<string[]>()
        .notNull(),
    shopImage: text('shop_image'),
    lastUpdated: integer('last_updated', { mode: 'timestamp' }),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const shopPageSpeed = sqliteTable('shop_pagespeed', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    shopId: integer('shop_id').references(() => shop.id),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    performance: integer('performance'),
    accessibility: integer('accessibility'),
    bestPractices: integer('best_practices'),
    seo: integer('seo'),
});

export const shopScrapeInfo = sqliteTable('shop_scrape_info', {
    shopId: integer('shop_id')
        .notNull()
        .primaryKey()
        .references(() => shop.id),
    extensions: text('extensions', { mode: 'json' })
        .notNull()
        .$type<Extension[]>(),
    scheduledTask: text('scheduled_task', { mode: 'json' })
        .notNull()
        .$type<ScheduledTask[]>(),
    queueInfo: text('queue_info', { mode: 'json' })
        .notNull()
        .$type<QueueInfo[]>(),
    cacheInfo: text('cache_info', { mode: 'json' })
        .notNull()
        .$type<CacheInfo>(),
    checks: text('checks').notNull().$type<CheckerChecks[]>(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const shopChangelog = sqliteTable('shop_changelog', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    shopId: integer('shop_id').references(() => shop.id),
    extensions: text('extensions', { mode: 'json' })
        .notNull()
        .$type<ExtensionDiff[]>(),
    oldShopwareVersion: text('old_shopware_version'),
    newShopwareVersion: text('new_shopware_version'),
    date: integer('date', { mode: 'timestamp' }).notNull(),
});

export const organization = sqliteTable('organization', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    name: text('name').notNull(),
    ownerId: integer('owner_id')
        .notNull()
        .references(() => user.id),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const user = sqliteTable('user', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    displayName: text('displayName').notNull(),
    email: text('email').notNull().unique(),
    password: text('password').notNull(),
    verifyCode: text('verify_code'),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const userNotification = sqliteTable('user_notification', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    userId: integer('user_id')
        .notNull()
        .references(() => user.id),
    key: text('key').notNull(),
    level: text('level').notNull(),
    title: text('title').notNull(),
    message: text('message').notNull(),
    link: text('link', { mode: 'json' }).notNull().$type<NotificationLink>(),
    read: integer('read', { mode: 'boolean' }).notNull().default(false),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const userPasskeys = sqliteTable('user_passkeys', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    userId: integer('userId')
        .notNull()
        .references(() => user.id),
    key: text('key', {mode: 'json'}).notNull().$type<RegistrationParsed>(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const userToOrganization = sqliteTable(
    'user_to_organization',
    {
        userId: integer('user_id')
            .notNull()
            .references(() => user.id),
        organizationId: integer('organization_id')
            .notNull()
            .references(() => organization.id),
    },
    (table) => {
        return {
            pk: primaryKey({ columns: [table.userId, table.organizationId] }),
        };
    },
);

export const schema = {
    user,
    shop,
    shopPageSpeed,
    shopScrapeInfo,
    shopChangelog,
    organization,
    userNotification,
    userToOrganization,
    userPasskeys,
};

export type Drizzle = DrizzleD1Database<typeof schema>;

export function getConnection(env: Bindings) {
    return drizzle(env.shopmonDB, { schema });
}
