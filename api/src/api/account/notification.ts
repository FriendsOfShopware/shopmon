import { getConnection, schema } from "../../db";
import { JsonResponse, NoContentResponse } from "../common/response";
import { eq, and } from "drizzle-orm";

export async function getNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const results = await con.query.userNotification.findMany({ where: eq(schema.userNotification.user_id, req.userId) })

    if (results === undefined) {
        return new JsonResponse([]);
    }

    for (const row of results) {
        row.link = JSON.parse(row.link);
        // @ts-ignore
        row.read = !!row.read;
    }

    return new JsonResponse(results);
}

export async function markAllReadNotification(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    await con.update(schema.userNotification).set({ read: 1 }).where(eq(schema.userNotification.user_id, req.userId)).execute();

    return new NoContentResponse();
}

export async function deleteNotification(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);
    const { id } = req.params;

    await con.delete(schema.userNotification).where(
        and(eq(schema.userNotification.id, parseInt(id)), eq(schema.userNotification.user_id, req.userId))
    ).execute();

    return new NoContentResponse();
}

export async function deleteAllNotifications(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    await con.delete(schema.userNotification).where(
        eq(schema.userNotification.user_id, req.userId)
    ).execute();

    return new NoContentResponse();
}
