import Teams from "../../repository/teams";
import { JsonResponse, NoContentResponse } from "../common/response";

export async function listMembers(req: Request): Promise<Response> {
    const { teamId } = req.params;

    return new JsonResponse(await Teams.listMembers(teamId));
}

export async function addMember(req: Request): Promise<Response> {
    const { teamId } = req.params;

    const json = await req.json();

    if (json.email === undefined) {
        return new JsonResponse({
            message: "Missing email"
        }, 400);
    }

    await Teams.addMember(teamId, json.email);

    return new NoContentResponse();
}

export async function removeMember(req: Request): Promise<Response> {
    const { teamId, userId } = req.params;

    if (userId == req.userId) {
        return new JsonResponse({
            message: "You cannot remove yourself"
        }, 400);
    }

    await Teams.removeMember(teamId, userId);

    return new NoContentResponse();
}