import { getConnection } from "../../db";

export async function accountMe(req: Request): Promise<Response> {
    const result = await getConnection().execute('SELECT id, email, created_at FROM user WHERE id = ?', [req.userId]);

    const json = result.rows[0];

    const teamResult = await getConnection().execute('SELECT team.id, team.name, team.created_at, (team.owner_id = user_to_team.user_id) as is_owner  FROM team INNER JOIN user_to_team ON user_to_team.team_id = team.id WHERE user_to_team.user_id = ?', [req.userId]);

    json.teams = teamResult.rows;
    
    return new Response(JSON.stringify(json), {
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}