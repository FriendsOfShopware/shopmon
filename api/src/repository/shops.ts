import { NotificationCreation } from '../../../frontend/src/types/notification';
import { UserSocketHelper } from '../object/UserSocket';
import Users from './users';
import { sendAlert } from '../mail/mail';
import { Drizzle, schema } from '../db';
import { eq } from 'drizzle-orm';
import type { Bindings } from '../router';

interface CreateShopRequest {
    team_id: number;
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

async function createShop(
    con: Drizzle,
    params: CreateShopRequest,
): Promise<number> {
    const result = await con
        .insert(schema.shop)
        .values({
            team_id: params.team_id,
            name: params.name,
            url: params.shop_url,
            client_id: params.client_id,
            client_secret: params.client_secret,
            created_at: new Date().toISOString(),
            shopware_version: params.version,
        })
        .execute();

    return result.meta.last_row_id;
}

async function deleteShop(con: Drizzle, id: number): Promise<void> {
    await con.delete(schema.shop).where(eq(schema.shop.id, id)).execute();
    await con
        .delete(schema.shopScrapeInfo)
        .where(eq(schema.shopScrapeInfo.shop, id))
        .execute();
}

async function getUsersOfShop(con: Drizzle, shopId: number) {
    const result = await con
        .select({
            id: schema.userToTeam.user_id,
            username: schema.user.username,
            email: schema.user.email,
        })
        .from(schema.userToTeam)
        .innerJoin(
            schema.shop,
            eq(schema.shop.team_id, schema.userToTeam.team_id),
        )
        .innerJoin(
            schema.user,
            eq(schema.user.id, schema.userToTeam.user_id),
        )
        .where(eq(schema.shop.id, shopId))
        .all();

    return result as User[];
}

async function notify(
    con: Drizzle,
    namespace: DurableObjectNamespace,
    shopId: number,
    key: string,
    notification: NotificationCreation,
): Promise<void> {
    const users = await getUsersOfShop(con, shopId);

    for (const user of users) {
        const createdNotification = await Users.createNotification(
            con,
            user.id,
            key,
            notification,
        );

        await UserSocketHelper.sendNotification(
            namespace,
            user.id.toString(),
            {
                notification: createdNotification,
            },
        );
    }
}

async function alert(
    con: Drizzle,
    env: Bindings,
    alert: ShopAlert,
): Promise<void> {
    const users = await getUsersOfShop(con, parseInt(alert.shopId));
    const alertKey = `alert_${alert.key}_${alert.shopId}`;

    const foundAlert = await env.kvStorage.get(alertKey);

    if (foundAlert !== null) {
        return;
    }

    await env.kvStorage.put(alertKey, '1', {
        expirationTtl: 86400, // one day
    });

    const result = await con.query.shop.findFirst({
        columns: {
            name: true,
        },
        where: eq(schema.shop.id, parseInt(alert.shopId)),
    });

    if (result === undefined) {
        return;
    }

    for (const user of users) {
        await sendAlert(env, result, user, alert);
    }
}

export default  {
    createShop,
    deleteShop,
    getUsersOfShop,
    notify,
    alert,
}
