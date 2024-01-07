import { getConnection, user as userTable } from "../../db";
import { ErrorResponse, NoContentResponse } from "../common/response";
import { eq } from 'drizzle-orm';

export async function confirmMail(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const { token } = req.params as { token: string };

    const result = await con.query.user.findFirst({
        columns: {
            id: true
        },
        where: eq(userTable.verify_code, token)
    })

    if (result === undefined) {
        return new ErrorResponse('Invalid confirm token', 400);
    }

    await con
        .update(userTable)
        .set({ verify_code: null })
        .where(eq(userTable.id, result.id));

    return new NoContentResponse();
}
