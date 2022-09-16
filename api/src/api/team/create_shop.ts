import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { encrypt } from "../../crypto";
import { getConnection } from "../../db";
import Shops from "../../repository/shops";
import { ErrorResponse, JsonResponse } from "../common/response";

export async function createShop(req: Request, env: Env): Promise<Response> {
    const json = await req.json() as { shop_url?: string, client_id?: string, client_secret?: string, name?: string };

    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    if (typeof json.shop_url !== 'string') {
        return new ErrorResponse('Invalid shop_url.', 400);
    }

    if (typeof json.client_id !== 'string') {
        return new ErrorResponse('Invalid client_id.', 400);
    }

    if (typeof json.client_secret !== 'string') {
        return new ErrorResponse('Invalid client_secret.', 400);
    }

    if (typeof json.name === "undefined" || typeof json.name !== 'string') {
        return new ErrorResponse('Invalid name.', 400);
    }

    const client = new HttpClient(new Shop('', json.shop_url, '', json.client_id, json.client_secret))

    let resp;
    try {
        resp = await client.get('/_info/config');
    } catch (e) {
        return new ErrorResponse('Cannot reach shop', 400);
    }

    const con = getConnection(env);

    const clientSecret = await encrypt(env.APP_SECRET, json.client_secret);

    try {
        const id = await Shops.createShop(con, {
            team_id: teamId,
            name: json.name || json.shop_url,
            client_id: json.client_id,
            client_secret: clientSecret,
            shop_url: json.shop_url,
            version: resp.body.version,
        });

        const scrapeObject = env.SHOPS_SCRAPE.get(env.SHOPS_SCRAPE.idFromName(id));

        await scrapeObject.fetch(`http://localhost/now?id=${id.toString()}`);

        const pagespeedObject = env.PAGESPEED_SCRAPE.get(env.PAGESPEED_SCRAPE.idFromName(id));

        await pagespeedObject.fetch(`http://localhost/now?id=${id.toString()}`);

        return new JsonResponse({ id });
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
        return new ErrorResponse(e.message || 'Unknown error')
    }
}
