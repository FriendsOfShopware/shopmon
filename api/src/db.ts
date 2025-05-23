import {
    sqliteTable,
    text,
    integer,
    primaryKey,
    unique,
} from 'drizzle-orm/sqlite-core';
import { drizzle as drizzleD1, DrizzleD1Database } from 'drizzle-orm/d1';
import { drizzle as drizzleLibSQL, LibSQLDatabase } from 'drizzle-orm/libsql';
import { createClient } from '@libsql/client';
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
import type { RegistrationJSON } from '@passwordless-id/webauthn/dist/esm/types';

type LastChangelog = {
    date: Date;
    from: string;
    to: string;
};

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
        .default([])
        .$type<string[]>()
        .notNull(),
    shopImage: text('shop_image'),
    lastChangelog: text('last_changelog', { mode: 'json' })
        .default('{}')
        .$type<LastChangelog>(),
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
    checks: text('checks', { mode: 'json' }).notNull().$type<CheckerChecks[]>(),
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

export const userNotification = sqliteTable(
    'user_notification',
    {
        id: integer('id').primaryKey({ autoIncrement: true }),
        userId: integer('user_id')
            .notNull()
            .references(() => user.id),
        key: text('key').notNull(),
        level: text('level').notNull(),
        title: text('title').notNull(),
        message: text('message').notNull(),
        link: text('link', { mode: 'json' })
            .notNull()
            .$type<NotificationLink>(),
        read: integer('read', { mode: 'boolean' }).notNull().default(false),
        createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    },
    (table) => {
        return {
            keyUnique: unique().on(table.userId, table.key),
        };
    },
);

export const userPasskeys = sqliteTable('user_passkeys', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    userId: integer('userId')
        .notNull()
        .references(() => user.id),
    key: text('key', { mode: 'json' }).notNull().$type<RegistrationJSON>(),
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

export type Drizzle = LibSQLDatabase<typeof schema>;

export function getConnection(env: Bindings) {
    if (
        typeof env.LIBSQL_URL === 'string' &&
        typeof env.LIBSQL_AUTH_TOKEN === 'string'
    ) {
        const client = createClient({
            url: env.LIBSQL_URL,
            authToken: env.LIBSQL_AUTH_TOKEN,
        });

        return drizzleLibSQL(client, { schema });
    }

    return drizzleD1(env.shopmonDB, { schema });
}

type ResultSet =
    | { lastInsertRowid: bigint | undefined }
    | { meta: { last_row_id: number } };

export function getLastInsertId(result: ResultSet): number {
    if ('lastInsertRowid' in result) {
        if (result.lastInsertRowid === undefined) {
            throw new Error('lastInsertRowid is undefined');
        }

        return new Number(result.lastInsertRowid).valueOf();
    }

    return result.meta.last_row_id;
}
