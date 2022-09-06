import { getConnection } from "../../db";
import bcryptjs from "bcryptjs";
import { randomString } from "../../util";
import { ErrorResponse, JsonResponse } from "../common/response";

export async function login(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const loginError = new ErrorResponse("Invalid email or password", 400);

    const json = await req.json() as { email?: string, password?: string };

    if (typeof json.email !== "string" || typeof json.password !== "string") {
        return new ErrorResponse('Missing email or password', 400);
    }

    const result = await con.execute("SELECT id, password FROM user WHERE email = ? AND verify_code IS NULL", [json.email]);

    if (!result.rows.length) {
        return loginError;
    }

    const user = result.rows[0];

    const passwordIsValid = await bcryptjs.compare(json.password, user.password as string);

    if (!passwordIsValid) {
        return loginError;
    }

    const authToken = `u-${user.id}-${randomString(20)}`;
    await env.kvStorage.put(
        authToken, 
        JSON.stringify({
            id: user.id
        }),
        {
            expirationTtl: 60 * 30, // 30 minutes
        }
    );

    return new JsonResponse({
        'token': authToken
    });
}