import { getConnection } from "../../db";
import { ErrorResponse } from "../common/response";

export async function validateTeam(req: Request, env: Env): Promise<Response|void> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    const con = getConnection(env)

    const res = await con.execute('SELECT team.owner_id as ownerId FROM user_to_team INNER JOIN team ON(team.id = user_to_team.team_id) WHERE user_id = ? AND team_id = ?', [req.userId, teamId]);

    if (res.rows.length === 0) {
        return new Response('Not Found.', { status: 404 });
    }

    req.team = {
        id: teamId,
        ownerId: res.rows[0].ownerId
    };
}

export async function validateTeamOwner(req: Request): Promise<Response|void> {
    if (req.team.ownerId !== req.userId) {
        return new Response('Forbidden.', { status: 403 });
    }
}