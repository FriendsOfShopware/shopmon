import {ErrorResponse, NoContentResponse} from "../common/response";
import {Shop} from "shopware-app-server-sdk/shop";
import {decrypt} from "../../crypto";
import {HttpClient} from "shopware-app-server-sdk/component/http-client";
import {getConnection} from "../../db";

export async function clearShopCache(req: Request, env: Env): Promise<Response> {
    const {shopId} = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env);
    const shopData = await con.execute('SELECT url,client_id,client_secret FROM shop WHERE id = ?', [shopId]);

    if (!shopData.rows.length) {
        return new ErrorResponse("No shop with this ID found", 404);
    }
    const clientSecret = await decrypt(env.APP_SECRET, shopData.rows[0].client_secret);
    const client = new HttpClient(new Shop('', shopData.rows[0].url, '', shopData.rows[0].client_id, clientSecret));

    try {
        await client.delete('/_action/cache');
    } catch (e) {
        return new ErrorResponse(e.response.body.errors[0].detail, e.response.statusCode);
    }

    return new NoContentResponse();
}