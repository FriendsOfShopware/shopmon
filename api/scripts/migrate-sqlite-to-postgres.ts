/**
 * Migration script to move data from SQLite to PostgreSQL
 *
 * Usage:
 *   DATABASE_URL=postgresql://... SQLITE_PATH=./shopmon.db bun run scripts/migrate-sqlite-to-postgres.ts
 *
 * This script:
 * 1. Reads all data from the SQLite database
 * 2. Loads gzipped shop scrape files and transforms them into PostgreSQL tables
 * 3. Inserts all data into PostgreSQL with proper relationships
 */

import { Database } from 'bun:sqlite';
import { readFile, readdir, stat } from 'node:fs/promises';
import { promisify } from 'node:util';
import { gunzip } from 'node:zlib';
import { drizzle as drizzlePostgres } from 'drizzle-orm/postgres-js';
import postgres from 'postgres';
import * as pgSchema from '../src/db.ts';

const pGunzip = promisify(gunzip);

// PostgreSQL connection
const databaseUrl = process.env.DATABASE_URL;
if (!databaseUrl) {
    console.error('DATABASE_URL environment variable is required');
    process.exit(1);
}

// SQLite connection
const sqlitePath = process.env.SQLITE_PATH || './shopmon.db';

interface ShopScrapeInfo {
    extensions: Array<{
        name: string;
        label: string;
        active: boolean;
        version: string;
        latestVersion: string | null;
        installed: boolean;
        ratingAverage: number | null;
        storeLink: string | null;
        changelog: Array<{
            version: string;
            text: string;
            creationDate: string;
            isCompatible: boolean;
        }> | null;
        installedAt: string | null;
    }>;
    scheduledTask: Array<{
        id: string;
        name: string;
        status: string;
        interval: number;
        overdue: boolean;
        lastExecutionTime: string;
        nextExecutionTime: string;
    }>;
    queueInfo: Array<{
        name: string;
        size: number;
    }>;
    cacheInfo: {
        environment: string;
        httpCache: boolean;
        cacheAdapter: string;
    };
    checks: Array<{
        id: string;
        level: string;
        message: string;
        source: string;
        link: string | null;
    }>;
    createdAt: string;
}

async function loadScrapeInfo(shopId: number): Promise<ShopScrapeInfo | null> {
    const filesDir = process.env.APP_FILES_DIR || 'files';
    const path = `${filesDir}/shops/${shopId}/scrape-info.json.gz`;

    try {
        const file = await readFile(path);
        const decompressed = await pGunzip(file);
        return JSON.parse(decompressed.toString());
    } catch (_e) {
        return null;
    }
}

function convertTimestamp(timestamp: number | null): Date | null {
    if (timestamp === null || timestamp === undefined) {
        return null;
    }
    return new Date(timestamp * 1000);
}

function convertTimestampRequired(timestamp: number): Date {
    return new Date(timestamp * 1000);
}

