import { Notification, NotificationCreation } from "../../../shared/notification";
import { Drizzle, schema } from '../db'
import { eq } from 'drizzle-orm';
import Teams from "./teams";

export default class Users {
    static async existsByEmail(con: Drizzle, email: string): Promise<boolean> {
        const result = await con.query.user.findFirst({
            columns: {
                id: true,
            },
            where: eq(schema.user.email, email)
        })

        return result !== undefined;
    }

    static async existsById(con: Drizzle, id: number): Promise<boolean> {
        const result = await con.query.user.findFirst({
            columns: {
                id: true,
            },
            where: eq(schema.user.id, id)
        })

        return result !== undefined;
    }

    static async delete(con: Drizzle, id: number): Promise<void> {
        const ownerTeams = await con.query.team.findMany({
            columns: {
                id: true,
            },
            where: eq(schema.team.owner_id, id)
        });

        const deletePromises = ownerTeams.map(row =>
            Teams.deleteTeam(con, row.id)
        );

        await Promise.all(deletePromises);

        await con
            .delete(schema.user)
            .where(eq(schema.user.id, id))
            .execute();

        await con
            .delete(schema.userToTeam)
            .where(eq(schema.userToTeam.user_id, id))
            .execute();
    }

    static async revokeUserSessions(session: KVNamespace, id: number): Promise<void> {
        const accessToken = await session.list({ prefix: `u-${id}-` });

        for (const key of accessToken.keys) {
            await session.delete(key.name);
        }

        const refreshToken = await session.list({ prefix: `r-${id}-` });

        for (const key of refreshToken.keys) {
            await session.delete(key.name);
        }
    }

    static async createNotification(con: Drizzle, userId: number, key: string, notification: NotificationCreation): Promise<Notification> {
        const result = await con
            .insert(schema.userNotification)
            .values({
                user_id: userId,
                key,
                level: notification.level,
                title: notification.title,
                message: notification.message,
                link: JSON.stringify(notification.link),
                created_at: (new Date()).toISOString(),
            })
            .onConflictDoUpdate({
                target: [schema.userNotification.user_id, schema.userNotification.key],
                set: {
                    read: 0
                }
            })
            .execute();

        const noti = notification as Notification;
        noti.read = false;
        noti.id = result.meta.last_row_id;

        return noti;
    }
}
