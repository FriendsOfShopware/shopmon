import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../db";

const fetchShopSQL = 'SELECT id, url, client_id, client_secret FROM shop WHERE last_scraped_at IS NULL OR last_scraped_at < DATE_SUB(NOW(), INTERVAL 1 HOUR) ORDER BY id ASC LIMIT 1';

export async function onSchedule() {
    const con = getConnection();

    const shops = await con.execute(fetchShopSQL);

    if (shops.rows.length === 0) {
        return;
    }

    const shop = shops.rows[0];

    await con.execute('UPDATE shop SET last_scraped_at = NOW() WHERE id = ?', [shop.id]);

    const client = new HttpClient(new Shop('', shop.url, '', shop.client_id, shop.client_secret));

    
    const responses = await Promise.allSettled([
        client.get('/_info/config'),
        client.get('/_action/extension/installed'),
        client.post('/search/scheduled-task')
    ]) as any

    await con.execute('UPDATE shop SET shopware_version = ? WHERE id = ?', [
        responses[0].value.body.version,
        shop.id,
    ]);

    const extensions = responses[1].value.body.map((extension: any) => {
        return {
            name: extension.name,
            active: extension.active,
            version: extension.version,
            latestVersion: extension.latestVersion,
            installed: extension.installedAt !== null,
        };
    });

    const scheduledTasks = responses[2].value.body.data.map((task: any) => {
        return {
            name: task.name,
            status: task.status,
            latestVersion: task.latestVersion,
            lastExecutionTime: task.lastExecutionTime,
            nextExecutionTime: task.nextExecutionTime,
        };
    });

    await con.execute('REPLACE INTO shop_scrape_info(shop_id, extensions, scheduled_task, created_at) VALUES(?, ?, ?, NOW())', [
        shop.id,
        JSON.stringify(extensions),
        JSON.stringify(scheduledTasks),
    ]);
}
