import { getConnection } from "../../db";
import { ErrorResponse, NoContentResponse } from "../common/response";

export async function confirmMail(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    // todo perform input validation
    const { token } = req.params as { token: string };

    const result = await con.execute('SELECT id FROM user WHERE verify_code = ?', [token])

    if (result.rows.length === 0) {
        return new ErrorResponse('Invalid confirm token', 400);
    }

    const userId = result.rows[0].id;

    await con.execute('UPDATE user SET verify_code = NULL WHERE id = ?', [userId]);

    return new NoContentResponse();
}