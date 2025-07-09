import { logger } from '@sentry/node';
import {
    ApiClientAuthenticationFailed,
    ApiClientRequestFailed,
    HttpClient,
    type HttpClientResponse,
    SimpleShop,
} from '@shopware-ag/app-server-sdk';
import { and, asc, eq, inArray } from 'drizzle-orm';
import { decrypt } from '../../crypto/index.ts';
import { type Drizzle, getConnection, schema } from '../../db.ts';
import { type CheckerInput, check } from '../../object/status/registery.ts';
import {
    getShopScrapeInfo,
    saveShopScrapeInfo,
} from '../../repository/scrapeInfo.ts';
import Shops, { type User } from '../../repository/shops.ts';
import type {
    CacheInfo,
    Extension,
    ExtensionChangelog,
    ExtensionDiff,
    QueueInfo,
} from '../../types/index.ts';
import versionCompare from '../../util.ts';

interface SQLShop {
    id: number;
    name: string;
    status: string;
    organizationId: string;
    organizationSlug: string;
    url: string;
    clientId: string;
    clientSecret: string;
    shopwareVersion: string;
    ignores: string[];
    connectionIssueCount: number;
}

interface ShopwareScheduledTask {
    id: string;
    name: string;
    runInterval: number;
    status: string;
    overdue: boolean;
    lastExecutionTime: string;
    nextExecutionTime: string;
}

interface ShopwareQueue {
    name: string;
    size: number;
}

interface ShopwareStoreExtension {
    name: string;
    ratingAverage: number;
    link: string;
    version: string;
    latestVersion: string;
    changelog: {
        version: string;
        text: string;
        creationDate: {
            date: string;
        };
    }[];
}

interface ShopwarePlugin {
    name: string;
    label: string;
    active: boolean;
    version: string;
    upgradeVersion: string | null;
    installedAt: string | null;
}

interface ShopwareApp {
    name: string;
    label: string;
    active: boolean;
    version: string;
    createdAt: string;
}

interface ShopwareConfig {
    version: string;
    adminWorker: {
        enableAdminWorker: boolean;
    };
}

const STATES: { [key: string]: number } = {
    green: 1,
    yellow: 2,
    red: 3,
};

export async function shopScrapeJob() {
    const con = getConnection();

    const shops = await con
        .select({
            id: schema.shop.id,
            name: schema.shop.name,
            status: schema.shop.status,
            ignores: schema.shop.ignores,
            url: schema.shop.url,
            clientId: schema.shop.clientId,
            clientSecret: schema.shop.clientSecret,
            shopwareVersion: schema.shop.shopwareVersion,
            organizationId: schema.shop.organizationId,
            organizationSlug: schema.organization.slug,
            shopImage: schema.shop.shopImage,
            connectionIssueCount: schema.shop.connectionIssueCount,
        })
        .from(schema.shop)
        .innerJoin(
            schema.organization,
            eq(schema.organization.id, schema.shop.organizationId),
        )
        .all();

    console.log(`Found ${shops.length} shops to scrape`);

    // Process shops in parallel with a limit
    const batchSize = 10;
    for (let i = 0; i < shops.length; i += batchSize) {
        const batch = shops.slice(i, i + batchSize);
        await Promise.all(batch.map((shop) => updateShop(shop, con)));
    }
}

export async function scrapeSingleShop(shopId: number) {
    const con = getConnection();

    const shop = await con
        .select({
            id: schema.shop.id,
            name: schema.shop.name,
            status: schema.shop.status,
            ignores: schema.shop.ignores,
            url: schema.shop.url,
            clientId: schema.shop.clientId,
            clientSecret: schema.shop.clientSecret,
            shopwareVersion: schema.shop.shopwareVersion,
            organizationId: schema.shop.organizationId,
            organizationSlug: schema.organization.slug,
            shopImage: schema.shop.shopImage,
            connectionIssueCount: schema.shop.connectionIssueCount,
        })
        .from(schema.shop)
        .innerJoin(
            schema.organization,
            eq(schema.organization.id, schema.shop.organizationId),
        )
        .where(eq(schema.shop.id, shopId))
        .get();

    if (!shop) {
        throw new Error(`Shop with ID ${shopId} not found`);
    }

    await updateShop(shop, con);
}

