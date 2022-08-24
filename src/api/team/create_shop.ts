import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../../db";

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
    
    await getConnection().execute('INSERT INTO shops (team_id, name, url, client_id, client_secret) VALUES (?, ?, ?, ?, ?)', [
        teamId, 
        json.name || json.shop_url,
        json.shop_url,
        json.client_id,
        json.client_secret
    ]);

    return new Response('', { status: 200 });
}
