import Teams from "../../repository/teams";
import { NoContentResponse } from "../common/response";

export async function deleteTeam(req: Request): Promise<Response> {
    await Teams.deleteTeam(req.team.id);

    return new NoContentResponse();
}