async function shouldNotify(
    con: Drizzle,
    users: User[],
    notificationKey: string,
): Promise<boolean> {
    const notificationResult = await con
        .select({
            created_at: schema.userNotification.createdAt,
        })
        .from(schema.userNotification)
        .where(
            and(
                inArray(
                    schema.userNotification.userId,
                    users.map((user) => user.id),
                ),
                eq(schema.userNotification.key, notificationKey),
                eq(schema.userNotification.read, false),
            ),
        )
        .orderBy(asc(schema.userNotification.createdAt))
        .all();

    if (notificationResult.length === 0) {
        return true;
    }

    const lastNotification = new Date(notificationResult[0].created_at);
    const timeDifference = lastNotification.getTime() - Date.now();

    return timeDifference >= 24 * 60 * 60 * 1000;
}

async function updateShop(shop: SQLShop, con: Drizzle) {
    if (shop.connectionIssueCount >= 3) {
        logger.info(
            `Shop ${shop.name} has too many connection issues, skipping it`,
        );
        return;
    }

    try {
        await con
            .update(schema.shop)
            .set({ lastScrapedAt: new Date() })
            .where(eq(schema.shop.id, shop.id))
            .execute();

        const clientSecret = await decrypt(
            process.env.APP_SECRET,
            shop.clientSecret,
        );

        const apiShop = new SimpleShop('', shop.url, '');
        apiShop.setShopCredentials(shop.clientId, clientSecret);
        const client = new HttpClient(apiShop);

        try {
            await client.getToken();
        } catch (e) {
            let error = e;

            if (e instanceof ApiClientRequestFailed) {
                let body = JSON.stringify(e.response.body);

                if (body.length > 50) {
                    body = `${body.substring(0, 50)}...`;
                }

                error = `Request failed with status code: ${
                    e.response.statusCode
                }: Body: ${body}`;
            } else if (e instanceof ApiClientAuthenticationFailed) {
                let body = JSON.stringify(e.response.body);

                if (body.length > 50) {
                    body = `${body.substring(0, 50)}...`;
                }

                error = `Authentication failed with status code: ${
                    e.response.statusCode
                }: Body: ${body}`;
            }

            await Shops.notify(
                con,
                shop.id,
                `shop.update-auth-error.${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be updated`,
                    message: `Could not connect to shop. Please check your credentials and try again.${error}`,
                    link: {
                        name: 'account.shops.detail',
                        params: {
                            shopId: shop.id.toString(),
                            slug: shop.organizationSlug,
                        },
                    },
                },
            );

            await Shops.alert(con, {
                key: `shop.update-auth-error.${shop.id}`,
                shopId: shop.id.toString(),
                subject: 'Refresh shop data error',
                message: `The shop data could not be refreshed. Please check your credentials and try again. ${error}`,
            });

            logger.info(
                `Shop ${shop.name} could not be refreshed, error is ${error}`,
            );

            await con
                .update(schema.shop)
                .set({
                    status: 'red',
                    connectionIssueCount: shop.connectionIssueCount + 1,
                })
                .where(eq(schema.shop.id, shop.id))
                .execute();

            return;
        }

        if (shop.connectionIssueCount > 0) {
            await con
                .update(schema.shop)
                .set({ connectionIssueCount: 0 })
                .where(eq(schema.shop.id, shop.id))
                .execute();
        }

        let responses: {
            config: HttpClientResponse<ShopwareConfig>;
            plugin: HttpClientResponse<{ data: ShopwarePlugin[] }>;
            app: HttpClientResponse<{ data: ShopwareApp[] }>;
            scheduledTask: HttpClientResponse<{
                data: ShopwareScheduledTask[];
            }>;
            queue: HttpClientResponse<
                ShopwareQueue[] | { data: ShopwareQueue[] }
            >;
            cacheInfo: HttpClientResponse<CacheInfo>;
        };

        try {
            const [config, plugin, app, scheduledTask, queue, cacheInfo] =
                await Promise.all([
                    client.get<ShopwareConfig>('/_info/config'),
                    client.post<{ data: ShopwarePlugin[] }>('/search/plugin'),
                    client.post<{ data: ShopwareApp[] }>('/search/app'),
                    client.post<{ data: ShopwareScheduledTask[] }>(
                        '/search/scheduled-task',
                    ),
                    versionCompare(shop.shopwareVersion, '6.4.7.0') < 0
                        ? client.post<{ data: ShopwareQueue[] }>(
                              '/search/message-queue-stats',
                          )
                        : client.get<ShopwareQueue[]>('/_info/queue.json'),
                    client.get<CacheInfo>('/_action/cache_info'),
                ]);

            responses = {
                config,
                plugin,
                app,
                scheduledTask,
                queue,
                cacheInfo,
            };
        } catch (e) {
            let error = '';

            if (e instanceof ApiClientRequestFailed) {
                error = `Request failed with status code: ${
                    e.response.statusCode
                }: Body: ${JSON.stringify(e.response.body)}`;
            } else if (e instanceof Error) {
                error = e.toString();
            }

            console.log(error);

            await Shops.notify(con, shop.id, `shop.not.updated_${shop.id}`, {
                level: 'error',
                title: `Shop: ${shop.name} could not be updated`,
                message:
                    'Could not connect to shop. Please check your credentials and try again.',
                link: {
                    name: 'account.shops.detail',
                    params: {
                        shopId: shop.id.toString(),
                        slug: shop.organizationSlug,
                    },
                },
            });

            await con
                .update(schema.shop)
                .set({ status: 'red', lastScrapedError: error.toString() })
                .where(eq(schema.shop.id, shop.id))
                .execute();

            return;
        }

        const extensions: Extension[] = [];

        for (const plugin of responses.plugin.body.data) {
            extensions.push({
                name: plugin.name,
                label: plugin.label,
                active: plugin.active,
                version: plugin.version,
                latestVersion: plugin.upgradeVersion,
                ratingAverage: null,
                storeLink: null,
                changelog: null,
                installed: plugin.installedAt !== null,
                installedAt: plugin.installedAt,
            } as Extension);
        }

        for (const app of responses.app.body.data) {
            extensions.push({
                name: app.name,
                label: app.label,
                active: app.active,
                version: app.version,
                latestVersion: null,
                ratingAverage: null,
                storeLink: null,
                changelog: null,
                installed: true,
                installedAt: app.createdAt,
            } as Extension);
        }

        const scheduledTasks = responses.scheduledTask.body.data.map(
            (task: ShopwareScheduledTask) => {
                return {
                    id: task.id,
                    name: task.name,
                    status: task.status,
                    interval: task.runInterval,
                    overdue:
                        new Date(task.nextExecutionTime).getTime() < Date.now(),
                    lastExecutionTime: task.lastExecutionTime,
                    nextExecutionTime: task.nextExecutionTime,
                };
            },
        );

        if (extensions.length) {
            const url = new URL(
                'https://api.shopware.com/pluginStore/pluginsByName',
            );
            url.searchParams.set('locale', 'en-GB');
            url.searchParams.set('shopwareVersion', shop.shopwareVersion);

            for (const extension of extensions) {
                url.searchParams.append('technicalNames[]', extension.name);
            }

            const storeResp = await fetch(url.toString());

            if (storeResp.ok) {
                const storePlugins =
                    (await storeResp.json()) as ShopwareStoreExtension[];

                for (const extension of extensions) {
                    const storePlugin = storePlugins.find(
                        (plugin: ShopwareStoreExtension) =>
                            plugin.name === extension.name,
                    );

                    if (storePlugin) {
                        extension.latestVersion = storePlugin.version;
                        extension.ratingAverage = storePlugin.ratingAverage;
                        extension.storeLink = storePlugin.link.replace(
                            'http://store.shopware.com:80',
                            'https://store.shopware.com',
                        );

                        if (storePlugin.latestVersion !== extension.version) {
                            const changelogs: ExtensionChangelog[] = [];

                            for (const changelog of storePlugin.changelog) {
                                if (
                                    versionCompare(
                                        changelog.version,
                                        extension.version,
                                    ) > 0
                                ) {
                                    const compare = versionCompare(
                                        changelog.version,
                                        extension.latestVersion,
                                    );

                                    changelogs.push({
                                        version: changelog.version,
                                        text: changelog.text,
                                        creationDate:
                                            changelog.creationDate.date,
                                        isCompatible:
                                            compare < 0 || compare === 0,
                                    } as ExtensionChangelog);
                                }
                            }

                            extension.changelog = changelogs;
                        }
                    }
                }
            }
        }

        const oldShopScrapeInfo = await getShopScrapeInfo(shop.id);

        const extensionsDiff: ExtensionDiff[] = [];
        if (oldShopScrapeInfo) {
            for (const oldExtension of oldShopScrapeInfo.extensions) {
                let exists = false;

                for (const newExtension of extensions) {
                    if (oldExtension.name === newExtension.name) {
                        let state: string | false = false;
                        let changelog: ExtensionChangelog[] | null = null;
                        if (oldExtension.version !== newExtension.version) {
                            state = 'updated';
                            changelog = oldExtension.changelog;
                        } else if (
                            oldExtension.active === true &&
                            newExtension.active === false
                        ) {
                            state = 'deactivated';
                        } else if (
                            oldExtension.active === false &&
                            newExtension.active === true
                        ) {
                            state = 'activated';
                        }

                        if (state) {
                            extensionsDiff.push({
                                name: newExtension.name,
                                label: newExtension.label,
                                state: state,
                                old_version: oldExtension.version,
                                new_version: newExtension.version,
                                changelog: changelog,
                                active: newExtension.active,
                            });
                        }

                        exists = true;
                    }
                }

                if (!exists) {
                    extensionsDiff.push({
                        name: oldExtension.name,
                        label: oldExtension.label,
                        state: 'removed',
                        old_version: oldExtension.version,
                        new_version: null,
                        changelog: null,
                        active: oldExtension.active,
                    });
                }
            }

            for (const newExtension of extensions) {
                let exists = false;

                for (const oldExtension of oldShopScrapeInfo.extensions) {
                    if (oldExtension.name === newExtension.name) {
                        exists = true;
                    }
                }

                if (!exists) {
                    extensionsDiff.push({
                        name: newExtension.name,
                        label: newExtension.label,
                        state: 'installed',
                        old_version: null,
                        new_version: newExtension.version,
                        changelog: null,
                        active: newExtension.active,
                    });
                }
            }
        }

        const shopUpdate: {
            date?: string;
            from?: string;
            to?: string;
        } = {};

        if (shop.shopwareVersion !== responses.config.body.version) {
            shopUpdate.from = shop.shopwareVersion;
            shopUpdate.to = responses.config.body.version;
            shopUpdate.date = new Date().toISOString();
        }

        const favicon = await getFavicon(shop.url);

        let queue: QueueInfo[] = [];
        if (versionCompare(shop.shopwareVersion, '6.4.7.0') < 0) {
            // For older versions, the response has a data property
            const queueData = responses.queue.body as { data: ShopwareQueue[] };
            queue = queueData.data.filter(
                (entry: ShopwareQueue) => entry.size > 0,
            );
        } else {
            // For newer versions, the response is directly an array
            const queueData = responses.queue.body as ShopwareQueue[];
            queue = queueData.filter((entry: ShopwareQueue) => entry.size > 0);
        }

        const input: CheckerInput = {
            extensions: extensions,
            config: responses.config.body,
            queueInfo: queue,
            scheduledTasks: scheduledTasks,
            cacheInfo: responses.cacheInfo.body,
            favicon: favicon,
            client,
            ignores: shop.ignores as string[],
        };

        const checkerResult = await check(input);

        if (STATES[shop.status] < STATES[checkerResult.status]) {
            const users = await Shops.getUsersOfShop(con, shop.id);
            const statusChangeKey = `shop.change-status.${shop.id}`;

            if (await shouldNotify(con, users, statusChangeKey)) {
                await Shops.notify(con, shop.id, statusChangeKey, {
                    level: 'warning',
                    title: `Shop: ${shop.name} status changed`,
                    message: `Status changed from ${shop.status} to ${checkerResult.status}`,
                    link: {
                        name: 'account.shops.detail',
                        params: {
                            shopId: shop.id.toString(),
                            slug: shop.organizationSlug,
                        },
                    },
                });

                await Shops.alert(con, {
                    key: statusChangeKey,
                    shopId: shop.id.toString(),
                    subject: `Shop ${shop.name} status changed to ${checkerResult.status}`,
                    message: `The Shop ${shop.name} has change its status from ${shop.status} to ${checkerResult.status}. Pleas visit Shopmon for details.`,
                });
            }
        }

        await con
            .update(schema.shop)
            .set({
                status: checkerResult.status,
                shopwareVersion: responses.config.body.version,
                favicon: favicon,
                lastScrapedError: null,
            })
            .where(eq(schema.shop.id, shop.id))
            .execute();

        await saveShopScrapeInfo(shop.id, {
            extensions: input.extensions,
            scheduledTask: input.scheduledTasks,
            queueInfo: input.queueInfo,
            cacheInfo: input.cacheInfo,
            checks: checkerResult.checks,
            createdAt: new Date(),
        });

        const hasShopUpdate = Object.keys(shopUpdate).length !== 0;

        if (extensionsDiff.length > 0 || hasShopUpdate) {
            const oldShopwareVersion = hasShopUpdate ? shopUpdate.from : null;
            const newShopwareVersion = hasShopUpdate ? shopUpdate.to : null;

            await con
                .insert(schema.shopChangelog)
                .values({
                    shopId: shop.id,
                    extensions: extensionsDiff,
                    oldShopwareVersion: oldShopwareVersion,
                    newShopwareVersion: newShopwareVersion,
                    date: new Date(),
                })
                .execute();
        }

        if (hasShopUpdate && shopUpdate.from && shopUpdate.to) {
            await con
                .update(schema.shop)
                .set({
                    lastChangelog: {
                        date: new Date(),
                        from: shopUpdate.from,
                        to: shopUpdate.to,
                    },
                })
                .where(eq(schema.shop.id, shop.id))
                .execute();
        }

        console.log(`Updated shop ${shop.id}`);
    } catch (e) {
        console.error(`Error updating shop ${shop.id}:`, e);
        throw e;
    }
}

async function getFavicon(url: string): Promise<string | null> {
    const shopHtml = await fetch(url, {
        redirect: 'follow',
    });

    if (!shopHtml.ok) {
        return null;
    }

    const text = await shopHtml.text();
    const match = text.match(
        /<link[^>]+rel=["']?(?:shortcut\s+)?icon["']?[^>]*>/i,
    );
    if (match) {
        const iconTag = match[0];
        const hrefMatch = iconTag.match(/href=["']([^"']+)["']/i);
        if (hrefMatch) {
            const iconUrl = hrefMatch[1];
            if (iconUrl.startsWith('http')) {
                return iconUrl;
            }
            if (iconUrl.startsWith('/')) {
                // If the URL is relative, construct the absolute URL
                const absoluteUrl = new URL(iconUrl, url);
                return absoluteUrl.toString();
            }
        }
    }

    return null;
}
