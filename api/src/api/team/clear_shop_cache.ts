import { ErrorResponse, NoContentResponse } from "../common/response";

export async function clearShopCache(req: Request, env: Env): Promise<Response> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const obj = env.SHOPS_SCRAPE.get(env.SHOPS_SCRAPE.idFromName(shopId.toString()));

    await obj.fetch(`http://localhost/now?id=${shopId.toString()}&userId=${req.userId.toString()}`);

    return new NoContentResponse();
}