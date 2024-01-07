import { SimpleShop, HttpClient } from "@friendsofshopware/app-server-sdk"
import { Drizzle, getConnection, schema } from "../db";
import versionCompare from 'version-compare'
import { Extension, ExtensionDiff, ExtensionChangelog, lastUpdated } from "../../../shared/shop";
import promiseAllProperties from '../helper/promise'
import { CheckerInput, CheckerRegistery } from "./status/registery";
import { createSentry } from "../toucan";
import Shops from "../repository/shops";
import { UserSocketHelper } from "./UserSocket";
import { decrypt } from "../crypto";
import { eq, and } from 'drizzle-orm';

interface SQLShop {
    id: number;
    name: string;
    team_id: number;
    url: string;
    shopware_version: string;
    client_id: string;
    client_secret: string;
    ignores: string | string[];
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
    name: string,
    size: number
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
            date: string
        }
    }[];
}

const SECONDS = 1000;
const MINUTES = 60 * SECONDS;

export class ShopScrape implements DurableObject {
    state: DurableObjectState;
    env: Env;

    constructor(state: DurableObjectState, env: Env) {
        this.env = env;
        this.state = state;
    }

    async fetch(request: Request): Promise<Response> {
        const url = new URL(request.url);
        const id = url.searchParams.get('id')

        await this.state.storage.put('id', id)

        if (url.pathname === '/cron') {
            const currentAlarm = await this.state.storage.getAlarm();
            if (currentAlarm === null) {
                await this.state.storage.setAlarm(Date.now() + 60 * MINUTES);
                console.log(`Set alarm for shop ${id} to one hour`)
            }

            return new Response('OK');
        } else if (url.pathname === '/now') {
            await this.state.storage.setAlarm(Date.now() + 5 * SECONDS);
            console.log(`Set alarm for shop ${id} to 5 seconds`)

            if (url.searchParams.has('userId')) {
                await this.state.storage.put('triggeredBy', url.searchParams.get('userId'))
            }

            return new Response('OK');
        } else if (url.pathname === '/delete') {
            await this.state.storage.deleteAll();

            return new Response('OK');
        }

        return new Response('', { status: 404 });
    }

