import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../../db";
import Shops from "../../repository/shops";
import { ErrorResponse, JsonResponse } from "../common/response";

export async function createShop(req: Request): Promise<Response> {
    const json = await req.json();

    const { teamId } = req.params;

    if (typeof json.shop_url !== 'string') {
        return new Response('Invalid shop_url.', { status: 400 });
    }

    if (typeof json.client_id !== 'string') {
        return new Response('Invalid client_id.', { status: 400 });
    }

    if (typeof json.client_secret !== 'string') {
        return new Response('Invalid client_secret.', { status: 400 });
    }

    const client = new HttpClient(new Shop('', json.shop_url, '', json.client_id, json.client_secret))

    let resp;
    try {
        resp = await client.get('/_info/config');
    } catch (e) {
        return new Response('Cannot reach shop', { status: 400 });
    }

    try {
        const id = await Shops.createShop({
            team_id: teamId,
            name: json.name || json.shop_url,
            client_id: json.client_id,
            client_secret: json.client_secret,
            shop_url: json.shop_url,
            version: resp.body.version,
        });

        return new JsonResponse({ id });
    } catch (e: any) {
        return new ErrorResponse(e.message || 'Unknown error')
    }
}
