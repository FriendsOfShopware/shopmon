import { Connection } from "@planetscale/database/dist";
import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../db";
import versionCompare from 'version-compare'
import { Extension, ExtensionChangelog } from "../../../shared/shop";
import promiseAllProperties from '../helper/promise'
 
interface SQLShop {
    id: string;
    url: string;
    shopware_version: string;
    client_id: string;
    client_secret: string;
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

            return new Response('OK');
        } else if (url.pathname === '/delete') {
            await this.state.storage.deleteAll();

            return new Response('OK');
        }

        return new Response('', {status: 404});
    }

    async alarm(): Promise<void> {
        const con = getConnection(this.env);

        const id = await this.state.storage.get('id')

        // ID is missing, so we can't do anything
        if (id === undefined) {
            await this.state.storage.deleteAll();
            return;
        }

        const fetchShopSQL = 'SELECT id, url, shopware_version, client_id, client_secret FROM shop WHERE id = ?';

        const shops = await con.execute(fetchShopSQL, [id]);

        // Shop is missing, so we can't do anything
        if (shops.rows.length === 0) {
            console.log(`cannot find shop: ${id}. Destroy self`)
            await this.state.storage.deleteAll();
            return;
        }


        await this.updateShop(shops.rows[0] as SQLShop, con);
        console.log(`Updated shop ${id}`)

        this.state.storage.setAlarm(Date.now() + 60 * MINUTES);
    }

    async updateShop(shop: SQLShop, con: Connection) {
    
        await con.execute('UPDATE shop SET last_scraped_at = NOW() WHERE id = ?', [shop.id]);
    
        const client = new HttpClient(new Shop('', shop.url, '', shop.client_id, shop.client_secret));
    
        try {
            await client.getToken();
        } catch (e) {
            await con.execute('UPDATE shop SET status = ?, last_scraped_error = ? WHERE id = ?', [
                'red',
                'The API authentication failed. Please check your credentials.',
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
                cacheInfo: client.get('/_action/cache_info'),
                home: fetch(shop.url)
            })
        } catch (e: any) {
            let error = e;
    
            if (e.response) {
                error = `Request failed with status code: ${e.response.statusCode}: Body: ${JSON.stringify(e.response.body)}`;
            }
    
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
                active: plugin.active,
                version: plugin.version,
                latestVersion: plugin.upgradeVersion,
                ratingAverage: null,
                storeLink: null,
                changelog: null,
                installed: plugin.installedAt !== null,
            } as Extension)
        }
    
        for (const app of responses.app.body.data) {
            extensions.push({
                name: app.name,
                active: app.active,
                version: app.version,
                latestVersion: null,
                ratingAverage: null,
                storeLink: null,
                changelog: null,
                installed: true, // When it's in DB its always installed
            } as Extension)
        }
    
        const scheduledTasks = responses.scheduledTask.body.data.map((task: any) => {
            return {
                name: task.name,
                status: task.status,
                lastExecutionTime: task.lastExecutionTime,
                nextExecutionTime: task.nextExecutionTime,
            };
        });
    
        if (extensions.length) {
            const url = new URL('https://api.shopware.com/pluginStore/pluginsByName')
            url.searchParams.set('locale', 'en-GB');
            url.searchParams.set('shopwareVersion', shop.shopware_version);
    
            for (let extension of extensions) {
                url.searchParams.append('technicalNames[]', extension.name);
            }
    
            const storeResp = await fetch(url.toString(), {
                cf: {
                    cacheTtl: 60 * 60 * 6 // 6 hours
                }
            })
    
            if (storeResp.ok) {
                const storePlugins = await storeResp.json() as any[];
    
                for (let extension of extensions) {
                    const storePlugin = storePlugins.find((plugin: any) => plugin.name === extension.name);
    
                    if (storePlugin) {
                        extension.latestVersion = storePlugin.version;
                        extension.ratingAverage = storePlugin.ratingAverage;
                        extension.storeLink = storePlugin.storeLink;
    
                        if (storePlugin.latestVersion != extension.version) {
                            const changelogs: ExtensionChangelog[] = [];
    
                            for (let changelog of storePlugin.changelog) {
                                if (versionCompare(changelog.version, extension.version) > 0) {
                                    changelogs.push({
                                        version: changelog.version,
                                        text: changelog.text,
                                        creationDate: changelog.creationDate.date,
                                    } as ExtensionChangelog)
                                }
                            }
    
                            extension.changelog = changelogs;
                        }
                    }
                }
            }
        }
    
        let favicon: string|null = null
    
        if (responses.home.ok) {
            const body = await responses.home.text();
            const match = /rel="shortcut icon"\s*href="(?<icon>.*)">/gm.exec(body);
    
            if (match && match.groups && match.groups.icon !== undefined) {
                favicon = match.groups.icon;
            }
        }
    
        await con.execute('UPDATE shop SET status = ?, shopware_version = ?, favicon = ?, last_scraped_error = null WHERE id = ?', [
            'green',
            responses.config.body.version,
            favicon,
            shop.id,
        ]);
    
        await con.execute('REPLACE INTO shop_scrape_info(shop_id, extensions, scheduled_task, queue_info, cache_info, created_at) VALUES(?, ?, ?, ?, ?, NOW())', [
            shop.id,
            JSON.stringify(extensions),
            JSON.stringify(scheduledTasks),
            JSON.stringify(responses.queue.body),
            JSON.stringify(responses.cacheInfo.body),
            favicon
        ]);
    }
}