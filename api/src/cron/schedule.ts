import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../db";
import promiseAllProperties from 'promise-all-properties';
import versionCompare from 'version-compare'
import type {Extension, ExtensionChangelog} from "../../../shared/shop";

const faviconRegex = /rel="shortcut icon"\s*href="(?<icon>.*)">/gm;
const fetchShopSQL = 'SELECT id, url, shopware_version , client_id, client_secret FROM shop WHERE last_scraped_at IS NULL OR last_scraped_at < DATE_SUB(NOW(), INTERVAL 1 HOUR) ORDER BY id ASC LIMIT 1';

export async function onSchedule() {
    const con = getConnection();

    const shops = await con.execute(fetchShopSQL);

    if (shops.rows.length === 0) {
        return;
    }

    const shop = shops.rows[0];

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
        const match = faviconRegex.exec(body);

        if (match && match.groups && match.groups.icon !== undefined) {
            favicon = match.groups.icon;
        }
    }

    await con.execute('UPDATE shop SET status = ?, shopware_version = ?, favicon = ? WHERE id = ?', [
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