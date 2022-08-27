import { getConnection } from "../../db";
import Shops from "../../repository/shops";
import { NoContentResponse } from "../common/response";

export async function deleteShop(req: Request): Promise<Response> {
    const { teamId, shopId } = req.params;

    const result = await getConnection().execute('SELECT shop.id, shop.url, shop.created_at, shop.shopware_version, shop.last_scraped_at, shop.shopware_version, shop_scrape_info.extensions, shop_scrape_info.scheduled_task FROM shop LEFT JOIN shop_scrape_info ON(shop_scrape_info.shop_id = shop.id) WHERE shop.id = ? AND shop.team_id = ?', [
        shopId,
        teamId
    ]);

    if (result.rows.length === 0) {
        return new Response('Not Found.', { status: 404 });
    }

    await Shops.deleteShop(shopId);

    return new NoContentResponse();
}