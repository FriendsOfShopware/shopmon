import { getConnection } from "../../db";
import { ErrorResponse, JsonResponse } from "../common/response";

export async function getShop(req: Request, env: Env): Promise<Response> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env)

    const result = await con.execute(`
    SELECT 
        shop.id, 
        shop.name, 
        shop.url, 
        shop.status,
        shop.created_at, 
        shop.shopware_version, 
        shop.last_scraped_at, 
        shop.last_scraped_error,
        shop.shopware_version,
        shop.ignores,
        shop_scrape_info.extensions, 
        shop_scrape_info.scheduled_task,
        shop_scrape_info.queue_info,
        shop_scrape_info.cache_info,
        shop_scrape_info.checks,
        shop.team_id,
        team.name as team_name
    FROM shop
        INNER JOIN team ON(team.id = shop.team_id)
        LEFT JOIN shop_scrape_info ON(shop_scrape_info.shop_id = shop.id) 
    WHERE 
        shop.id = ?
    `, [
        shopId
    ]);

    if (result.rows.length === 0) {
        return new ErrorResponse('Not Found.', 400);
    }

    const shop = result.rows[0];

    const pagespeed = await con.execute('SELECT * FROM shop_pagespeed WHERE shop_id = ? ORDER BY created_at DESC', [shopId]);

    shop.extensions = JSON.parse(shop.extensions);
    shop.scheduled_task = JSON.parse(shop.scheduled_task);
    shop.queue_info = JSON.parse(shop.queue_info);
    shop.cache_info = JSON.parse(shop.cache_info);
    shop.checks = JSON.parse(shop.checks);
    shop.ignores = JSON.parse(shop.ignores);
    shop.pagespeed = pagespeed.rows;

    return new JsonResponse(shop);
}