async function migrate() {
    console.log('Starting migration from SQLite to PostgreSQL...');
    console.log(`SQLite path: ${sqlitePath}`);
    console.log(`PostgreSQL URL: ${databaseUrl?.replace(/:[^:@]+@/, ':***@')}`);

    // Connect to databases
    const sqlite = new Database(sqlitePath);
    const pgClient = postgres(databaseUrl);
    const pg = drizzlePostgres(pgClient, { schema: pgSchema.schema });

    try {
        // Migrate organizations
        console.log('\n--- Migrating organizations ---');
        const organizations = sqlite.query('SELECT * FROM organization').all() as any[];
        console.log(`Found ${organizations.length} organizations`);

        for (const org of organizations) {
            await pg.insert(pgSchema.organization).values({
                id: org.id,
                name: org.name,
                slug: org.slug,
                logo: org.logo,
                createdAt: convertTimestampRequired(org.created_at),
                metadata: org.metadata,
            }).onConflictDoNothing();
        }
        console.log('Organizations migrated successfully');

        // Migrate users
        console.log('\n--- Migrating users ---');
        const users = sqlite.query('SELECT * FROM user').all() as any[];
        console.log(`Found ${users.length} users`);

        for (const user of users) {
            await pg.insert(pgSchema.user).values({
                id: user.id,
                name: user.name,
                email: user.email,
                emailVerified: Boolean(user.email_verified),
                image: user.image,
                createdAt: convertTimestampRequired(user.created_at),
                updatedAt: convertTimestampRequired(user.updated_at),
                role: user.role || 'user',
                banned: Boolean(user.banned),
                banReason: user.ban_reason,
                banExpires: convertTimestamp(user.ban_expires),
                notifications: user.notifications ? JSON.parse(user.notifications) : [],
            }).onConflictDoNothing();
        }
        console.log('Users migrated successfully');

        // Migrate accounts
        console.log('\n--- Migrating accounts ---');
        const accounts = sqlite.query('SELECT * FROM account').all() as any[];
        console.log(`Found ${accounts.length} accounts`);

        for (const account of accounts) {
            await pg.insert(pgSchema.account).values({
                id: account.id,
                accountId: account.account_id,
                providerId: account.provider_id,
                userId: account.user_id,
                accessToken: account.access_token,
                refreshToken: account.refresh_token,
                idToken: account.id_token,
                accessTokenExpiresAt: convertTimestamp(account.access_token_expires_at),
                refreshTokenExpiresAt: convertTimestamp(account.refresh_token_expires_at),
                scope: account.scope,
                password: account.password,
                createdAt: convertTimestampRequired(account.created_at),
                updatedAt: convertTimestampRequired(account.updated_at),
            }).onConflictDoNothing();
        }
        console.log('Accounts migrated successfully');

        // Migrate sessions
        console.log('\n--- Migrating sessions ---');
        const sessions = sqlite.query('SELECT * FROM session').all() as any[];
        console.log(`Found ${sessions.length} sessions`);

        for (const session of sessions) {
            await pg.insert(pgSchema.session).values({
                id: session.id,
                expiresAt: convertTimestampRequired(session.expires_at),
                token: session.token,
                createdAt: convertTimestampRequired(session.created_at),
                updatedAt: convertTimestampRequired(session.updated_at),
                ipAddress: session.ip_address,
                userAgent: session.user_agent,
                userId: session.user_id,
                impersonatedBy: session.impersonated_by,
                activeOrganizationId: session.active_organization_id,
            }).onConflictDoNothing();
        }
        console.log('Sessions migrated successfully');

        // Migrate verifications
        console.log('\n--- Migrating verifications ---');
        const verifications = sqlite.query('SELECT * FROM verification').all() as any[];
        console.log(`Found ${verifications.length} verifications`);

        for (const verification of verifications) {
            await pg.insert(pgSchema.verification).values({
                id: verification.id,
                identifier: verification.identifier,
                value: verification.value,
                expiresAt: convertTimestampRequired(verification.expires_at),
                createdAt: convertTimestampRequired(verification.created_at),
                updatedAt: convertTimestampRequired(verification.updated_at),
            }).onConflictDoNothing();
        }
        console.log('Verifications migrated successfully');

        // Migrate passkeys
        console.log('\n--- Migrating passkeys ---');
        const passkeys = sqlite.query('SELECT * FROM passkey').all() as any[];
        console.log(`Found ${passkeys.length} passkeys`);

        for (const passkey of passkeys) {
            await pg.insert(pgSchema.passkey).values({
                id: passkey.id,
                name: passkey.name,
                publicKey: passkey.public_key,
                userId: passkey.user_id,
                credentialID: passkey.credential_id,
                counter: passkey.counter,
                deviceType: passkey.device_type,
                backedUp: Boolean(passkey.backed_up),
                transports: passkey.transports,
                createdAt: convertTimestamp(passkey.created_at),
                aaguid: passkey.aaguid,
            }).onConflictDoNothing();
        }
        console.log('Passkeys migrated successfully');

        // Migrate members
        console.log('\n--- Migrating members ---');
        const members = sqlite.query('SELECT * FROM member').all() as any[];
        console.log(`Found ${members.length} members`);

        for (const member of members) {
            await pg.insert(pgSchema.member).values({
                id: member.id,
                organizationId: member.organization_id,
                userId: member.user_id,
                role: member.role || 'member',
                createdAt: convertTimestampRequired(member.created_at),
            }).onConflictDoNothing();
        }
        console.log('Members migrated successfully');

        // Migrate invitations
        console.log('\n--- Migrating invitations ---');
        const invitations = sqlite.query('SELECT * FROM invitation').all() as any[];
        console.log(`Found ${invitations.length} invitations`);

        for (const invitation of invitations) {
            await pg.insert(pgSchema.invitation).values({
                id: invitation.id,
                organizationId: invitation.organization_id,
                email: invitation.email,
                role: invitation.role,
                status: invitation.status || 'pending',
                expiresAt: convertTimestampRequired(invitation.expires_at),
                inviterId: invitation.inviter_id,
            }).onConflictDoNothing();
        }
        console.log('Invitations migrated successfully');

        // Migrate SSO providers
        console.log('\n--- Migrating SSO providers ---');
        const ssoProviders = sqlite.query('SELECT * FROM sso_provider').all() as any[];
        console.log(`Found ${ssoProviders.length} SSO providers`);

        for (const sso of ssoProviders) {
            await pg.insert(pgSchema.ssoProvider).values({
                id: sso.id,
                issuer: sso.issuer,
                oidcConfig: sso.oidc_config,
                samlConfig: sso.saml_config,
                userId: sso.user_id,
                providerId: sso.provider_id,
                organizationId: sso.organization_id,
                domain: sso.domain,
            }).onConflictDoNothing();
        }
        console.log('SSO providers migrated successfully');

        // Migrate projects
        console.log('\n--- Migrating projects ---');
        const projects = sqlite.query('SELECT * FROM project').all() as any[];
        console.log(`Found ${projects.length} projects`);

        for (const project of projects) {
            await pg.insert(pgSchema.project).values({
                id: project.id,
                organizationId: project.organization_id,
                name: project.name,
                description: project.description,
                createdAt: convertTimestampRequired(project.created_at),
                updatedAt: convertTimestampRequired(project.updated_at),
            }).onConflictDoNothing();
        }
        // Reset the sequence for project id
        await pgClient`SELECT setval('project_id_seq', (SELECT COALESCE(MAX(id), 0) FROM project))`;
        console.log('Projects migrated successfully');

        // Migrate locks
        console.log('\n--- Migrating locks ---');
        const locks = sqlite.query('SELECT * FROM lock').all() as any[];
        console.log(`Found ${locks.length} locks`);

        for (const lock of locks) {
            await pg.insert(pgSchema.lock).values({
                key: lock.key,
                expires: convertTimestampRequired(lock.expires),
                createdAt: convertTimestampRequired(lock.created_at),
            }).onConflictDoNothing();
        }
        console.log('Locks migrated successfully');

        // Migrate shops
        console.log('\n--- Migrating shops ---');
        const shops = sqlite.query('SELECT * FROM shop').all() as any[];
        console.log(`Found ${shops.length} shops`);

        for (const shop of shops) {
            await pg.insert(pgSchema.shop).values({
                id: shop.id,
                organizationId: shop.organization_id,
                projectId: shop.project_id,
                name: shop.name,
                status: shop.status || 'green',
                url: shop.url,
                favicon: shop.favicon,
                clientId: shop.client_id,
                clientSecret: shop.client_secret,
                shopwareVersion: shop.shopware_version,
                lastScrapedAt: convertTimestamp(shop.last_scraped_at),
                lastScrapedError: shop.last_scraped_error,
                ignores: shop.ignores ? JSON.parse(shop.ignores) : [],
                shopImage: shop.shop_image,
                lastChangelog: shop.last_changelog ? JSON.parse(shop.last_changelog) : {},
                connectionIssueCount: shop.connection_issue_count || 0,
                sitespeedEnabled: Boolean(shop.sitespeed_enabled),
                sitespeedUrls: shop.sitespeed_urls ? JSON.parse(shop.sitespeed_urls) : [],
                createdAt: convertTimestampRequired(shop.created_at),
            }).onConflictDoNothing();
        }
        // Reset the sequence for shop id
        await pgClient`SELECT setval('shop_id_seq', (SELECT COALESCE(MAX(id), 0) FROM shop))`;
        console.log('Shops migrated successfully');

        // Migrate shop sitespeed
        console.log('\n--- Migrating shop sitespeed ---');
        const sitespeeds = sqlite.query('SELECT * FROM shop_sitespeed').all() as any[];
        console.log(`Found ${sitespeeds.length} sitespeed records`);

        for (const ss of sitespeeds) {
            await pg.insert(pgSchema.shopSitespeed).values({
                id: ss.id,
                shopId: ss.shop_id,
                createdAt: convertTimestampRequired(ss.created_at),
                ttfb: ss.ttfb,
                fullyLoaded: ss.fully_loaded,
                largestContentfulPaint: ss.largest_contentful_paint,
                firstContentfulPaint: ss.first_contentful_paint,
                cumulativeLayoutShift: ss.cumulative_layout_shift,
                transferSize: ss.transfer_size,
            }).onConflictDoNothing();
        }
        // Reset the sequence
        await pgClient`SELECT setval('shop_sitespeed_id_seq', (SELECT COALESCE(MAX(id), 0) FROM shop_sitespeed))`;
        console.log('Shop sitespeed migrated successfully');

        // Migrate shop changelog
        console.log('\n--- Migrating shop changelog ---');
        const changelogs = sqlite.query('SELECT * FROM shop_changelog').all() as any[];
        console.log(`Found ${changelogs.length} changelog records`);

        for (const cl of changelogs) {
            await pg.insert(pgSchema.shopChangelog).values({
                id: cl.id,
                shopId: cl.shop_id,
                extensions: cl.extensions ? JSON.parse(cl.extensions) : [],
                oldShopwareVersion: cl.old_shopware_version,
                newShopwareVersion: cl.new_shopware_version,
                date: convertTimestampRequired(cl.date),
            }).onConflictDoNothing();
        }
        // Reset the sequence
        await pgClient`SELECT setval('shop_changelog_id_seq', (SELECT COALESCE(MAX(id), 0) FROM shop_changelog))`;
        console.log('Shop changelog migrated successfully');

        // Migrate user notifications
        console.log('\n--- Migrating user notifications ---');
        const notifications = sqlite.query('SELECT * FROM user_notification').all() as any[];
        console.log(`Found ${notifications.length} notification records`);

        for (const notif of notifications) {
            await pg.insert(pgSchema.userNotification).values({
                id: notif.id,
                userId: notif.user_id,
                key: notif.key,
                level: notif.level,
                title: notif.title,
                message: notif.message,
                link: notif.link ? JSON.parse(notif.link) : null,
                read: Boolean(notif.read),
                createdAt: convertTimestampRequired(notif.created_at),
            }).onConflictDoNothing();
        }
        // Reset the sequence
        await pgClient`SELECT setval('user_notification_id_seq', (SELECT COALESCE(MAX(id), 0) FROM user_notification))`;
        console.log('User notifications migrated successfully');

        // Migrate deployment tokens
        console.log('\n--- Migrating deployment tokens ---');
        const deploymentTokens = sqlite.query('SELECT * FROM deployment_token').all() as any[];
        console.log(`Found ${deploymentTokens.length} deployment tokens`);

        for (const token of deploymentTokens) {
            await pg.insert(pgSchema.deploymentToken).values({
                id: token.id,
                shopId: token.shop_id,
                token: token.token,
                name: token.name,
                createdAt: convertTimestampRequired(token.created_at),
                lastUsedAt: convertTimestamp(token.last_used_at),
            }).onConflictDoNothing();
        }
        console.log('Deployment tokens migrated successfully');

        // Migrate deployments
        console.log('\n--- Migrating deployments ---');
        const deployments = sqlite.query('SELECT * FROM deployment').all() as any[];
        console.log(`Found ${deployments.length} deployments`);

        for (const dep of deployments) {
            await pg.insert(pgSchema.deployment).values({
                id: dep.id,
                shopId: dep.shop_id,
                name: dep.name,
                command: dep.command,
                output: dep.output,
                returnCode: dep.return_code,
                startDate: convertTimestampRequired(dep.start_date),
                endDate: convertTimestampRequired(dep.end_date),
                executionTime: dep.execution_time,
                composer: dep.composer ? JSON.parse(dep.composer) : {},
                createdAt: convertTimestampRequired(dep.created_at),
            }).onConflictDoNothing();
        }
        // Reset the sequence
        await pgClient`SELECT setval('deployment_id_seq', (SELECT COALESCE(MAX(id), 0) FROM deployment))`;
        console.log('Deployments migrated successfully');

        // Migrate shop scrape info from gzipped files
        console.log('\n--- Migrating shop scrape info from gzipped files ---');
        let scrapeInfoCount = 0;
        let extensionCount = 0;
        let taskCount = 0;
        let queueCount = 0;
        let checkCount = 0;

        for (const shop of shops) {
            const scrapeInfo = await loadScrapeInfo(shop.id);
            if (!scrapeInfo) {
                continue;
            }

            scrapeInfoCount++;
            const now = new Date();
            const createdAt = scrapeInfo.createdAt ? new Date(scrapeInfo.createdAt) : now;

            // Insert scrape info
            await pg.insert(pgSchema.shopScrapeInfo).values({
                shopId: shop.id,
                cacheEnvironment: scrapeInfo.cacheInfo?.environment || null,
                cacheHttpCache: scrapeInfo.cacheInfo?.httpCache ?? null,
                cacheAdapter: scrapeInfo.cacheInfo?.cacheAdapter || null,
                createdAt: createdAt,
                updatedAt: createdAt,
            }).onConflictDoNothing();

            // Insert extensions
            if (scrapeInfo.extensions && scrapeInfo.extensions.length > 0) {
                for (const ext of scrapeInfo.extensions) {
                    await pg.insert(pgSchema.shopExtension).values({
                        shopId: shop.id,
                        name: ext.name,
                        label: ext.label,
                        active: ext.active,
                        version: ext.version,
                        latestVersion: ext.latestVersion,
                        installed: ext.installed,
                        ratingAverage: ext.ratingAverage,
                        storeLink: ext.storeLink,
                        changelog: ext.changelog,
                        installedAt: ext.installedAt ? new Date(ext.installedAt) : null,
                    });
                    extensionCount++;
                }
            }

            // Insert scheduled tasks
            if (scrapeInfo.scheduledTask && scrapeInfo.scheduledTask.length > 0) {
                for (const task of scrapeInfo.scheduledTask) {
                    await pg.insert(pgSchema.shopScheduledTask).values({
                        shopId: shop.id,
                        taskId: task.id,
                        name: task.name,
                        status: task.status,
                        interval: task.interval,
                        overdue: task.overdue,
                        lastExecutionTime: task.lastExecutionTime ? new Date(task.lastExecutionTime) : null,
                        nextExecutionTime: task.nextExecutionTime ? new Date(task.nextExecutionTime) : null,
                    });
                    taskCount++;
                }
            }

            // Insert queue info
            if (scrapeInfo.queueInfo && scrapeInfo.queueInfo.length > 0) {
                for (const q of scrapeInfo.queueInfo) {
                    await pg.insert(pgSchema.shopQueueInfo).values({
                        shopId: shop.id,
                        name: q.name,
                        size: q.size,
                    });
                    queueCount++;
                }
            }

            // Insert checks
            if (scrapeInfo.checks && scrapeInfo.checks.length > 0) {
                for (const check of scrapeInfo.checks) {
                    await pg.insert(pgSchema.shopCheck).values({
                        shopId: shop.id,
                        checkId: check.id,
                        level: check.level,
                        message: check.message,
                        source: check.source,
                        link: check.link,
                    });
                    checkCount++;
                }
            }
        }

        console.log(`Migrated scrape info for ${scrapeInfoCount} shops`);
        console.log(`  - ${extensionCount} extensions`);
        console.log(`  - ${taskCount} scheduled tasks`);
        console.log(`  - ${queueCount} queue info records`);
        console.log(`  - ${checkCount} checks`);

        console.log('\n=== Migration completed successfully! ===');

    } catch (error) {
        console.error('Migration failed:', error);
        throw error;
    } finally {
        sqlite.close();
        await pgClient.end();
    }
}

// Run the migration
migrate().catch((error) => {
    console.error('Migration failed:', error);
    process.exit(1);
});
