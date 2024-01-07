import { getConnection, schema } from "../../db";
import { eq } from 'drizzle-orm';
import { JsonResponse } from "../common/response";

export async function listShops(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con.query.shop.findMany({
        columns: {
            id: true,
            name: true,
            url: true,
            favicon: true,
            created_at: true,
            last_scraped_at: true,
            status: true,
            last_scraped_error: true,
            shopware_version: true
        },
        where: eq(schema.shop.team_id, parseInt(req.team.id))
    })

    return new JsonResponse(result)
}
