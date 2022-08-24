import { getConnection } from "../../db";

export async function validateTeam(req: Request): Promise<Response|void> {
    const { teamId } = req.params;

    const res = await getConnection().execute('SELECT 1 FROM user_to_team WHERE user_id = ? AND team_id = ?', [req.userId, teamId]);


    if (res.rows.length === 0) {
        return new Response('Not Found.', { status: 404 });
    }
}