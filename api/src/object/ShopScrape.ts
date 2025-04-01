import {
    SimpleShop,
    HttpClient,
    ApiClientRequestFailed,
    ApiClientAuthenticationFailed,
    HttpClientResponse,
} from '@shopware-ag/app-server-sdk';
import { Drizzle, getConnection, schema } from '../db';
import versionCompare from 'version-compare';
import { CheckerInput, check } from './status/registery';
import { createSentry } from '../toucan';
import Shops, { User } from '../repository/shops';
import { UserSocketHelper } from './UserSocket';
import { decrypt } from '../crypto';
import { and, asc, eq, inArray } from 'drizzle-orm';
import type { Bindings } from '../router';
import { Extension, ExtensionChangelog, ExtensionDiff } from '../types';

interface SQLShop {
    id: number;
    name: string;
    status: string;
    organizationId: number;
    url: string;
    clientId: string;
    clientSecret: string;
    shopwareVersion: string;
    ignores: string[];
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

const SECONDS = 1000;
const MINUTES = 60 * SECONDS;
const STATES: { [key: string]: number } = {
    green: 1,
    yellow: 2,
    red: 3,
};

export class ShopScrape implements DurableObject {
    state: DurableObjectState;
    env: Bindings;

    constructor(state: DurableObjectState, env: Bindings) {
        this.env = env;
        this.state = state;
    }

    async fetch(request: Request): Promise<Response> {
        const url = new URL(request.url);
        const id = url.searchParams.get('id');

        await this.state.storage.put('id', id);

        if (url.pathname === '/cron') {
            const currentAlarm = await this.state.storage.getAlarm();
            if (currentAlarm === null) {
                await this.state.storage.setAlarm(Date.now() + 60 * MINUTES);
                console.log(`Set alarm for shop ${id} to one hour`);
            }

            return new Response('OK');
        }
        if (url.pathname === '/now') {
            await this.state.storage.setAlarm(Date.now() + 5 * SECONDS);
            console.log(`Set alarm for shop ${id} to 5 seconds`);

            if (url.searchParams.has('userId')) {
                await this.state.storage.put(
                    'triggeredBy',
                    url.searchParams.get('userId'),
                );
            }

            return new Response('OK');
        }
        if (url.pathname === '/delete') {
            await this.state.storage.deleteAll();

            return new Response('OK');
        }

        return new Response('', { status: 404 });
    }

    async alarm(): Promise<void> {
        const con = getConnection(this.env);

        const id = (await this.state.storage.get('id')) as string | undefined;

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const shop = await con.query.shop.findFirst({
            columns: {
                id: true,
                name: true,
                status: true,
                ignores: true,
                url: true,
                clientId: true,
                clientSecret: true,
                shopwareVersion: true,
                organizationId: true,
            },
            where: eq(schema.shop.id, parseInt(id)),
        });

        // Shop is missing, so we can't do anything
        if (shop === undefined) {
            console.log(`cannot find shop: ${id}. Destroy self`);
            await this.state.storage.deleteAll();
            return;
        }

        try {
            await this.updateShop(shop, con);
        } catch (e) {
            const sentry = createSentry(this.state, this.env);

            sentry.setExtra('shopId', id);
            sentry.captureException(e);
        }

        console.log(`Updated shop ${id}`);

        this.state.storage.setAlarm(Date.now() + 60 * MINUTES);

        const triggeredBy = await this.state.storage.get<string>('triggeredBy');
        if (triggeredBy !== undefined) {
            await this.state.storage.delete('triggeredBy');

            await UserSocketHelper.sendNotification(
                this.env.USER_SOCKET,
                triggeredBy,
                {
                    shopUpdate: {
                        id: shop.id,
                        organizationId: shop.organizationId,
                    },
                },
            );
        }
    }

    async shouldNotify(
        users: User[],
        notificationKey: string,
    ): Promise<boolean> {
        const con = getConnection(this.env);
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
        const timeDifference =
            lastNotification.getTime() - new Date().getTime();

        return timeDifference >= 24 * 60 * 60 * 1000;
    }

