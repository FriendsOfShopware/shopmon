import { and, eq } from 'drizzle-orm';
import { type Drizzle, schema } from '#src/db.ts';

async function findAllByUserId(con: Drizzle, userId: string) {
    return await con.query.userNotification.findMany({
        where: eq(schema.userNotification.userId, userId),
    });
}

async function deleteById(con: Drizzle, userId: string, id: number) {
    await con
        .delete(schema.userNotification)
        .where(
            and(
                eq(schema.userNotification.id, id),
                eq(schema.userNotification.userId, userId),
            ),
        )
        .execute();
}

async function deleteAllByUserId(con: Drizzle, userId: string) {
    await con
        .delete(schema.userNotification)
        .where(eq(schema.userNotification.userId, userId))
        .execute();
}

async function markAllRead(con: Drizzle, userId: string) {
    await con
        .update(schema.userNotification)
        .set({ read: true })
        .where(eq(schema.userNotification.userId, userId))
        .execute();
}

export default {
    findAllByUserId,
    deleteById,
    deleteAllByUserId,
    markAllRead,
};
