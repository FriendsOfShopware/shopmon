import { getConnection } from "../db";

const fetchShopSQL = 'SELECT id FROM shop';

export async function onSchedule(env: Env) {
    const con = getConnection(env);

    const shops = await con.execute(fetchShopSQL);

    for (let shop of shops.rows) {
        const scrapeObject = env.SHOPS_SCRAPE.get(env.SHOPS_SCRAPE.idFromName(shop.id.toString()));

        await scrapeObject.fetch(`http://localhost/cron?id=${shop.id.toString()}`);

        console.log(`Scheduled shop ${shop.id}`);
    }
}