    async alarm(): Promise<void> {
        const con = getConnection(this.env);

        const id = await this.state.storage.get('id') as string | undefined

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const shop = await con.query.shop.findFirst({
            columns: {
                id: true,
                ignores: true,
                url: true,
                client_id: true,
                client_secret: true,
            },
            where: eq(schema.shop.id, parseInt(id))
        }) as SQLShop | undefined;


        // Shop is missing, so we can't do anything
        if (shop === undefined) {
            console.log(`cannot find shop: ${id}. Destroy self`)
            await this.state.storage.deleteAll();
            return;
        }

        shop.ignores = JSON.parse(shop.ignores as string || '[]')

        try {
            await this.updateShop(shop, con);
        } catch (e) {
            const sentry = createSentry(this.state, this.env);

            sentry.setExtra('shopId', id);
            sentry.captureException(e);
        }

        console.log(`Updated shop ${id}`)

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
                        team_id: shop.team_id,
                    }
                }
            )
        }
    }

    async updateShop(shop: SQLShop, con: Drizzle) {
        await con.update(schema.shop).set({ last_scraped_at: new Date().toUTCString() }).where(eq(schema.shop.id, shop.id)).execute();

        const clientSecret = await decrypt(this.env.APP_SECRET, shop.client_secret);

        const apiShop = new SimpleShop('', shop.url, '');
        apiShop.setShopCredentials(shop.client_id, clientSecret);
        const client = new HttpClient(apiShop);

        try {
            await client.getToken();
        } catch (e: any) {
            let error = "";

            if (typeof e?.response.body === "object") {
                if (e?.response.body.errors) {
                    error = JSON.stringify(e?.response.body.errors[0]);
                } else {
                    error = JSON.stringify(e?.response.body);
                }
            }

            await Shops.notify(
                con,
                this.env.USER_SOCKET,
                shop.id,
                `shop.update-auth-error.${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be updated`,
                    message: 'Could not connect to shop. Please check your credentials and try again.' + error,
                    link: { name: 'account.shops.detail', params: { shopId: shop.id.toString(), teamId: shop.team_id.toString() } }
                }
            );

            await Shops.alert(
                con,
                this.env,
                {
                    key: `shop.update-auth-error.${shop.id}`,
                    shopId: shop.id.toString(),
                    subject: 'Shop Update Error',
                    message: 'The Shop could not be updated. Please check your credentials and try again.' + error

                }
            )

            await con.update(schema.shop).set({ status: 'red' }).where(eq(schema.shop.id, shop.id)).execute();

            return;
        }

        let responses;

        try {
            responses = await promiseAllProperties({
                config: client.get('/_info/config'),
                plugin: client.post('/search/plugin'),
                app: client.post('/search/app'),
                scheduledTask: client.post('/search/scheduled-task'),
                queue: versionCompare(shop.shopware_version, '6.4.7.0') < 0 ? client.post('/search/message-queue-stats') : client.get('/_info/queue.json'),
                cacheInfo: client.get('/_action/cache_info')
            })
            // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (e: any) {
            let error = e;

            if (e.response) {
                error = `Request failed with status code: ${e.response.statusCode}: Body: ${JSON.stringify(e.response.body)}`;
            }

            await Shops.notify(
                con,
                this.env.USER_SOCKET,
                shop.id,
                `shop.not.updated_${shop.id}`,
                {
                    level: 'error',
                    title: `Shop: ${shop.name} could not be updated`,
                    message: `Could not connect to shop. Please check your credentials and try again.`,
                    link: { name: 'account.shops.detail', params: { shopId: shop.id.toString(), teamId: shop.team_id.toString() } }
                }
            )

            await con.update(schema.shop).set({ status: 'red', last_scraped_error: error }).where(eq(schema.shop.id, shop.id)).execute();

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
            } as Extension)
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
            } as Extension)
        }

        const scheduledTasks = responses.scheduledTask.body.data.map((task: ShopwareScheduledTask) => {
            return {
                id: task.id,
                name: task.name,
                status: task.status,
                interval: task.runInterval,
                overdue: new Date(task.nextExecutionTime).getTime() < new Date().getTime(),
                lastExecutionTime: task.lastExecutionTime,
                nextExecutionTime: task.nextExecutionTime,
            };
        });

        if (extensions.length) {
            const url = new URL('https://api.shopware.com/pluginStore/pluginsByName')
            url.searchParams.set('locale', 'en-GB');
            url.searchParams.set('shopwareVersion', shop.shopware_version);

            for (const extension of extensions) {
                url.searchParams.append('technicalNames[]', extension.name);
            }

            const storeResp = await fetch(url.toString(), {
                cf: {
                    cacheTtl: 60 * 60 * 6 // 6 hours
                }
            })

            if (storeResp.ok) {
                const storePlugins = await storeResp.json() as ShopwareStoreExtension[];

                for (const extension of extensions) {
                    const storePlugin = storePlugins.find((plugin: ShopwareStoreExtension) => plugin.name === extension.name);

                    if (storePlugin) {
                        extension.latestVersion = storePlugin.version;
                        extension.ratingAverage = storePlugin.ratingAverage;
                        extension.storeLink = storePlugin.link.replace('http://store.shopware.com:80', 'https://store.shopware.com');

                        if (storePlugin.latestVersion != extension.version) {
                            const changelogs: ExtensionChangelog[] = [];

                            for (const changelog of storePlugin.changelog) {
                                if (versionCompare(changelog.version, extension.version) > 0) {
                                    const compare = versionCompare(changelog.version, extension.latestVersion);

                                    changelogs.push({
                                        version: changelog.version,
                                        text: changelog.text,
                                        creationDate: changelog.creationDate.date,
                                        isCompatible: compare < 0 || compare === 0,
                                    } as ExtensionChangelog)
                                }
                            }

                            extension.changelog = changelogs;
                        }
                    }
                }
            }
        }

        const resultCurrentExtensions = await con.query.shopScrapeInfo.findFirst({
            columns: {
                extensions: true,
            },
            where: eq(schema.shopScrapeInfo.shop, shop.id)
        })

        const extensionsDiff: ExtensionDiff[] = [];
        if (resultCurrentExtensions) {

            const currentExtensions = JSON.parse(resultCurrentExtensions.extensions);

            for (const oldExtension of currentExtensions) {
                let exists = false;

                for (const newExtension of extensions) {
                    if (oldExtension.name === newExtension.name) {
                        let state: string | false = false;
                        let changelog: ExtensionChangelog[] | null = null;
                        if (oldExtension.version !== newExtension.version) {
                            state = 'updated';
                            changelog = oldExtension.changelog;
                        }
                        else if (oldExtension.active === true && newExtension.active === false) {
                            state = 'deactivated';
                        }
                        else if (oldExtension.active === false && newExtension.active === true) {
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
                                active: newExtension.active
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
                        active: oldExtension.active
                    });
                }
            }

            for (const newExtension of extensions) {
                let exists = false;

                for (const oldExtension of currentExtensions) {
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
                        active: newExtension.active
                    });
                }
            }
        }

        const shopUpdate: lastUpdated = {};

        if (shop.shopware_version !== responses.config.body.version) {
            shopUpdate.from = shop.shopware_version,
                shopUpdate.to = responses.config.body.version;
            shopUpdate.date = new Date().toISOString();
        }

        const favicon = `https://t2.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=${shop.url}&size=32`;

        let queue = null;
        if (versionCompare(shop.shopware_version, '6.4.7.0') < 0) {
            queue = responses.queue.body.data.filter((entry: ShopwareQueue) => entry.size > 0);
        } else {
            queue = responses.queue.body.filter((entry: ShopwareQueue) => entry.size > 0);
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
        }

        const checkerResult = await CheckerRegistery.check(input);

        await con.update(schema.shop).set({
            status: checkerResult.status,
            shopware_version: responses.config.body.version,
            favicon: favicon,
            last_scraped_error: null
        }).where(eq(schema.shop.id, shop.id)).execute();

        // delete
        await con.delete(schema.shopScrapeInfo).where(eq(schema.shopScrapeInfo.shop, shop.id)).execute();
        await con.insert(schema.shopScrapeInfo).values({
            shop: shop.id,
            extensions: JSON.stringify(input.extensions),
            scheduled_task: JSON.stringify(input.scheduledTasks),
            queue_info: JSON.stringify(input.queueInfo),
            cache_info: JSON.stringify(input.cacheInfo),
            checks: JSON.stringify(checkerResult.checks),
            created_at: new Date().toISOString(),
        })

        const hasShopUpdate = Object.keys(shopUpdate).length !== 0;

        if (extensionsDiff.length > 0 || hasShopUpdate) {
            const oldShopwareVersion = hasShopUpdate ? shopUpdate.from : null;
            const newShopwareVersion = hasShopUpdate ? shopUpdate.to : null;

            await con
                .insert(schema.shopChangelog)
                .values({
                    shop_id: shop.id,
                    extensions: JSON.stringify(extensionsDiff),
                    old_shopware_version: oldShopwareVersion,
                    new_shopware_version: newShopwareVersion,
                    date: new Date().toISOString(),
                }).execute();
        }

        if (hasShopUpdate) {
            await con.update(schema.shop).set({ last_updated: new Date().toISOString() }).where(eq(schema.shop.id, shop.id)).execute();
        }
    }
}
