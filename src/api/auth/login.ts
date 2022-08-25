import { getConnection, getKv } from "../../db";
import bcryptjs from "bcryptjs";
import { randomString } from "../../util";
import { ErrorResponse, JsonResponse } from "../common/response";

const loginError = new ErrorResponse("Invalid email or password", 400);

export async function login(req: Request): Promise<Response> {
    const json = await req.json();

    if (json.email === undefined || json.password === undefined) {
        return new ErrorResponse('Missing email or password', 400);
    }

    const result = await getConnection().execute("SELECT id, password FROM user WHERE email = ? AND verify_code IS NULL", [json.email]);

    if (!result.rows.length) {
        return loginError;
    }

    const user = result.rows[0];

    const passwordIsValid = await bcryptjs.compare(json.password, user.password as string);

    if (!passwordIsValid) {
        return loginError;
    }

    const authToken = `u-${user.id}-${randomString(20)}`;
    await getKv().put(
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