import { Connection } from "@planetscale/database/dist";
import { Notification, NotificationCreation } from "../../../shared/notification";
import Teams from "./teams";

export default class Users {
    static async delete(con: Connection, id: string): Promise<void> {
        const ownerTeams = await con.execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);

        const deletePromises = ownerTeams.rows.map(async row => 
            Teams.deleteTeam(con, row.id) 
        );
        
        await Promise.all(deletePromises);

        await con.execute(`DELETE FROM user WHERE id = ?`, [id]);
        await con.execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }

    static async revokeUserSessions(session: KVNamespace, id: string): Promise<void> {
        const accessToken = await session.list({ prefix: `u-${id}-` });
    
        for (const key of accessToken.keys) {
            await session.delete(key.name);
        }

        const refreshToken = await session.list({ prefix: `r-${id}-` });
    
        for (const key of refreshToken.keys) {
            await session.delete(key.name);
        }
    }

    static async createNotification(con: Connection, userId: number, key: string, notification: NotificationCreation): Promise<Notification> {
        await con.execute('INSERT INTO user_notification (user_id, `key`, level, title, message, link, `read`) VALUES (?, ?, ?, ?, ?, ?, 0) ON DUPLICATE KEY UPDATE `read` = 0', [
            userId,
            key,
            notification.level,
            notification.title,
            notification.message,
            JSON.stringify(notification.link)
        ]);

        const result = await con.execute('SELECT id from user_notification WHERE user_id = ? AND `key` = ?', [userId, key]);

        const noti = notification as Notification;
        noti.read = false;
        noti.id = result.rows[0].id;

        return noti;
    }
}