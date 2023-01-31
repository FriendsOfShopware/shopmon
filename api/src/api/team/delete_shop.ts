import { getConnection } from "../../db";
import Shops from "../../repository/shops";
import { ErrorResponse, NoContentResponse } from "../common/response";

export async function deleteShop(req: Request, env: Env): Promise<Response> {
    const { teamId, shopId } = req.params as { teamId?: string, shopId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env)

    const result = await con.execute('SELECT shop.id, shop.url, shop.created_at, shop.shopware_version, shop.last_scraped_at, shop.shopware_version, shop_scrape_info.extensions, shop_scrape_info.scheduled_task FROM shop LEFT JOIN shop_scrape_info ON(shop_scrape_info.shop_id = shop.id) WHERE shop.id = ? AND shop.team_id = ?', [
        shopId,
        teamId
    ]);

    if (result.rows.length === 0) {
        return new Response('Not Found.', { status: 404 });
    }

    const convertedShopId = parseInt(shopId);

    if (isNaN(convertedShopId)) {
        return new ErrorResponse('Invalid shopId', 400);
    }

    await Shops.deleteShop(con, convertedShopId);

    const scrapeObject = env.SHOPS_SCRAPE.get(env.SHOPS_SCRAPE.idFromName(shopId))

    await scrapeObject.fetch(`http://localhost/delete?id=${shopId.toString()}`);

    const pageSpeedObject = env.PAGESPEED_SCRAPE.get(env.PAGESPEED_SCRAPE.idFromName(shopId))

    await pageSpeedObject.fetch(`http://localhost/delete?id=${shopId.toString()}`);

    return new NoContentResponse();
}