import { getConnection } from "../../db";
import { ErrorResponse, NoContentResponse } from "../common/response";

export async function confirmMail(req: Request): Promise<Response> {
    const { token } = req.params;

    const result = await getConnection().execute('SELECT id FROM user WHERE verify_code = ?', [token])

    if (result.rows.length === 0) {
        return new ErrorResponse('Invalid confirm token', 400);
    }

    const userId = result.rows[0].id;

    await getConnection().execute('UPDATE user SET verify_code = NULL WHERE id = ?', [userId]);

    return new NoContentResponse();
}