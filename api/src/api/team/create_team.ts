import { getConnection } from "../../db";
import Teams from "../../repository/teams";
import { ErrorResponse, JsonResponse } from "../common/response";
import { init } from "@mmyoji/object-validator";

type CreateTeamRequest = {
    name: string;
    ownerId: number;
}

const validate = init<CreateTeamRequest>({
    name: {
        type: "string",
        required: true
    },
    ownerId: {
        type: "number",
        required: true
    },
});

export async function createTeam(req: Request, env: Env): Promise<Response> {
    const json = await req.json() as CreateTeamRequest;

    const errors = validate(json);
    if (errors.length) {
        return new JsonResponse(errors, 400);
    }

    if (json.ownerId as string !== req.userId as string) {
        return new Response('Forbidden.', { status: 403 });
    }

    const con = getConnection(env);
    try {
        const teamId = await Teams.createTeam(con, json.name, json.ownerId as string);
        return new JsonResponse({ teamId });
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
        return new ErrorResponse(e.message || 'Unknown error')
    }
}