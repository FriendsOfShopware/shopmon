import { getConnection } from "../../db";
import Teams from "../../repository/teams";
import { ErrorResponse, NoContentResponse } from "../common/response";

export async function deleteTeam(req: Request, env: Env): Promise<Response> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    const con = getConnection(env);
    await Teams.deleteTeam(con, parseInt(teamId));

    return new NoContentResponse();
}
