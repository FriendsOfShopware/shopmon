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
