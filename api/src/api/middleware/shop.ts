import { getConnection, schema } from "../../db";
import { ErrorResponse } from "../common/response";
import { and, eq } from 'drizzle-orm';

export async function validateShop(req: Request, env: Env): Promise<Response | void> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env)

    const result = await con.query.shop.findFirst({
        columns: {
            id: true,
        },
        where: and(eq(schema.shop.id, parseInt(shopId)), eq(schema.shop.team_id, parseInt(req.team.id)))
    });

    if (result === undefined) {
        return new Response('Not Found.', { status: 404 });
    }
}