    async updateShop(shop: SQLShop, con: Drizzle) {
        await con
            .update(schema.shop)
            .set({ lastScrapedAt: new Date() })
            .where(eq(schema.shop.id, shop.id))
            .execute();

        const clientSecret = await decrypt(
            this.env.APP_SECRET,
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
                error = `Request failed with status code: ${
                    e.response.statusCode
                }: Body: ${JSON.stringify(e.response.body)}`;
            } else if (e instanceof ApiClientAuthenticationFailed) {
                error = `Authentication failed with status code: ${
                    e.response.statusCode
                }: Body: ${JSON.stringify(e.response.body)}`;
            }

            await Shops.notify(
                con,
                this.env.USER_SOCKET,
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
                            organizationId: shop.organizationId.toString(),
                        },
                    },
                },
            );

            await Shops.alert(con, this.env, {
                key: `shop.update-auth-error.${shop.id}`,
                shopId: shop.id.toString(),
                subject: 'Shop Update Error',
                message: `The Shop could not be updated. Please check your credentials and try again.${error}`,
            });

            await con
                .update(schema.shop)
                .set({ status: 'red' })
                .where(eq(schema.shop.id, shop.id))
                .execute();

            return;
        }

        let responses: {
            config: HttpClientResponse;
            plugin: HttpClientResponse;
            app: HttpClientResponse;
            scheduledTask: HttpClientResponse;
            queue: HttpClientResponse;
            cacheInfo: HttpClientResponse;
        };

        try {
            const [config, plugin, app, scheduledTask, queue, cacheInfo] =
                await Promise.all([
                    client.get('/_info/config'),
                    client.post('/search/plugin'),
                    client.post('/search/app'),
                    client.post('/search/scheduled-task'),
                    versionCompare(shop.shopwareVersion, '6.4.7.0') < 0
                        ? client.post('/search/message-queue-stats')
                        : client.get('/_info/queue.json'),
                    client.get('/_action/cache_info'),
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

            await Shops.notify(
                con,
                this.env.USER_SOCKET,
                shop.id,
                `shop.not.updated_${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be updated`,
                    message:
                        'Could not connect to shop. Please check your credentials and try again.',
                    link: {
                        name: 'account.shops.detail',
                        params: {
                            shopId: shop.id.toString(),
                            organizationId: shop.organizationId.toString(),
                        },
                    },
                },
            );

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
                installed: true, // When it's in DB its always installed
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
                        new Date(task.nextExecutionTime).getTime() <
                        new Date().getTime(),
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

            const storeResp = await fetch(url.toString(), {
                cf: {
                    cacheTtl: 60 * 60 * 6, // 6 hours
                },
            });

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

        const resultCurrentExtensions =
            await con.query.shopScrapeInfo.findFirst({
                columns: {
                    extensions: true,
                },
                where: eq(schema.shopScrapeInfo.shopId, shop.id),
            });

        const extensionsDiff: ExtensionDiff[] = [];
        if (resultCurrentExtensions) {
            for (const oldExtension of resultCurrentExtensions.extensions) {
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

                for (const oldExtension of resultCurrentExtensions.extensions) {
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

        const favicon = `https://t2.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=${shop.url}&size=32`;

        let queue = null;
        if (versionCompare(shop.shopwareVersion, '6.4.7.0') < 0) {
            queue = responses.queue.body.data.filter(
                (entry: ShopwareQueue) => entry.size > 0,
            );
        } else {
            queue = responses.queue.body.filter(
                (entry: ShopwareQueue) => entry.size > 0,
            );
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

            if (await this.shouldNotify(users, statusChangeKey)) {
                await Shops.notify(
                    con,
                    this.env.USER_SOCKET,
                    shop.id,
                    statusChangeKey,
                    {
                        level: 'warning',
                        title: `Shop: ${shop.name} status changed`,
                        message: `Status changed from ${shop.status} to ${checkerResult.status}`,
                        link: {
                            name: 'account.shops.detail',
                            params: {
                                shopId: shop.id.toString(),
                                organizationId: shop.organizationId.toString(),
                            },
                        },
                    },
                );

                await Shops.alert(con, this.env, {
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

        // delete
        await con
            .delete(schema.shopScrapeInfo)
            .where(eq(schema.shopScrapeInfo.shopId, shop.id))
            .execute();

        await con.insert(schema.shopScrapeInfo).values({
            shopId: shop.id,
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
    }
}
