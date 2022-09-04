import { NoContentResponse } from "../common/response";

export async function refreshShop(req: Request, env: Env): Promise<Response> {
    const { shopId } = req.params;

    const obj = env.SHOPS_SCRAPE.get(env.SHOPS_SCRAPE.idFromName(shopId.toString()));

    await obj.fetch(`http://localhost/now?id=${shopId.toString()}`);

    return new NoContentResponse();
}