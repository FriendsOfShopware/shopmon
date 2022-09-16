import {getConnection} from "../../db";
import Teams from "../../repository/teams";
import {ErrorResponse, JsonResponse, NoContentResponse} from "../common/response";
import {init} from "@mmyoji/object-validator";

type UpdateTeamRequest = {
    name: string;
}

const validate = init<UpdateTeamRequest>({
    name: {
        type: "string",
        required: true
    }
});

export async function updateTeam(req: Request, env: Env): Promise<Response> {
    const json = await req.json() as UpdateTeamRequest;

    const errors = validate(json);
    if (errors.length) {
        return new JsonResponse(errors, 400);
    }

    const { teamId } = req.params as { teamId?: string };

    if (typeof teamId !== "string") {
        return new ErrorResponse('Missing teamId', 400);
    }

    const con = getConnection(env);
    await Teams.updateTeam(con, teamId, json.name);

    return new NoContentResponse();
}