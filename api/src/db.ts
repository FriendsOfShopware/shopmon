import { relations } from 'drizzle-orm';
import { drizzle as drizzlePostgres } from 'drizzle-orm/postgres-js';
import {
    boolean,
    integer,
    jsonb,
    pgTable,
    real,
    serial,
    text,
    timestamp,
    unique,
    varchar,
} from 'drizzle-orm/pg-core';
import postgres from 'postgres';
import type { ExtensionChangelog, ExtensionDiff, NotificationLink } from './types/index.ts';

type LastChangelog = {
    date: Date;
    from: string;
    to: string;
};

export const organization = pgTable('organization', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    slug: text('slug').unique(),
    logo: text('logo'),
    createdAt: timestamp('created_at').notNull(),
    metadata: text('metadata'),
});

export const project = pgTable('project', {
    id: serial('id').primaryKey(),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    name: text('name').notNull(),
    description: text('description'),
    createdAt: timestamp('created_at').notNull(),
    updatedAt: timestamp('updated_at').notNull(),
});

export const shop = pgTable('shop', {
    id: serial('id').primaryKey(),
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
    lastScrapedAt: timestamp('last_scraped_at'),
    lastScrapedError: text('last_scraped_error'),
    ignores: jsonb('ignores')
        .default([])
        .$type<string[]>()
        .notNull(),
    shopImage: text('shop_image'),
    lastChangelog: jsonb('last_changelog')
        .default({})
        .$type<LastChangelog>(),
    connectionIssueCount: integer('connection_issue_count')
        .default(0)
        .notNull(),
    sitespeedEnabled: boolean('sitespeed_enabled')
        .default(false)
        .notNull(),
    sitespeedUrls: jsonb('sitespeed_urls')
        .default([])
        .$type<string[]>()
        .notNull(),
    createdAt: timestamp('created_at').notNull(),
});

export const shopSitespeed = pgTable('shop_sitespeed', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id').references(() => shop.id, { onDelete: 'cascade' }),
    createdAt: timestamp('created_at').notNull(),
    ttfb: integer('ttfb'),
    fullyLoaded: integer('fully_loaded'),
    largestContentfulPaint: integer('largest_contentful_paint'),
    firstContentfulPaint: integer('first_contentful_paint'),
    cumulativeLayoutShift: integer('cumulative_layout_shift'),
    transferSize: integer('transfer_size'),
});

export const shopChangelog = pgTable('shop_changelog', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id').references(() => shop.id, { onDelete: 'cascade' }),
    extensions: jsonb('extensions')
        .notNull()
        .$type<ExtensionDiff[]>(),
    oldShopwareVersion: text('old_shopware_version'),
    newShopwareVersion: text('new_shopware_version'),
    date: timestamp('date').notNull(),
});

// New tables for shop scrape info (previously stored in gzipped files)

export const shopScrapeInfo = pgTable('shop_scrape_info', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' })
        .unique(),
    // Cache info fields
    cacheEnvironment: varchar('cache_environment', { length: 255 }),
    cacheHttpCache: boolean('cache_http_cache'),
    cacheAdapter: varchar('cache_adapter', { length: 255 }),
    createdAt: timestamp('created_at').notNull(),
    updatedAt: timestamp('updated_at').notNull(),
});

export const shopExtension = pgTable('shop_extension', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    name: varchar('name', { length: 255 }).notNull(),
    label: text('label').notNull(),
    active: boolean('active').notNull(),
    version: varchar('version', { length: 100 }).notNull(),
    latestVersion: varchar('latest_version', { length: 100 }),
    installed: boolean('installed').notNull(),
    ratingAverage: real('rating_average'),
    storeLink: text('store_link'),
    changelog: jsonb('changelog').$type<ExtensionChangelog[] | null>(),
    installedAt: timestamp('installed_at'),
});

export const shopScheduledTask = pgTable('shop_scheduled_task', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    taskId: varchar('task_id', { length: 255 }).notNull(),
    name: varchar('name', { length: 255 }).notNull(),
    status: varchar('status', { length: 50 }).notNull(),
    interval: integer('interval').notNull(),
    overdue: boolean('overdue').notNull(),
    lastExecutionTime: timestamp('last_execution_time'),
    nextExecutionTime: timestamp('next_execution_time'),
});

export const shopQueueInfo = pgTable('shop_queue_info', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    name: varchar('name', { length: 255 }).notNull(),
    size: integer('size').notNull(),
});

export const shopCheck = pgTable('shop_check', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    checkId: varchar('check_id', { length: 255 }).notNull(),
    level: varchar('level', { length: 20 }).notNull(), // green, yellow, red
    message: text('message').notNull(),
    source: varchar('source', { length: 255 }).notNull(),
    link: text('link'),
});

