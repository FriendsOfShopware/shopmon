import { getConnection } from "../../db";
import { ErrorResponse } from "../common/response";

export async function validateShop(req: Request, env: Env): Promise<Response|void> {
    const { shopId } = req.params as { shopId?: string };

    if (typeof shopId !== "string") {
        return new ErrorResponse('Missing shopId', 400);
    }

    const con = getConnection(env)

    const res = await con.execute('SELECT 1 FROM shop WHERE id = ? AND team_id = ?', [shopId, req.team.id]);

    if (res.rows.length === 0) {
        return new Response('Not Found.', { status: 404 });
    }
}