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

    // Authentificate only once
    await client.getToken();
    
    const responses = await Promise.allSettled([
        client.get('/_info/config'),
        client.get('/_action/extension/installed'),
        client.post('/search/scheduled-task'),
        client.get('/_info/queue.json'),
        client.get('/_action/cache_info')
    ]) as any

    for (let [i, response] of responses.entries()) {
        if (response.status === 'rejected') {
            let error = response.reason;

            if (response.reason.response) {
                error = `Request#${i} failed with status code: ${response.reason.response.statusCode}: Body: ${JSON.stringify(response.reason.response.body)}`;
            }

            await con.execute('UPDATE shop SET status = ?, last_scrapted_error = ? WHERE id = ?', [
                'red',
                error,
                shop.id,
            ]);
            return;
        }
    }

    await con.execute('UPDATE shop SET status = ?, shopware_version = ? WHERE id = ?', [
        'green',
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
            lastExecutionTime: task.lastExecutionTime,
            nextExecutionTime: task.nextExecutionTime,
        };
    });

    await con.execute('REPLACE INTO shop_scrape_info(shop_id, extensions, scheduled_task, queue_info, cache_info, created_at) VALUES(?, ?, ?, ?, ?, NOW())', [
        shop.id,
        JSON.stringify(extensions),
        JSON.stringify(scheduledTasks),
        JSON.stringify(responses[3].value.body),
        JSON.stringify(responses[4].value.body)
    ]);
}