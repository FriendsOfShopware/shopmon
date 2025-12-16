/**
 * Migration script to transfer data from SQLite to PostgreSQL
 *
 * Usage:
 *   bun migrate-sqlite-to-postgres.ts [path-to-sqlite-db]
 *
 * Default SQLite path: ./shopmon.db
 * Requires DATABASE_URL environment variable for PostgreSQL connection
 */

import { Database } from 'bun:sqlite';
import { resolve } from 'node:path';
import { homedir } from 'node:os';
import { closeConnection, getConnection, schema } from '#src/db.ts';

// Expand ~ to home directory
function expandPath(path: string): string {
    if (path.startsWith('~/')) {
        return resolve(homedir(), path.slice(2));
    }
    return resolve(path);
}

const sqlitePath = expandPath(process.argv[2] || './shopmon.db');

console.log(`Migrating from SQLite (${sqlitePath}) to PostgreSQL...`);

// Check if file exists
import { existsSync } from 'node:fs';
if (!existsSync(sqlitePath)) {
    console.error(`Error: SQLite database file not found: ${sqlitePath}`);
    process.exit(1);
}

// Open SQLite database (can't use readonly with WAL mode)
const sqlite = new Database(sqlitePath);

// Test connection
try {
    sqlite.query('SELECT 1').get();
    console.log('SQLite connection successful');
} catch (e) {
    console.error('Failed to connect to SQLite:', (e as Error).message);
    process.exit(1);
}

// Get PostgreSQL connection
const pg = getConnection();

// Helper to get all rows from SQLite table
function getAll<T>(table: string): T[] {
    try {
        return sqlite.query(`SELECT * FROM ${table}`).all() as T[];
    } catch (e) {
        console.warn(`Warning: Could not read table ${table}:`, (e as Error).message);
        return [];
    }
}

// Helper to convert SQLite timestamp (integer or string) to Date
function toDate(value: number | string | null): Date | null {
    if (value === null || value === undefined) return null;

    // If it's a string (ISO date or date string), parse it directly
    if (typeof value === 'string') {
        const parsed = new Date(value);
        if (!Number.isNaN(parsed.getTime())) {
            return parsed;
        }
        return null;
    }

    // If it's a number, treat as Unix timestamp (seconds since epoch)
    return new Date(value * 1000);
}

// Helper to convert SQLite boolean (0/1) to boolean
function toBool(value: number | null): boolean {
    return value === 1;
}

// Helper to parse JSON fields
// biome-ignore lint/suspicious/noExplicitAny: Migration script needs flexible JSON parsing
function parseJson<T>(value: string | null, defaultValue: T): any {
    if (value === null || value === undefined) return defaultValue;
    try {
        return JSON.parse(value);
    } catch {
        return defaultValue;
    }
}

// Helper to batch insert with chunks (PostgreSQL has max 65534 parameters)
// biome-ignore lint/suspicious/noExplicitAny: Migration helper
async function batchInsert(table: any, values: any[], batchSize = 500) {
    for (let i = 0; i < values.length; i += batchSize) {
        const batch = values.slice(i, i + batchSize);
        await pg.insert(table).values(batch).onConflictDoNothing();
    }
}

