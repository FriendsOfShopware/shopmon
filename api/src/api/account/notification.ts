import { getConnection } from "../../db";
import { JsonResponse, NoContentResponse } from "../common/response";

export async function getNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con.execute(`SELECT * FROM user_notification WHERE user_id = ?`, [req.userId]);

    for (const row of result.rows) {
        row.link = JSON.parse(row.link);
    }

    return new JsonResponse(result.rows);
}

export async function deleteAllNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    await con.execute(`DELETE FROM user_notification WHERE user_id = ?`, [req.userId]);

    return new NoContentResponse();
}