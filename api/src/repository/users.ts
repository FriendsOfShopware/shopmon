import { Drizzle, getLastInsertId, schema } from '../db';
import { eq } from 'drizzle-orm';
import Organization from './organization';

async function existsByEmail(
    con: Drizzle,
    email: string,
): Promise<number | null> {
    const result = await con.query.user.findFirst({
        columns: {
            id: true,
        },
        where: eq(schema.user.email, email),
    });

    return result ? result.id : null;
}

async function existsById(con: Drizzle, id: number): Promise<boolean> {
    const result = await con.query.user.findFirst({
        columns: {
            id: true,
        },
        where: eq(schema.user.id, id),
    });

    return result !== undefined;
}

async function deleteById(con: Drizzle, id: number): Promise<void> {
    const ownerOrganizations = await con.query.organization.findMany({
        columns: {
            id: true,
        },
        where: eq(schema.organization.ownerId, id),
    });

    const deletePromises = ownerOrganizations.map((row) =>
        Organization.deleteById(con, row.id),
    );

    await Promise.all(deletePromises);

    await con.delete(schema.user).where(eq(schema.user.id, id)).execute();
    await con
        .delete(schema.userNotification)
        .where(eq(schema.userNotification.userId, id))
        .execute();

    await con
        .delete(schema.userToOrganization)
        .where(eq(schema.userToOrganization.userId, id))
        .execute();
}

async function revokeUserSessions(
    session: KVNamespace,
    id: number,
): Promise<void> {
    const accessToken = await session.list({ prefix: `u-${id}-` });

    for (const key of accessToken.keys) {
        await session.delete(key.name);
    }

    const refreshToken = await session.list({ prefix: `r-${id}-` });

    for (const key of refreshToken.keys) {
        await session.delete(key.name);
    }
}

async function createNotification(
    con: Drizzle,
    userId: number,
    key: string,
    notification: Omit<
        typeof schema.userNotification.$inferInsert,
        'createdAt' | 'key' | 'userId'
    >,
) {
    const result = await con
        .insert(schema.userNotification)
        .values({
            userId: userId,
            key,
            level: notification.level,
            title: notification.title,
            message: notification.message,
            link: notification.link,
            createdAt: new Date(),
        })
        .onConflictDoUpdate({
            target: [
                schema.userNotification.userId,
                schema.userNotification.key,
            ],
            set: {
                read: false,
            },
        })
        .execute();

    const lastId = getLastInsertId(result);

    const notificationResponse: typeof schema.userNotification.$inferSelect = {
        ...notification,
        key,
        userId,
        id: lastId,
        read: false,
        createdAt: new Date(),
    };

    return notificationResponse;
}

export default {
    existsByEmail,
    existsById,
    deleteById,
    revokeUserSessions,
    createNotification,
};