async function migrate() {
    console.log('\n--- Starting migration ---\n');

    // 1. Migrate organizations (no dependencies)
    console.log('Migrating organizations...');
    const organizations = getAll<{
        id: string;
        name: string;
        slug: string | null;
        logo: string | null;
        created_at: number | string;
        metadata: string | null;
    }>('organization');

    if (organizations.length > 0) {
        await pg.insert(schema.organization).values(
            organizations.map(o => ({
                id: o.id,
                name: o.name,
                slug: o.slug,
                logo: o.logo,
                createdAt: toDate(o.created_at)!,
                metadata: o.metadata,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${organizations.length} organizations`);
    }

    // 2. Migrate users (no dependencies)
    console.log('Migrating users...');
    const users = getAll<{
        id: string;
        name: string;
        email: string;
        email_verified: number;
        image: string | null;
        created_at: number | string;
        updated_at: number | string;
        role: string;
        banned: number | null;
        ban_reason: string | null;
        ban_expires: number | null;
        notifications: string | null;
    }>('user');

    if (users.length > 0) {
        await pg.insert(schema.user).values(
            users.map(u => ({
                id: u.id,
                name: u.name,
                email: u.email,
                emailVerified: toBool(u.email_verified),
                image: u.image,
                createdAt: toDate(u.created_at)!,
                updatedAt: toDate(u.updated_at)!,
                role: u.role,
                banned: u.banned !== null ? toBool(u.banned) : false,
                banReason: u.ban_reason,
                banExpires: toDate(u.ban_expires),
                notifications: parseJson<string[]>(u.notifications, []),
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${users.length} users`);
    }

    // 3. Migrate projects (depends on organization)
    console.log('Migrating projects...');
    const projects = getAll<{
        id: number;
        organization_id: string;
        name: string;
        description: string | null;
        created_at: number | string;
        updated_at: number | string;
    }>('project');

    if (projects.length > 0) {
        await pg.insert(schema.project).values(
            projects.map(p => ({
                id: p.id,
                organizationId: p.organization_id,
                name: p.name,
                description: p.description,
                createdAt: toDate(p.created_at)!,
                updatedAt: toDate(p.updated_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${projects.length} projects`);

        // Reset sequence for serial column
        const maxProjectId = Math.max(...projects.map(p => p.id));
        await pg.execute(`SELECT setval('project_id_seq', ${maxProjectId}, true)`);
    }

    // 4. Migrate shops (depends on organization, project)
    console.log('Migrating shops...');
    const shops = getAll<{
        id: number;
        organization_id: string;
        project_id: number;
        name: string;
        status: string;
        url: string;
        favicon: string | null;
        client_id: string;
        client_secret: string;
        shopware_version: string;
        last_scraped_at: number | null;
        last_scraped_error: string | null;
        ignores: string;
        shop_image: string | null;
        last_changelog: string | null;
        connection_issue_count: number;
        sitespeed_enabled: number;
        sitespeed_urls: string;
        created_at: number | string;
    }>('shop');

    if (shops.length > 0) {
        await pg.insert(schema.shop).values(
            shops.map(s => ({
                id: s.id,
                organizationId: s.organization_id,
                projectId: s.project_id,
                name: s.name,
                status: s.status,
                url: s.url,
                favicon: s.favicon,
                clientId: s.client_id,
                clientSecret: s.client_secret,
                shopwareVersion: s.shopware_version,
                lastScrapedAt: toDate(s.last_scraped_at),
                lastScrapedError: s.last_scraped_error,
                ignores: parseJson<string[]>(s.ignores, []),
                shopImage: s.shop_image,
                lastChangelog: parseJson(s.last_changelog, {}),
                connectionIssueCount: s.connection_issue_count,
                sitespeedEnabled: toBool(s.sitespeed_enabled),
                sitespeedUrls: parseJson<string[]>(s.sitespeed_urls, []),
                createdAt: toDate(s.created_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${shops.length} shops`);

        const maxShopId = Math.max(...shops.map(s => s.id));
        await pg.execute(`SELECT setval('shop_id_seq', ${maxShopId}, true)`);
    }

    // 5. Migrate sessions (depends on user)
    console.log('Migrating sessions...');
    const sessions = getAll<{
        id: string;
        expires_at: number | string;
        token: string;
        created_at: number | string;
        updated_at: number | string;
        ip_address: string | null;
        user_agent: string | null;
        user_id: string;
        impersonated_by: string | null;
        active_organization_id: string | null;
    }>('session');

    if (sessions.length > 0) {
        await pg.insert(schema.session).values(
            sessions.map(s => ({
                id: s.id,
                expiresAt: toDate(s.expires_at)!,
                token: s.token,
                createdAt: toDate(s.created_at)!,
                updatedAt: toDate(s.updated_at)!,
                ipAddress: s.ip_address,
                userAgent: s.user_agent,
                userId: s.user_id,
                impersonatedBy: s.impersonated_by,
                activeOrganizationId: s.active_organization_id,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${sessions.length} sessions`);
    }

    // 6. Migrate accounts (depends on user)
    console.log('Migrating accounts...');
    const accounts = getAll<{
        id: string;
        account_id: string;
        provider_id: string;
        user_id: string;
        access_token: string | null;
        refresh_token: string | null;
        id_token: string | null;
        access_token_expires_at: number | null;
        refresh_token_expires_at: number | null;
        scope: string | null;
        password: string | null;
        created_at: number | string;
        updated_at: number | string;
    }>('account');

    if (accounts.length > 0) {
        await pg.insert(schema.account).values(
            accounts.map(a => ({
                id: a.id,
                accountId: a.account_id,
                providerId: a.provider_id,
                userId: a.user_id,
                accessToken: a.access_token,
                refreshToken: a.refresh_token,
                idToken: a.id_token,
                accessTokenExpiresAt: toDate(a.access_token_expires_at),
                refreshTokenExpiresAt: toDate(a.refresh_token_expires_at),
                scope: a.scope,
                password: a.password,
                createdAt: toDate(a.created_at)!,
                updatedAt: toDate(a.updated_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${accounts.length} accounts`);
    }

    // 7. Migrate verifications
    console.log('Migrating verifications...');
    const verifications = getAll<{
        id: string;
        identifier: string;
        value: string;
        expires_at: number | string;
        created_at: number | string;
        updated_at: number | string;
    }>('verification');

    if (verifications.length > 0) {
        await pg.insert(schema.verification).values(
            verifications.map(v => ({
                id: v.id,
                identifier: v.identifier,
                value: v.value,
                expiresAt: toDate(v.expires_at)!,
                createdAt: toDate(v.created_at)!,
                updatedAt: toDate(v.updated_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${verifications.length} verifications`);
    }

    // 8. Migrate passkeys (depends on user)
    console.log('Migrating passkeys...');
    const passkeys = getAll<{
        id: string;
        name: string | null;
        public_key: string;
        user_id: string;
        credential_id: string;
        counter: number;
        device_type: string;
        backed_up: number;
        transports: string | null;
        created_at: number | null;
        aaguid: string | null;
    }>('passkey');

    if (passkeys.length > 0) {
        await pg.insert(schema.passkey).values(
            passkeys.map(p => ({
                id: p.id,
                name: p.name,
                publicKey: p.public_key,
                userId: p.user_id,
                credentialID: p.credential_id,
                counter: p.counter,
                deviceType: p.device_type,
                backedUp: toBool(p.backed_up),
                transports: p.transports,
                createdAt: toDate(p.created_at),
                aaguid: p.aaguid,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${passkeys.length} passkeys`);
    }

    // 9. Migrate members (depends on organization, user)
    console.log('Migrating members...');
    const members = getAll<{
        id: string;
        organization_id: string;
        user_id: string;
        role: string;
        created_at: number | string;
    }>('member');

    if (members.length > 0) {
        await pg.insert(schema.member).values(
            members.map(m => ({
                id: m.id,
                organizationId: m.organization_id,
                userId: m.user_id,
                role: m.role,
                createdAt: toDate(m.created_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${members.length} members`);
    }

    // 10. Migrate invitations (depends on organization, user)
    console.log('Migrating invitations...');
    const invitations = getAll<{
        id: string;
        organization_id: string;
        email: string;
        role: string | null;
        status: string;
        expires_at: number | string;
        inviter_id: string;
    }>('invitation');

    if (invitations.length > 0) {
        await pg.insert(schema.invitation).values(
            invitations.map(i => ({
                id: i.id,
                organizationId: i.organization_id,
                email: i.email,
                role: i.role,
                status: i.status,
                expiresAt: toDate(i.expires_at)!,
                inviterId: i.inviter_id,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${invitations.length} invitations`);
    }

    // 11. Migrate SSO providers
    console.log('Migrating SSO providers...');
    const ssoProviders = getAll<{
        id: string;
        issuer: string;
        oidc_config: string | null;
        saml_config: string | null;
        user_id: string | null;
        provider_id: string;
        organization_id: string | null;
        domain: string;
    }>('sso_provider');

    if (ssoProviders.length > 0) {
        await pg.insert(schema.ssoProvider).values(
            ssoProviders.map(s => ({
                id: s.id,
                issuer: s.issuer,
                oidcConfig: s.oidc_config,
                samlConfig: s.saml_config,
                userId: s.user_id,
                providerId: s.provider_id,
                organizationId: s.organization_id,
                domain: s.domain,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${ssoProviders.length} SSO providers`);
    }

    // 12. Migrate project API keys (depends on project)
    console.log('Migrating project API keys...');
    const apiKeys = getAll<{
        id: string;
        project_id: number;
        name: string;
        token: string;
        scopes: string;
        created_at: number | string;
        last_used_at: number | null;
    }>('project_api_key');

    if (apiKeys.length > 0) {
        await pg.insert(schema.projectApiKey).values(
            apiKeys.map(k => ({
                id: k.id,
                projectId: k.project_id,
                name: k.name,
                token: k.token,
                scopes: parseJson(k.scopes, []),
                createdAt: toDate(k.created_at)!,
                lastUsedAt: toDate(k.last_used_at),
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${apiKeys.length} API keys`);
    }

    // 13. Migrate user notifications (depends on user)
    console.log('Migrating user notifications...');
    const notifications = getAll<{
        id: number;
        user_id: string;
        key: string;
        level: string;
        title: string;
        message: string;
        link: string;
        read: number;
        created_at: number | string;
    }>('user_notification');

    if (notifications.length > 0) {
        await pg.insert(schema.userNotification).values(
            notifications.map(n => ({
                id: n.id,
                userId: n.user_id,
                key: n.key,
                level: n.level,
                title: n.title,
                message: n.message,
                link: parseJson(n.link, { text: '', url: '' }),
                read: toBool(n.read),
                createdAt: toDate(n.created_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${notifications.length} notifications`);

        const maxNotifId = Math.max(...notifications.map(n => n.id));
        await pg.execute(`SELECT setval('user_notification_id_seq', ${maxNotifId}, true)`);
    }

    // 14. Migrate shop extensions (depends on shop)
    console.log('Migrating shop extensions...');
    const extensions = getAll<{
        id: number;
        shop_id: number;
        name: string;
        label: string;
        active: number;
        version: string;
        latest_version: string | null;
        installed: number;
        rating_average: number | null;
        store_link: string | null;
        changelog: string | null;
        installed_at: string | null;
    }>('shop_extension');

    if (extensions.length > 0) {
        await pg.insert(schema.shopExtension).values(
            extensions.map(e => ({
                id: e.id,
                shopId: e.shop_id,
                name: e.name,
                label: e.label,
                active: toBool(e.active),
                version: e.version,
                latestVersion: e.latest_version,
                installed: toBool(e.installed),
                ratingAverage: e.rating_average,
                storeLink: e.store_link,
                changelog: parseJson(e.changelog, null),
                installedAt: e.installed_at,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${extensions.length} shop extensions`);

        const maxExtId = Math.max(...extensions.map(e => e.id));
        await pg.execute(`SELECT setval('shop_extension_id_seq', ${maxExtId}, true)`);
    }

    // 15. Migrate shop scheduled tasks (depends on shop)
    console.log('Migrating shop scheduled tasks...');
    const scheduledTasks = getAll<{
        id: number;
        shop_id: number;
        task_id: string;
        name: string;
        status: string;
        interval: number;
        overdue: number;
        last_execution_time: string | null;
        next_execution_time: string | null;
    }>('shop_scheduled_task');

    if (scheduledTasks.length > 0) {
        await pg.insert(schema.shopScheduledTask).values(
            scheduledTasks.map(t => ({
                id: t.id,
                shopId: t.shop_id,
                taskId: t.task_id,
                name: t.name,
                status: t.status,
                interval: t.interval,
                overdue: toBool(t.overdue),
                lastExecutionTime: t.last_execution_time,
                nextExecutionTime: t.next_execution_time,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${scheduledTasks.length} scheduled tasks`);

        const maxTaskId = Math.max(...scheduledTasks.map(t => t.id));
        await pg.execute(`SELECT setval('shop_scheduled_task_id_seq', ${maxTaskId}, true)`);
    }

    // 16. Migrate shop queues (depends on shop)
    console.log('Migrating shop queues...');
    const queues = getAll<{
        id: number;
        shop_id: number;
        name: string;
        size: number;
    }>('shop_queue');

    if (queues.length > 0) {
        await pg.insert(schema.shopQueue).values(
            queues.map(q => ({
                id: q.id,
                shopId: q.shop_id,
                name: q.name,
                size: q.size,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${queues.length} shop queues`);

        const maxQueueId = Math.max(...queues.map(q => q.id));
        await pg.execute(`SELECT setval('shop_queue_id_seq', ${maxQueueId}, true)`);
    }

    // 17. Migrate shop cache (depends on shop)
    console.log('Migrating shop cache...');
    const caches = getAll<{
        id: number;
        shop_id: number;
        environment: string;
        http_cache: number;
        cache_adapter: string;
    }>('shop_cache');

    if (caches.length > 0) {
        await pg.insert(schema.shopCache).values(
            caches.map(c => ({
                id: c.id,
                shopId: c.shop_id,
                environment: c.environment,
                httpCache: toBool(c.http_cache),
                cacheAdapter: c.cache_adapter,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${caches.length} shop caches`);

        const maxCacheId = Math.max(...caches.map(c => c.id));
        await pg.execute(`SELECT setval('shop_cache_id_seq', ${maxCacheId}, true)`);
    }

    // 18. Migrate shop checks (depends on shop)
    console.log('Migrating shop checks...');
    const checks = getAll<{
        id: number;
        shop_id: number;
        check_id: string;
        level: string;
        message: string;
        source: string;
        link: string | null;
    }>('shop_check');

    if (checks.length > 0) {
        await pg.insert(schema.shopCheck).values(
            checks.map(c => ({
                id: c.id,
                shopId: c.shop_id,
                checkId: c.check_id,
                level: c.level,
                message: c.message,
                source: c.source,
                link: c.link,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${checks.length} shop checks`);

        const maxCheckId = Math.max(...checks.map(c => c.id));
        await pg.execute(`SELECT setval('shop_check_id_seq', ${maxCheckId}, true)`);
    }

    // 19. Migrate shop sitespeed (depends on shop)
    console.log('Migrating shop sitespeed...');
    const sitespeeds = getAll<{
        id: number;
        shop_id: number;
        created_at: number | string;
        ttfb: number | null;
        fully_loaded: number | null;
        largest_contentful_paint: number | null;
        first_contentful_paint: number | null;
        cumulative_layout_shift: number | null;
        transfer_size: number | null;
    }>('shop_sitespeed');

    if (sitespeeds.length > 0) {
        const sitespeedValues = sitespeeds.map(s => ({
            id: s.id,
            shopId: s.shop_id,
            createdAt: toDate(s.created_at)!,
            ttfb: s.ttfb,
            fullyLoaded: s.fully_loaded,
            largestContentfulPaint: s.largest_contentful_paint,
            firstContentfulPaint: s.first_contentful_paint,
            cumulativeLayoutShift: s.cumulative_layout_shift,
            transferSize: s.transfer_size,
        }));
        await batchInsert(schema.shopSitespeed, sitespeedValues);
        console.log(`  Migrated ${sitespeeds.length} sitespeed records`);

        const maxSitespeedId = Math.max(...sitespeeds.map(s => s.id));
        await pg.execute(`SELECT setval('shop_sitespeed_id_seq', ${maxSitespeedId}, true)`);
    }

    // 20. Migrate shop changelog (depends on shop)
    console.log('Migrating shop changelog...');
    const changelogs = getAll<{
        id: number;
        shop_id: number;
        extensions: string;
        old_shopware_version: string | null;
        new_shopware_version: string | null;
        date: number | string;
    }>('shop_changelog');

    if (changelogs.length > 0) {
        const changelogValues = changelogs.map(c => ({
            id: c.id,
            shopId: c.shop_id,
            extensions: parseJson(c.extensions, []),
            oldShopwareVersion: c.old_shopware_version,
            newShopwareVersion: c.new_shopware_version,
            date: toDate(c.date)!,
        }));
        await batchInsert(schema.shopChangelog, changelogValues);
        console.log(`  Migrated ${changelogs.length} changelog entries`);

        const maxChangelogId = Math.max(...changelogs.map(c => c.id));
        await pg.execute(`SELECT setval('shop_changelog_id_seq', ${maxChangelogId}, true)`);
    }

    // 21. Migrate locks
    console.log('Migrating locks...');
    const locks = getAll<{
        key: string;
        expires: number | string;
        created_at: number | string;
    }>('lock');

    if (locks.length > 0) {
        await pg.insert(schema.lock).values(
            locks.map(l => ({
                key: l.key,
                expires: toDate(l.expires)!,
                createdAt: toDate(l.created_at)!,
            }))
        ).onConflictDoNothing();
        console.log(`  Migrated ${locks.length} locks`);
    }

    console.log('\n--- Migration completed successfully! ---\n');
}

// Run migration
migrate()
    .catch((error) => {
        console.error('Migration failed:', error);
        process.exit(1);
    })
    .finally(async () => {
        sqlite.close();
        await closeConnection();
    });
