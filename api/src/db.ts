import { createClient } from '@libsql/client';
import { relations } from 'drizzle-orm';
import {
    drizzle as drizzleSqlite,
    type LibSQLDatabase,
} from 'drizzle-orm/libsql';
import { integer, sqliteTable, text, unique } from 'drizzle-orm/sqlite-core';
import type { ExtensionDiff, NotificationLink } from './types/index.ts';

type LastChangelog = {
    date: Date;
    from: string;
    to: string;
};

export const organization = sqliteTable('organization', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    slug: text('slug').unique(),
    logo: text('logo'),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    metadata: text('metadata'),
});

export const project = sqliteTable('project', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    name: text('name').notNull(),
    description: text('description'),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    updatedAt: integer('updated_at', { mode: 'timestamp' }).notNull(),
});

export const shop = sqliteTable('shop', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id),
    projectId: integer('project_id')
        .notNull()
        .references(() => project.id),
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
    connectionIssueCount: integer('connection_issue_count')
        .default(0)
        .notNull(),
    sitespeedEnabled: integer('sitespeed_enabled', { mode: 'boolean' })
        .default(false)
        .notNull(),
    sitespeedUrls: text('sitespeed_urls', { mode: 'json' })
        .default([])
        .$type<string[]>()
        .notNull(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const shopSitespeed = sqliteTable('shop_sitespeed', {
    id: integer('id').primaryKey({ autoIncrement: true }),
    shopId: integer('shop_id').references(() => shop.id),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
    ttfb: integer('ttfb'),
    fullyLoaded: integer('fully_loaded'),
    largestContentfulPaint: integer('largest_contentful_paint'),
    firstContentfulPaint: integer('first_contentful_paint'),
    cumulativeLayoutShift: integer('cumulative_layout_shift'),
    transferSize: integer('transfer_size'),
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
    role: text('role').default('user').notNull(),
    banned: integer('banned', { mode: 'boolean' }),
    banReason: text('ban_reason'),
    banExpires: integer('ban_expires', { mode: 'timestamp' }),
    notifications: text('notifications', { mode: 'json' })
        .default([])
        .$type<string[]>()
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
    activeOrganizationId: text('active_organization_id'),
    impersonatedBy: text('impersonated_by'),
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
    aaguid: text('aaguid'),
});

export const lock = sqliteTable('lock', {
    key: text('key').primaryKey(),
    expires: integer('expires', { mode: 'timestamp' }).notNull(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const member = sqliteTable('member', {
    id: text('id').primaryKey(),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    role: text('role').default('member').notNull(),
    createdAt: integer('created_at', { mode: 'timestamp' }).notNull(),
});

export const invitation = sqliteTable('invitation', {
    id: text('id').primaryKey(),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    email: text('email').notNull(),
    role: text('role'),
    status: text('status').default('pending').notNull(),
    expiresAt: integer('expires_at', { mode: 'timestamp' }).notNull(),
    inviterId: text('inviter_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
});

export const ssoProvider = sqliteTable('sso_provider', {
    id: text('id').primaryKey(),
    issuer: text('issuer').notNull(),
    oidcConfig: text('oidc_config'),
    samlConfig: text('saml_config'),
    userId: text('user_id').references(() => user.id, { onDelete: 'cascade' }),
    providerId: text('provider_id').notNull().unique(),
    organizationId: text('organization_id'),
    domain: text('domain').notNull(),
});

// Relations
export const projectRelations = relations(project, ({ one, many }) => ({
    organization: one(organization, {
        fields: [project.organizationId],
        references: [organization.id],
    }),
    shops: many(shop),
}));

export const shopRelations = relations(shop, ({ one }) => ({
    organization: one(organization, {
        fields: [shop.organizationId],
        references: [organization.id],
    }),
    project: one(project, {
        fields: [shop.projectId],
        references: [project.id],
    }),
}));

export const organizationRelations = relations(organization, ({ many }) => ({
    projects: many(project),
    shops: many(shop),
}));

export const schema = {
    shop,
    shopSitespeed,
    shopChangelog,
    userNotification,

    // Better Auth
    user,
    session,
    account,
    verification,
    passkey,
    organization,
    project,
    member,
    invitation,
    ssoProvider,

    // Relations
    projectRelations,
    shopRelations,
    organizationRelations,
};

export type Drizzle = LibSQLDatabase<typeof schema>;
let drizzle: Drizzle | undefined;

export function getConnection(applyPragmas = true) {
    if (drizzle !== undefined) {
        return drizzle;
    }

    const dbPath = process.env.APP_DATABASE_PATH || 'shopmon.db';

    const client = createClient({
        url: `file:${dbPath}`,
    });

    if (applyPragmas) {
        const promises = [
            client.execute('PRAGMA journal_mode = WAL'),
            client.execute('PRAGMA cache_size = -64000'),
            client.execute('PRAGMA foreign_keys = ON'),
            client.execute('PRAGMA synchronous = NORMAL'),
            client.execute('PRAGMA temp_store = MEMORY'),
            client.execute('PRAGMA wal_autocheckpoint = 0'),
        ];
        Promise.all(promises).then(() => {
            console.log('Database PRAGMAs applied successfully');
        });
    }

    drizzle = drizzleSqlite(client, { schema });

    return drizzle;
}
