import { getConnection, schema } from "../../db";
import { ErrorResponse, JsonResponse } from "../common/response";
import { eq, desc } from 'drizzle-orm';

export async function getShop(req: Request, env: Env): Promise<Response> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env)

    const shop = await con.select({
        id: schema.shop.id,
        name: schema.shop.name,
        url: schema.shop.url,
        status: schema.shop.status,
        created_at: schema.shop.created_at,
        shopware_version: schema.shop.shopware_version,
        last_scraped_at: schema.shop.last_scraped_at,
        last_scraped_error: schema.shop.last_scraped_error,
        last_updated: schema.shop.last_updated,
        ignores: schema.shop.ignores,
        shop_image: schema.shop.shop_image,
        extensions: schema.shopScrapeInfo.extensions,
        scheduled_task: schema.shopScrapeInfo.scheduled_task,
        queue_info: schema.shopScrapeInfo.queue_info,
        cache_info: schema.shopScrapeInfo.cache_info,
        checks: schema.shopScrapeInfo.checks,
        team_id: schema.shop.team_id,
        team_name: schema.team.name,
    })
        .from(schema.shop)
        .innerJoin(schema.team, eq(schema.team.id, schema.shop.team_id))
        .leftJoin(schema.shopScrapeInfo, eq(schema.shopScrapeInfo.shop, schema.shop.id))
        .where(eq(schema.shop.id, parseInt(shopId)))
        .get();

    if (shop === undefined) {
        return new ErrorResponse('Not Found.', 400);
    }

    const pageSpeed = await con.query.shopPageSpeed.findMany({
        where: eq(schema.shopPageSpeed.shop_id, parseInt(shopId)),
        orderBy: [desc(schema.shopPageSpeed.created_at)]
    })

    const shopChangelog = await con.query.shopChangelog.findMany({
        where: eq(schema.shopChangelog.shop_id, parseInt(shopId)),
        orderBy: [desc(schema.shopChangelog.date)]
    });

    for (const row of shopChangelog) {
        row.extensions = JSON.parse(row.extensions);
    }

    shop.extensions = JSON.parse(shop.extensions || '{}');
    shop.scheduled_task = JSON.parse(shop.scheduled_task || '{}');
    shop.queue_info = JSON.parse(shop.queue_info || '{}');
    shop.cache_info = JSON.parse(shop.cache_info || '{}');
    shop.checks = JSON.parse(shop.checks || '{}');
    shop.ignores = JSON.parse(shop.ignores || '[]');
    // @ts-ignore
    shop.pagespeed = pageSpeed;
    // @ts-ignore
    shop.changelog = shopChangelog;

    return new JsonResponse(shop);
}
