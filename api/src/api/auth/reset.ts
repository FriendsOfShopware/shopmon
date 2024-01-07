import { getConnection, user as userTable } from "../../db";
import { sendMailResetPassword } from "../../mail/mail";
import { randomString } from "../../util";
import { ErrorResponse, JsonResponse, NoContentResponse } from "../common/response";
import { eq } from 'drizzle-orm';
import bcryptjs from "bcryptjs";
import Users from "../../repository/users";

export async function resetPasswordMail(req: Request, env: Env) {
    let { email } = await req.json() as { email: string };

    const token = randomString(32);

    const con = getConnection(env);

    email = email.toLowerCase();

    const result = await con.query.user.findFirst({
        columns: {
            id: true
        },
        where: eq(userTable.email, email)
    })

    if (result === undefined) {
        return new NoContentResponse()
    }

    await env.kvStorage.put(`reset_${token}`, result.id.toString(), {
        expirationTtl: 60 * 60, // 1 hour
    });

    await sendMailResetPassword(env, email, token);

    return new NoContentResponse()
}

export async function resetAvailable(req: Request, env: Env) {
    const { token } = await req.params as { token: string };

    const result = await env.kvStorage.get(`reset_${token}`);

    if (result === null) {
        return new JsonResponse({}, 404);
    }

    return new JsonResponse({}, 200);
}

export async function confirmResetPassword(req: Request, env: Env): Promise<Response> {
    const { password } = await req.json() as { password?: string };
    const { token } = req.params as { token?: string };

    if (typeof password !== "string") {
        return new ErrorResponse('Missing password', 400);
    }

    const con = getConnection(env)
    const id = await env.kvStorage.get(`reset_${token}`);

    if (!id) {
        return new ErrorResponse('Invalid token', 400);
    }

    const salt = await bcryptjs.genSalt(10);
    const newPassword = await bcryptjs.hash(password, salt);

    await con
        .update(userTable)
        .set({ password: newPassword })
        .where(eq(userTable.id, parseInt(id)));

    await env.kvStorage.delete(`reset_${token}`);

    await Users.revokeUserSessions(env.kvStorage, id);

    return new NoContentResponse()
}
