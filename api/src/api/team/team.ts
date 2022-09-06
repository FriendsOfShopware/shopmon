import Teams from "../../repository/teams";
import { ErrorResponse, NoContentResponse } from "../common/response";

export async function deleteTeam(req: Request): Promise<Response> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    // todo fix bug missing con
    await Teams.deleteTeam(con, teamId);

    return new NoContentResponse();
}