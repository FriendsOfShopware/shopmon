import { and, eq } from 'drizzle-orm';
import { type Drizzle, getConnection, schema } from '../db.ts';

async function existsByEmail(
    con: Drizzle,
    email: string,
): Promise<string | null> {
    const result = await con.query.user.findFirst({
        columns: {
            id: true,
        },
        where: eq(schema.user.email, email),
    });

    return result ? result.id : null;
}

async function existsById(con: Drizzle, id: string): Promise<boolean> {
    const result = await con.query.user.findFirst({
        columns: {
            id: true,
        },
        where: eq(schema.user.id, id),
    });

    return result !== undefined;
}

async function deleteById(con: Drizzle, id: string): Promise<void> {
    await con
        .delete(schema.userNotification)
        .where(eq(schema.userNotification.userId, id))
        .execute();
}

async function createNotification(
    con: Drizzle,
    userId: string,
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

    const notificationResponse: typeof schema.userNotification.$inferSelect = {
        ...notification,
        key,
        userId,
        id: result.lastInsertRowid as unknown as number,
        read: false,
        createdAt: new Date(),
    };

    return notificationResponse;
}

async function hasAccessToProject(userId: string, projectId: number) {
    return await getConnection()
        .select({
            id: schema.project.id,
            organizationId: schema.project.organizationId,
        })
        .from(schema.project)
        .innerJoin(
            schema.organization,
            eq(schema.project.organizationId, schema.organization.id),
        )
        .innerJoin(
            schema.member,
            eq(schema.organization.id, schema.member.organizationId),
        )
        .where(
            and(
                eq(schema.project.id, projectId),
                eq(schema.member.userId, userId),
            ),
        )
        .get();
}

export default {
    existsByEmail,
    existsById,
    deleteById,
    createNotification,
    hasAccessToProject,
};
