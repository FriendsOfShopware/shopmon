import { Database } from 'bun:sqlite';
import {
    type BunSQLiteDatabase,
    drizzle as drizzleSqlite,
} from 'drizzle-orm/bun-sqlite';
import {
    integer,
    primaryKey,
    sqliteTable,
    text,
    unique,
} from 'drizzle-orm/sqlite-core';
import type {
    CacheInfo,
    CheckerChecks,
    Extension,
    ExtensionDiff,
    NotificationLink,
    QueueInfo,
    ScheduledTask,
} from './types/index.ts';

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
    ownerId: text('owner_id')
        .notNull()
        .references(() => user.id),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const userNotification = sqliteTable(
    'user_notification',
    {
        id: integer('id').primaryKey({ autoIncrement: true }),
        userId: text('user_id')
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

export const user = sqliteTable('user', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    email: text('email').notNull().unique(),
    emailVerified: integer('email_verified', { mode: 'boolean' })
        .$defaultFn(() => !1)
        .notNull(),
    image: text('image'),
    createdAt: integer('created_at', { mode: 'timestamp' })
        .$defaultFn(() => new Date())
        .notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' })
        .$defaultFn(() => new Date())
        .notNull(),
});

export const session = sqliteTable('session', {
    id: text('id').primaryKey(),
    expiresAt: integer('expires_at', { mode: 'timestamp' }).notNull(),
    token: text('token').notNull().unique(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' }).notNull(),
    ipAddress: text('ip_address'),
    userAgent: text('user_agent'),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
});

export const account = sqliteTable('account', {
    id: text('id').primaryKey(),
    accountId: text('account_id').notNull(),
    providerId: text('provider_id').notNull(),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    accessToken: text('access_token'),
    refreshToken: text('refresh_token'),
    idToken: text('id_token'),
    accessTokenExpiresAt: integer('access_token_expires_at', {
        mode: 'timestamp',
    }),
    refreshTokenExpiresAt: integer('refresh_token_expires_at', {
        mode: 'timestamp',
    }),
    scope: text('scope'),
    password: text('password'),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' }).notNull(),
});

export const verification = sqliteTable('verification', {
    id: text('id').primaryKey(),
    identifier: text('identifier').notNull(),
    value: text('value').notNull(),
    expiresAt: integer('expires_at', { mode: 'timestamp' }).notNull(),
    createdAt: integer('created_at', { mode: 'timestamp' }).$defaultFn(
        () => new Date(),
    ),
    updatedAt: integer('updated_at', { mode: 'timestamp' }).$defaultFn(
        () => new Date(),
    ),
});

export const passkey = sqliteTable('passkey', {
    id: text('id').primaryKey(),
    name: text('name'),
    publicKey: text('public_key').notNull(),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    credentialID: text('credential_i_d').notNull(),
    counter: integer('counter').notNull(),
    deviceType: text('device_type').notNull(),
    backedUp: integer('backed_up', { mode: 'boolean' }).notNull(),
    transports: text('transports'),
    createdAt: integer('created_at', { mode: 'timestamp' }),
});

export const userToOrganization = sqliteTable(
    'user_to_organization',
    {
        userId: text('user_id')
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
    shop,
    shopPageSpeed,
    shopScrapeInfo,
    shopChangelog,
    organization,
    userNotification,
    userToOrganization,

    // Better Auth
    user,
    session,
    account,
    verification,
    passkey,
};

export type Drizzle = BunSQLiteDatabase<typeof schema>;
let drizzle: Drizzle | undefined = undefined;
let dbClient: Database | undefined = undefined;

export function getConnection(applyPragmas = true) {
    if (drizzle !== undefined) {
        return drizzle;
    }

    const dbPath = process.env.APP_DATABASE_PATH || 'shopmon.db';
    dbClient = new Database(dbPath);

    if (applyPragmas) {
        // Enable Write-Ahead Logging for better concurrency
        dbClient.exec('PRAGMA journal_mode = WAL');
        // Increase cache size (negative value = KB, so -64000 = 64MB)
        dbClient.exec('PRAGMA cache_size = -64000');
        // Enable foreign key constraints
        dbClient.exec('PRAGMA foreign_keys = ON');
        // Synchronous mode - NORMAL is safe and faster than FULL
        dbClient.exec('PRAGMA synchronous = NORMAL');
        // Temp store in memory for better performance
        dbClient.exec('PRAGMA temp_store = MEMORY');
        // Increase busy timeout to 5 seconds
        dbClient.exec('PRAGMA busy_timeout = 5000');
        // Enable query optimizer
        dbClient.exec('PRAGMA optimize');
    }

    drizzle = drizzleSqlite(dbClient, { schema });

    return drizzle;
}

export function closeConnection() {
    if (dbClient) {
        try {
            // Run optimize before closing
            dbClient.exec('PRAGMA optimize');
            dbClient.close();
            dbClient = undefined;
            drizzle = undefined;
            console.log('Database connection closed gracefully');
        } catch (error) {
            console.error('Error closing database connection:', error);
        }
    }
}
