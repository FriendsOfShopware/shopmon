import Teams from "../../repository/teams";
import { JsonResponse, NoContentResponse } from "../common/response";

export async function listMembers(req: Request): Promise<Response> {
    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new JsonResponse('Missing teamId', 400);
    }

    // todo fix bug missing con
    return new JsonResponse(await Teams.listMembers(con, teamId));
}

export async function addMember(req: Request): Promise<Response> {
    const { teamId } = req.params as { teamId?: string };
    const json = await req.json() as { email?: string };

    if (typeof teamId !== "string") {
        return new JsonResponse('Missing teamId', 400);
    }

    if (typeof json.email !== "string") {
        return new JsonResponse('Missing email', 400);
    }

    if (json.email === undefined) {
        return new JsonResponse({
            message: "Missing email"
        }, 400);
    }

    // todo fix bug missing con
    await Teams.addMember(con, teamId, json.email);

    return new NoContentResponse();
}

export async function removeMember(req: Request): Promise<Response> {
    const { teamId, userId } = req.params as { teamId?: string, userId?: string };

    if (typeof teamId !== "string") {
        return new JsonResponse('Missing teamId', 400);
    }

    if (typeof userId !== "string") {
        return new JsonResponse('Missing userId', 400);
    }

    if (userId == req.userId) {
        return new JsonResponse({
            message: "You cannot remove yourself"
        }, 400);
    }

    // todo fix bug missing con
    await Teams.removeMember(con, teamId, userId);

    return new NoContentResponse();
}