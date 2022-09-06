import { getConnection } from "../../db";
import { ErrorResponse } from "../common/response";

export async function listShops(req: Request, env: Env): Promise<Response> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }
    
    const con = getConnection(env);

    const res = await con.execute('SELECT id, name, url, favicon, created_at, last_scraped_at, status, last_scraped_error, shopware_version FROM shop WHERE team_id = ?', [teamId]);

    return new Response(JSON.stringify(res.rows), { 
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}