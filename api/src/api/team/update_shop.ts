import { init } from "@mmyoji/object-validator";
import { HttpClient } from "shopware-app-server-sdk/component/http-client";
import { Shop } from "shopware-app-server-sdk/shop";
import { getConnection } from "../../db";
import { ErrorResponse, JsonResponse, NoContentResponse } from "../common/response";

type UpdateShopRequest = {
    name: string;
    shop_url: string;
    teamId: number;
    client_id: string;
    client_secret: string;
}

const validate = init<UpdateShopRequest>({
    name: {
        type: "string",
        required: true
    },
    teamId: {
        type: "number",
        required: true
    },
    shop_url: {
        type: "string",
    },
    client_id: {
        type: "string",
    },
    client_secret: {
        type: "string"
    }
});

export async function updateShop(req: Request, env: Env): Promise<Response> {
    const json = await req.json() as UpdateShopRequest;

    const errors = validate(json);
    if (errors.length) {
        return new JsonResponse(errors, 400);
    }

    const { shopId, teamId } = req.params;

    const con = getConnection(env);

    await con.execute('UPDATE shop SET name = ? WHERE id = ?', [json.name, shopId]);

    if (json.teamId != teamId) {
        const team = await con.execute('SELECT 1 FROM user_to_team WHERE user_id = ? AND team_id = ?', [req.userId, json.teamId]);

        if (!team.rows.length) {
            return new ErrorResponse("You are not a member of this team", 403);
        }

        await con.execute('UPDATE shop SET team_id = ? WHERE id = ?', [json.teamId, shopId]);
    }

    if (json.shop_url && json.client_id && json.client_secret) {
        // Try out the new credentials
        const client = new HttpClient(new Shop('', json.shop_url, '', json.client_id, json.client_secret))

        try {
            await client.get('/_info/config');
        } catch (e) {
            return new ErrorResponse('Cannot reach shop', 400);
        }

        await con.execute('UPDATE shop SET url = ?, client_id = ?, client_secret = ? WHERE id = ?', [json.shop_url, json.client_id, json.client_secret, shopId]);
    }

    return new NoContentResponse();
}