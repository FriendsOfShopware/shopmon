import { Connection } from "@planetscale/database/dist";
import { NotificationCreation } from "../../../shared/notification";
import { UserSocketHelper } from "../object/UserSocket";
import Users from "./users";
import {sendAlert} from "../mail/mail";

interface CreateShopRequest {
    team_id: string;
    name: string;
    version: string;
    shop_url: string;
    client_id: string;
    client_secret: string;
}

export interface ShopAlert {
    shopId: string;
    key: string;
    subject: string;
    message: string;
}

export interface Shop {
    name: string;
}

export interface User {
    id: number;
    username: string;
    email: string;
}

export default class Shops {
    static async createShop(con: Connection, params: CreateShopRequest): Promise<string> {
        try {
            const result = await con.execute('INSERT INTO shop (team_id, name, url, client_id, client_secret, created_at, shopware_version) VALUES (?, ?, ?, ?, ?, NOW(), ?)', [
                params.team_id,
                params.name,
                params.shop_url,
                params.client_id,
                params.client_secret,
                params.version,
            ]);

            return result.insertId as string;
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        } catch (e: any) {
            if (e.error?.code == 'ALREADY_EXISTS') {
                throw new Error('Shop already exists.');
            }

            throw e;
        }
    }

    static async deleteShop(con: Connection, id: number): Promise<void> {
        await con.execute('DELETE FROM shop WHERE id = ?', [
            id
        ]);

        await con.execute('DELETE FROM shop_scrape_info WHERE shop_id = ?', [
            id
        ]);
    }

    static async getUsersOfShop(con: Connection, shopId: string) {
        const result = await con.execute(`
        SELECT
            user_to_team.user_id as id,
            user.username,
            user.email
        FROM
            user_to_team
        INNER JOIN shop ON(shop.team_id = user_to_team.team_id)
        INNER JOIN user ON(user.id = user_to_team.user_id)
        WHERE shop.id = ?
        `, [shopId])

        return result.rows as User[]
    }

    static async notify(con: Connection, namespace: DurableObjectNamespace, shopId: string, key: string, notification: NotificationCreation): Promise<void> {
        const users = await this.getUsersOfShop(con, shopId);

        for (const user of users) {
            const createdNotification = await Users.createNotification(con, user.id, key, notification);

            await UserSocketHelper.sendNotification(
                namespace,
                user.id.toString(),
                {
                    notification: createdNotification
                }
            )
        }
    }

    static async alert(con: Connection, env: Env, alert: ShopAlert): Promise<void> {
        const users = await this.getUsersOfShop(con, alert.shopId);
        const alertKey = `alert_${alert.key}_${alert.shopId}`;

        const foundAlert = await env.kvStorage.get(alertKey);

        if (foundAlert !== null) {
            return;
        }

        await env.kvStorage.put(alertKey, "1", {
            expirationTtl: 86400, // one day
        });

        const result = await con.execute(`SELECT name FROM shop WHERE id = ?`, [alert.shopId])
        const shop = result.rows[0] as Shop;

        for (const user of users) {
            await sendAlert(env, shop, user, alert)
        }
    }
}