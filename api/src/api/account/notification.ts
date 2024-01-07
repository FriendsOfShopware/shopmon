import { getConnection, userNotification as userNotificationTable } from "../../db";
import { JsonResponse, NoContentResponse } from "../common/response";
import { eq, and } from "drizzle-orm";

export async function getNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const results = await con.query.userNotification.findMany({ where: eq(userNotificationTable.user_id, parseInt(req.userId)) })

    if (results === undefined) {
        return new JsonResponse([]);
    }

    for (const row of results) {
        row.link = JSON.parse(row.link);
        row.read = !!row.read;
    }

    return new JsonResponse(results);
}

export async function markAllReadNotification(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    await con.update(userNotificationTable).set({ read: 1 }).where(eq(userNotificationTable.user_id, parseInt(req.userId))).execute();

    return new NoContentResponse();
}

export async function deleteNotification(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);
    const { id } = req.params;

    await con.delete(userNotificationTable).where(
        and(eq(userNotificationTable.id, parseInt(id)), eq(userNotificationTable.user_id, parseInt(req.userId)))
    ).execute();

    return new NoContentResponse();
}

export async function deleteAllNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    await con.delete(userNotificationTable).where(
        eq(userNotificationTable.user_id, parseInt(req.userId))
    ).execute();

    return new NoContentResponse();
}
