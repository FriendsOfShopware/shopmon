import { Connection } from "@planetscale/database/dist";
import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../db";
import versionCompare from 'version-compare'
import { Extension, ExtensionDiff, ExtensionChangelog, lastUpdated } from "../../../shared/shop";
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
                    message: 'The Shop could not be updated. Please check your credentials and try again.' + error

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

        const resultCurrentExtensions = await con.execute('SELECT extensions FROM shop_scrape_info WHERE shop_id = ?', [
            shop.id
        ]);

        const extensionsDiff: ExtensionDiff[] = [];
        if (resultCurrentExtensions.rows.length > 0) {
            
            const currentExtensions = JSON.parse(resultCurrentExtensions.rows[0].extensions);

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
                        
                        if(state) {
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
            shopUpdate.date = new Date().toISOString().slice(0, 19).replace('T', ' ');
        }
        
        const favicon = `https://t2.gstatic.com/faviconV2?client=SOCIAL&type=FAVICON&fallback_opts=TYPE,SIZE,URL&url=${shop.url}&size=32`;

        let queue = null;
        if ( versionCompare(shop.shopware_version, '6.4.7.0') < 0 ) {
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

        const hasShopUpdate = Object.keys(shopUpdate).length !== 0;

        if (extensionsDiff.length > 0 || hasShopUpdate) {
            const oldShopwareVersion = hasShopUpdate ? shopUpdate.from : null;
            const newShopwareVersion = hasShopUpdate ? shopUpdate.to : null;

            await con.execute(
                'INSERT INTO shop_changelog(shop_id, extensions, old_shopware_version, new_shopware_version, date) VALUES(?, ?, ?, ?, NOW())', [
                    shop.id, 
                    JSON.stringify(extensionsDiff), 
                    oldShopwareVersion, 
                    newShopwareVersion
                ]
            );
        }

        if (hasShopUpdate) {
            await con.execute('UPDATE shop SET last_updated = ? WHERE id = ?', [
                JSON.stringify(shopUpdate),
                shop.id,
            ]);
        }
    }
}
