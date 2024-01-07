import { init } from "@mmyoji/object-validator";
import { SimpleShop, HttpClient } from "@friendsofshopware/app-server-sdk"
import { encrypt } from "../../crypto";
import { getConnection, schema } from "../../db";
import { eq, and } from 'drizzle-orm';
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
        await con.update(schema.shop).set({ name: json.name }).where(eq(schema.shop.id, parseInt(shopId))).execute();
    }

    if (json.ignores) {
        await con.update(schema.shop).set({ ignores: JSON.stringify(json.ignores) }).where(eq(schema.shop.id, parseInt(shopId))).execute();
    }

    if (json.team_id && json.team_id != parseInt(req.team.id)) {
        const team = await con.query.team.findFirst({
            columns: {
                id: true,
            },
            where: and(
                eq(schema.team.id, json.team_id),
                eq(schema.team.owner_id, req.userId)
            )
        });

        if (team === undefined) {
            return new ErrorResponse("You are not a member of this team", 403);
        }

        await con.update(schema.shop).set({ team_id: json.team_id }).where(eq(schema.shop.id, parseInt(shopId))).execute();
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

        await con.update(schema.shop).set({ url: json.url, client_id: json.client_id, client_secret: clientSecret }).where(eq(schema.shop.id, parseInt(shopId))).execute();
    }

    return new NoContentResponse();
}
