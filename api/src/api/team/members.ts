import { getConnection } from "../../db";
import Teams from "../../repository/teams";
import { JsonResponse, NoContentResponse } from "../common/response";

export async function listMembers(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    return new JsonResponse(await Teams.listMembers(con, parseInt(req.team.id)));
}

export async function addMember(req: Request, env: Env): Promise<Response> {
    const json = await req.json() as { email?: string };

    if (typeof json.email !== "string") {
        return new JsonResponse('Missing email', 400);
    }

    if (json.email === undefined) {
        return new JsonResponse({
            message: "Missing email"
        }, 400);
    }

    const con = getConnection(env);

    try {
        await Teams.addMember(con, req.team.id, json.email);
    } catch (e) {
        return new JsonResponse({
            message: e?.toString()
        }, 400);
    }

    return new NoContentResponse();
}

export async function removeMember(req: Request, env: Env): Promise<Response> {
    const { userId } = req.params as { userId?: string };

    if (typeof userId !== "string") {
        return new JsonResponse('Missing userId', 400);
    }

    if (userId.toString() === req.userId.toString()) {
        return new JsonResponse({
            message: "You cannot remove yourself"
        }, 400);
    }

    const con = getConnection(env);

    await Teams.removeMember(con, req.team.id, userId);

    return new NoContentResponse();
}
