import { getConnection, getKv } from "../../db";
import bcryptjs from "bcryptjs";
import { randomString } from "../../util";
import { NoContentResponse } from "../common/response";

export async function login(req: Request): Promise<Response> {
    const json = await req.json();

    if (json.email === undefined || json.password === undefined) {
        return new Response('Missing email or password', {
            status: 400
        });
    }

    const result = await getConnection().execute("SELECT id, password FROM user WHERE email = ?", [json.email]);

    if (!result.rows.length) {
        return new Response('User not found', {
            status: 404
        });
    }

    const user = result.rows[0];

    const passwordIsValid = bcryptjs.compareSync(json.password, user.password as string);

    if (!passwordIsValid) {
        return new Response('Invalid password', {
            status: 400
        });
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

    return new NoContentResponse({
        'token': authToken
    });
}