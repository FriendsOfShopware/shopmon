import { getConnection } from "../../db";

export async function listShops(req: Request): Promise<Response> {
    const { teamId } = req.params;

    const res = await getConnection().execute('SELECT id, shop_url, created_at FROM shops WHERE team_id = ?', [teamId]);

    return new Response(JSON.stringify(res.rows), { 
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}