export const userNotification = pgTable(
    'user_notification',
    {
        id: serial('id').primaryKey(),
        userId: text('user_id')
            .notNull()
            .references(() => user.id),
        key: text('key').notNull(),
        level: text('level').notNull(),
        title: text('title').notNull(),
        message: text('message').notNull(),
        link: jsonb('link')
            .notNull()
            .$type<NotificationLink>(),
        read: boolean('read').notNull().default(false),
        createdAt: timestamp('created_at').notNull(),
    },
    (table) => {
        return {
            keyUnique: unique().on(table.userId, table.key),
        };
    },
);

export const user = pgTable('user', {
    id: text('id').primaryKey(),
    name: text('name').notNull(),
    email: text('email').notNull().unique(),
    emailVerified: boolean('email_verified')
        .$defaultFn(() => false)
        .notNull(),
    image: text('image'),
    createdAt: timestamp('created_at')
        .$defaultFn(() => new Date())
        .notNull(),
    updatedAt: timestamp('updated_at')
        .$defaultFn(() => new Date())
        .notNull(),
    role: text('role').default('user').notNull(),
    banned: boolean('banned').default(false),
    banReason: text('ban_reason'),
    banExpires: timestamp('ban_expires'),
    notifications: jsonb('notifications')
        .default([])
        .$type<string[]>(),
});

export const session = pgTable('session', {
    id: text('id').primaryKey(),
    expiresAt: timestamp('expires_at').notNull(),
    token: text('token').notNull().unique(),
    createdAt: timestamp('created_at')
        .$defaultFn(() => new Date())
        .notNull(),
    updatedAt: timestamp('updated_at')
        .$defaultFn(() => new Date())
        .$onUpdate(() => new Date())
        .notNull(),
    ipAddress: text('ip_address'),
    userAgent: text('user_agent'),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    impersonatedBy: text('impersonated_by'),
    activeOrganizationId: text('active_organization_id'),
});

export const account = pgTable('account', {
    id: text('id').primaryKey(),
    accountId: text('account_id').notNull(),
    providerId: text('provider_id').notNull(),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    accessToken: text('access_token'),
    refreshToken: text('refresh_token'),
    idToken: text('id_token'),
    accessTokenExpiresAt: timestamp('access_token_expires_at'),
    refreshTokenExpiresAt: timestamp('refresh_token_expires_at'),
    scope: text('scope'),
    password: text('password'),
    createdAt: timestamp('created_at')
        .$defaultFn(() => new Date())
        .notNull(),
    updatedAt: timestamp('updated_at')
        .$onUpdate(() => new Date())
        .notNull(),
});

export const verification = pgTable('verification', {
    id: text('id').primaryKey(),
    identifier: text('identifier').notNull(),
    value: text('value').notNull(),
    expiresAt: timestamp('expires_at').notNull(),
    createdAt: timestamp('created_at')
        .$defaultFn(() => new Date())
        .notNull(),
    updatedAt: timestamp('updated_at')
        .$defaultFn(() => new Date())
        .$onUpdate(() => new Date())
        .notNull(),
});

export const passkey = pgTable('passkey', {
    id: text('id').primaryKey(),
    name: text('name'),
    publicKey: text('public_key').notNull(),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    credentialID: text('credential_id').notNull(),
    counter: integer('counter').notNull(),
    deviceType: text('device_type').notNull(),
    backedUp: boolean('backed_up').notNull(),
    transports: text('transports'),
    createdAt: timestamp('created_at'),
    aaguid: text('aaguid'),
});

export const lock = pgTable('lock', {
    key: text('key').primaryKey(),
    expires: timestamp('expires').notNull(),
    createdAt: timestamp('created_at').notNull(),
});

