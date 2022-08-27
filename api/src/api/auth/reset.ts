import { getConnection, getKv } from "../../db";
import { sendMailResetPassword } from "../../mail/mail";
import { randomString } from "../../util";
import { ErrorResponse, JsonResponse, NoContentResponse } from "../common/response";
import bcryptjs from "bcryptjs";

export async function resetPasswordMail(req: Request) {
    const {email} = await req.json();

    const token = randomString(32);

    const result = await getConnection().execute('SELECT id FROM user WHERE email = ?', [email]);

    if (result.rows.length === 0) {
        return new NoContentResponse()
    }

    await getKv().put(`reset_${token}`, result.rows[0].id.toString());

    await sendMailResetPassword(email, token);

    return new NoContentResponse()
}

export async function resetAvailable(req: Request) {
    const {token} = await req.params;

    const result = await getKv().get(`reset_${token}`);

    if (result === null) {
        return new JsonResponse({}, 404);
    }

    return new JsonResponse({}, 200);
}

export async function confirmResetPassword(req: Request): Promise<Response> {
    const {password} = await req.json();
    const {token} = req.params;

    const id = await getKv().get(`reset_${token}`);

    if (!id) {
        return new ErrorResponse('Invalid token', 400);
    }

    const salt = await bcryptjs.genSalt(10);
    const newPassword = await bcryptjs.hash(password, salt);

    await getConnection().execute('UPDATE user SET password = ? WHERE id = ?', [newPassword, id]);
    await getKv().delete(`reset_${token}`);

    return new NoContentResponse()
}