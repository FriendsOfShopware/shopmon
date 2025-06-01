import { eq } from 'drizzle-orm';
import { type Drizzle, schema } from '../db.ts';
import { sendAlert } from '../mail/mail.ts';
import Users from './users.ts';

const alertMap = new Map<string, string>();

interface CreateShopRequest {
    organizationId: number;
    name: string;
    version: string;
    shopUrl: string;
    clientId: string;
    clientSecret: string;
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
    id: string;
    displayName: string;
    email: string;
}

async function createShop(
    con: Drizzle,
    params: CreateShopRequest,
): Promise<number> {
    const result = await con
        .insert(schema.shop)
        .values({
            organizationId: params.organizationId,
            name: params.name,
            url: params.shopUrl,
            clientId: params.clientId,
            clientSecret: params.clientSecret,
            createdAt: new Date(),
            shopwareVersion: params.version,
        })
        .execute();

    return result.lastInsertRowid as number;
}

async function deleteShop(con: Drizzle, id: number): Promise<void> {
    await con
        .delete(schema.shopScrapeInfo)
        .where(eq(schema.shopScrapeInfo.shopId, id))
        .execute();
    await con
        .delete(schema.shopChangelog)
        .where(eq(schema.shopChangelog.shopId, id))
        .execute();
    await con
        .delete(schema.shopPageSpeed)
        .where(eq(schema.shopPageSpeed.shopId, id))
        .execute();
    await con.delete(schema.shop).where(eq(schema.shop.id, id)).execute();
}

async function getUsersOfShop(con: Drizzle, shopId: number) {
    const result = await con
        .select({
            id: schema.userToOrganization.userId,
            displayName: schema.user.name,
            email: schema.user.email,
        })
        .from(schema.shop)
        .innerJoin(
            schema.userToOrganization,
            eq(
                schema.shop.organizationId,
                schema.userToOrganization.organizationId,
            ),
        )
        .innerJoin(
            schema.user,
            eq(schema.user.id, schema.userToOrganization.userId),
        )
        .where(eq(schema.shop.id, shopId))
        .all();

    return result;
}

async function notify(
    con: Drizzle,
    shopId: number,
    key: string,
    notification: Omit<
        typeof schema.userNotification.$inferInsert,
        'createdAt' | 'key' | 'userId'
    >,
): Promise<void> {
    const users = await getUsersOfShop(con, shopId);

    for (const user of users) {
        await Users.createNotification(con, user.id, key, notification);
    }
}

async function alert(con: Drizzle, alert: ShopAlert): Promise<void> {
    const users = await getUsersOfShop(con, Number.parseInt(alert.shopId));
    const alertKey = `alert_${alert.key}_${alert.shopId}`;

    if (alertMap.has(alertKey)) {
        return;
    }

    alertMap.set(alertKey, '1');

    const result = await con.query.shop.findFirst({
        columns: {
            name: true,
        },
        where: eq(schema.shop.id, Number.parseInt(alert.shopId)),
    });

    if (result === undefined) {
        return;
    }

    for (const user of users) {
        await sendAlert(result, user, alert);
    }
}

export default {
    createShop,
    deleteShop,
    getUsersOfShop,
    notify,
    alert,
};
