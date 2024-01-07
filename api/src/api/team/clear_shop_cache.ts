import { ErrorResponse, NoContentResponse } from "../common/response";
import { SimpleShop, HttpClient } from "@friendsofshopware/app-server-sdk"
import { decrypt } from "../../crypto";
import { getConnection, schema } from "../../db";
import { eq } from 'drizzle-orm';

export async function clearShopCache(req: Request, env: Env): Promise<Response> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env);

    const shopData = await con.query.shop.findFirst({
        columns: {
            url: true,
            client_id: true,
            client_secret: true,
        },
        where: eq(schema.shop.id, parseInt(shopId))
    });

    if (shopData === undefined) {
        return new ErrorResponse("No shop with this ID found", 404);
    }
    const clientSecret = await decrypt(env.APP_SECRET, shopData.client_secret);
    const shop = new SimpleShop('', shopData.url, '');
    shop.setShopCredentials(shopData.client_id, clientSecret);
    const client = new HttpClient(shop);

    try {
        await client.delete('/_action/cache');
    } catch (e: any) {
        return new ErrorResponse(e.response.body.errors[0].detail, e.response.statusCode);
    }

    return new NoContentResponse();
}
