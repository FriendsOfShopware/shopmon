import { getConnection, schema } from "../../db";
import { ErrorResponse } from "../common/response";
import { and, eq } from 'drizzle-orm';

export async function validateTeam(req: Request, env: Env): Promise<Response | void> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    const con = getConnection(env)

    const result = await con.select({
        ownerId: schema.team.owner_id,
    })
        .from(schema.team)
        .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.team.id))
        .where(and(eq(schema.userToTeam.user_id, req.userId), eq(schema.team.id, parseInt(teamId))))
        .get()

    if (result === undefined) {
        return new Response('Not Found.', { status: 404 });
    }

    req.team = {
        id: teamId,
        ownerId: result.ownerId,
    };
}

export async function validateTeamOwner(req: Request): Promise<Response | void> {
    if (req.team.ownerId !== req.userId) {
        return new Response('Forbidden.', { status: 403 });
    }
}
