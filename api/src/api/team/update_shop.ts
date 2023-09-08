import { init } from "@mmyoji/object-validator";
import { SimpleShop, HttpClient } from "@friendsofshopware/app-server-sdk"
import { encrypt } from "../../crypto";
import { getConnection } from "../../db";
import { ErrorResponse, JsonResponse, NoContentResponse } from "../common/response";

type UpdateShopRequest = {
    name: string;
    url: string;
    ignores: string[];
    team_id: number;
    client_id: string;
    client_secret: string;
}

const validate = init<UpdateShopRequest>({
    name: {
        type: "string"
    },
    team_id: {
        type: "number"
    },
    url: {
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

    const { shopId } = req.params;

    const con = getConnection(env);

    if (json.name) {
        await con.execute('UPDATE shop SET name = ? WHERE id = ?', [json.name, shopId]);
    }

    if (json.ignores) {
        await con.execute('UPDATE shop SET ignores = ? WHERE id = ?', [JSON.stringify(json.ignores), shopId]);
    }

    if (json.team_id && json.team_id != parseInt(req.team.id)) {
        const team = await con.execute('SELECT 1 FROM user_to_team WHERE user_id = ? AND team_id = ?', [req.userId, req.team.id]);

        if (!team.rows.length) {
            return new ErrorResponse("You are not a member of this team", 403);
        }

        await con.execute('UPDATE shop SET team_id = ? WHERE id = ?', [json.team_id, shopId]);
    }

    if (json.url && json.client_id && json.client_secret) {
        // Try out the new credentials
        const shop = new SimpleShop('', json.url, '')
        shop.setShopCredentials(json.client_id, json.client_secret);

        const client = new HttpClient(shop)

        try {
            await client.get('/_info/config');
        } catch (e) {
            return new ErrorResponse('Cannot reach shop', 400);
        }

        const clientSecret = await encrypt(env.APP_SECRET, json.client_secret);

        await con.execute('UPDATE shop SET url = ?, client_id = ?, client_secret = ? WHERE id = ?', [json.url, json.client_id, clientSecret, shopId]);
    }

    return new NoContentResponse();
}