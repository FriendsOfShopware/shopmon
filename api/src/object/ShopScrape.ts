import { Connection } from "@planetscale/database/dist";
import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../db";
import versionCompare from 'version-compare'
import { Extension, ExtensionChangelog } from "../../../shared/shop";
import promiseAllProperties from '../helper/promise'
import { CheckerInput, CheckerRegistery } from "./status/registery";
import { createSentry } from "../sentry";
import Shops from "../repository/shops";
import { UserSocketHelper } from "./UserSocket";
import { decrypt } from "../crypto";
 
interface SQLShop {
    id: string;
    name: string;
    team_id: string;
    url: string;
    shopware_version: string;
    client_id: string;
    client_secret: string;
    ignores: string|string[];
}

interface ShopwareScheduledTask {
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

        const id = await this.state.storage.get('id')

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const fetchShopSQL = 'SELECT id, name, url, shopware_version, client_id, client_secret, team_id, ignores FROM shop WHERE id = ?';

        const shops = await con.execute(fetchShopSQL, [id]);

        // Shop is missing, so we can't do anything
        if (shops.rows.length === 0) {
            console.log(`cannot find shop: ${id}. Destroy self`)
            await this.state.storage.deleteAll();
            return;
        }

        const shop = shops.rows[0] as SQLShop;
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
                        id: parseInt(shop.id),
                        team_id: parseInt(shop.team_id),
                    }
                }
            )
        }
    }

    async updateShop(shop: SQLShop, con: Connection) {
    
        await con.execute('UPDATE shop SET last_scraped_at = NOW() WHERE id = ?', [shop.id]);

        const clientSecret = await decrypt(this.env.APP_SECRET, shop.client_secret);
    
        const client = new HttpClient(new Shop('', shop.url, '', shop.client_id, clientSecret));
    
        try {
            await client.getToken();
        } catch (e) {
            await Shops.notify(
                con, 
                this.env.USER_SOCKET, 
                shop.id, 
                `shop.update-auth-error.${shop.id}`,
                {
                    level: 'error', 
                    title: `Shop: ${shop.name} could not be updated`, 
                    message: 'Could not connect to shop. Please check your credentials and try again.', 
                    link: { name: 'account.shops.detail', params: { shopId: shop.id, teamId: shop.team_id } }
                }
            );

            await Shops.alert(
                con,
                this.env,
                {
                    key: `shop.update-auth-error.${shop.id}`,
                    shopId: shop.id,
                    subject: 'Shop Update Error',
                    message: 'The Shop could not be updated. Please check your credentials and try again.'

                }
            )

            await con.execute('UPDATE shop SET status = ? WHERE id = ?', [
                'red',
                shop.id,
            ]);
            return;
        }
    
        let responses;
    
        try {
            responses = await promiseAllProperties({
                config: client.get('/_info/config'),
                plugin: client.post('/search/plugin'),
                app: client.post('/search/app'),
                scheduledTask: client.post('/search/scheduled-task'),
                queue: client.get('/_info/queue.json'),
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
                    link: { name: 'account.shops.detail', params: { shopId: shop.id, teamId: shop.team_id } }
                }
            )
    
            await con.execute('UPDATE shop SET status = ?, last_scraped_error = ? WHERE id = ?', [
                'red',
                error,
                shop.id,
            ]);
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
    
        const favicon = `https://t2.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=${shop.url}&size=32`;

        const queue = responses.queue.body.filter((entry: ShopwareQueue) => entry.size > 0);

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
    
        await con.execute('UPDATE shop SET status = ?, shopware_version = ?, favicon = ?, last_scraped_error = null WHERE id = ?', [
            checkerResult.status,
            responses.config.body.version,
            favicon,
            shop.id,
        ]);
    
        await con.execute('REPLACE INTO shop_scrape_info(shop_id, extensions, scheduled_task, queue_info, cache_info, checks, created_at) VALUES(?, ?, ?, ?, ?, ?, NOW())', [
            shop.id,
            JSON.stringify(input.extensions),
            JSON.stringify(input.scheduledTasks),
            JSON.stringify(input.queueInfo),
            JSON.stringify(input.cacheInfo),
            JSON.stringify(checkerResult.checks),
            favicon
        ]);
    }
}