export const member = pgTable('member', {
    id: text('id').primaryKey(),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    userId: text('user_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
    role: text('role').default('member').notNull(),
    createdAt: timestamp('created_at').notNull(),
});

export const invitation = pgTable('invitation', {
    id: text('id').primaryKey(),
    organizationId: text('organization_id')
        .notNull()
        .references(() => organization.id, { onDelete: 'cascade' }),
    email: text('email').notNull(),
    role: text('role'),
    status: text('status').default('pending').notNull(),
    expiresAt: timestamp('expires_at').notNull(),
    inviterId: text('inviter_id')
        .notNull()
        .references(() => user.id, { onDelete: 'cascade' }),
});

export const ssoProvider = pgTable('sso_provider', {
    id: text('id').primaryKey(),
    issuer: text('issuer').notNull(),
    oidcConfig: text('oidc_config'),
    samlConfig: text('saml_config'),
    userId: text('user_id').references(() => user.id, { onDelete: 'cascade' }),
    providerId: text('provider_id').notNull().unique(),
    organizationId: text('organization_id'),
    domain: text('domain').notNull(),
});

export const deploymentToken = pgTable('deployment_token', {
    id: text('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    token: text('token').notNull().unique(),
    name: text('name').notNull(),
    createdAt: timestamp('created_at').notNull(),
    lastUsedAt: timestamp('last_used_at'),
});

export const deployment = pgTable('deployment', {
    id: serial('id').primaryKey(),
    shopId: integer('shop_id')
        .notNull()
        .references(() => shop.id, { onDelete: 'cascade' }),
    name: text('name').notNull(),
    command: text('command').notNull(),
    output: text('output').notNull(),
    returnCode: integer('return_code').notNull(),
    startDate: timestamp('start_date').notNull(),
    endDate: timestamp('end_date').notNull(),
    executionTime: text('execution_time').notNull(), // stored as string to preserve decimal precision
    composer: jsonb('composer')
        .default({})
        .$type<Record<string, string>>(),
    createdAt: timestamp('created_at').notNull(),
});

// Relations
export const projectRelations = relations(project, ({ one, many }) => ({
    organization: one(organization, {
        fields: [project.organizationId],
        references: [organization.id],
    }),
    shops: many(shop),
}));

export const shopRelations = relations(shop, ({ one, many }) => ({
    organization: one(organization, {
        fields: [shop.organizationId],
        references: [organization.id],
    }),
    project: one(project, {
        fields: [shop.projectId],
        references: [project.id],
    }),
    scrapeInfo: one(shopScrapeInfo, {
        fields: [shop.id],
        references: [shopScrapeInfo.shopId],
    }),
    extensions: many(shopExtension),
    scheduledTasks: many(shopScheduledTask),
    queueInfo: many(shopQueueInfo),
    checks: many(shopCheck),
}));

export const organizationRelations = relations(organization, ({ many }) => ({
    projects: many(project),
    shops: many(shop),
}));

export const deploymentRelations = relations(deployment, ({ one }) => ({
    shop: one(shop, {
        fields: [deployment.shopId],
        references: [shop.id],
    }),
}));

export const deploymentTokenRelations = relations(
    deploymentToken,
    ({ one }) => ({
        shop: one(shop, {
            fields: [deploymentToken.shopId],
            references: [shop.id],
        }),
    }),
);

export const shopScrapeInfoRelations = relations(shopScrapeInfo, ({ one }) => ({
    shop: one(shop, {
        fields: [shopScrapeInfo.shopId],
        references: [shop.id],
    }),
}));

export const shopExtensionRelations = relations(shopExtension, ({ one }) => ({
    shop: one(shop, {
        fields: [shopExtension.shopId],
        references: [shop.id],
    }),
}));

export const shopScheduledTaskRelations = relations(shopScheduledTask, ({ one }) => ({
    shop: one(shop, {
        fields: [shopScheduledTask.shopId],
        references: [shop.id],
    }),
}));

export const shopQueueInfoRelations = relations(shopQueueInfo, ({ one }) => ({
    shop: one(shop, {
        fields: [shopQueueInfo.shopId],
        references: [shop.id],
    }),
}));

export const shopCheckRelations = relations(shopCheck, ({ one }) => ({
    shop: one(shop, {
        fields: [shopCheck.shopId],
        references: [shop.id],
    }),
}));

export const schema = {
    shop,
    shopSitespeed,
    shopChangelog,
    userNotification,

    // Shop scrape data tables
    shopScrapeInfo,
    shopExtension,
    shopScheduledTask,
    shopQueueInfo,
    shopCheck,

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

    // Deployments
    deployment,
    deploymentToken,

    // Relations
    projectRelations,
    shopRelations,
    organizationRelations,
    deploymentRelations,
    deploymentTokenRelations,
    shopScrapeInfoRelations,
    shopExtensionRelations,
    shopScheduledTaskRelations,
    shopQueueInfoRelations,
    shopCheckRelations,
};

export type Drizzle = ReturnType<typeof drizzlePostgres<typeof schema>>;
let drizzle: Drizzle | undefined;

export function getConnection() {
    if (drizzle !== undefined) {
        return drizzle;
    }

    const databaseUrl = process.env.DATABASE_URL || 'postgresql://localhost:5432/shopmon';

    const client = postgres(databaseUrl);

    drizzle = drizzlePostgres(client, { schema });

    return drizzle;
}
