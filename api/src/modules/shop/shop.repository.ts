import { eq } from 'drizzle-orm';
import { type Drizzle, getConnection, schema } from '#src/db.ts';

interface CreateShopRequest {
    organizationId: string;
    name: string;
    version: string;
    shopUrl: string;
    clientId: string;
    clientSecret: string;
    projectId: number;
}

export interface Shop {
    name: string;
}

export interface User {
    id: string;
    displayName: string;
    email: string;
    notifications?: string[];
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
            projectId: params.projectId,
        })
        .returning({ id: schema.shop.id });

    return result[0].id;
}

async function deleteShop(con: Drizzle, id: number): Promise<void> {
    // Delete database records
    await con
        .delete(schema.shopChangelog)
        .where(eq(schema.shopChangelog.shopId, id))
        .execute();
    await con
        .delete(schema.shopSitespeed)
        .where(eq(schema.shopSitespeed.shopId, id))
        .execute();
    await con.delete(schema.shop).where(eq(schema.shop.id, id)).execute();
}

async function getUsersOfShop(con: Drizzle, shopId: number) {
    const result = await con
        .select({
            id: schema.member.userId,
            displayName: schema.user.name,
            email: schema.user.email,
            notifications: schema.user.notifications,
        })
        .from(schema.shop)
        .innerJoin(
            schema.member,
            eq(schema.shop.organizationId, schema.member.organizationId),
        )
        .innerJoin(schema.user, eq(schema.user.id, schema.member.userId))
        .where(eq(schema.shop.id, shopId));

    // Filter users who have subscribed to notifications for this shop
    const shopKey = `shop-${shopId}`;
    return result.filter((user) => {
        const notifications = user.notifications || [];
        return notifications.includes(shopKey);
    });
}

async function deleteShopsByOrganization(organizationId: string) {
    const con = getConnection();
    const shops = await con
        .select({
            id: schema.shop.id,
        })
        .from(schema.shop)
        .where(eq(schema.shop.organizationId, organizationId));

    const promises = [];
    for (const shop of shops) {
        promises.push(deleteShop(con, shop.id));
    }
    await Promise.all(promises);
}

async function countByProject(
    con: Drizzle,
    projectId: number,
): Promise<number> {
    const result = await con
        .select({
            id: schema.shop.id,
        })
        .from(schema.shop)
        .where(eq(schema.shop.projectId, projectId));

    return result.length;
}

export default {
    createShop,
    deleteShop,
    getUsersOfShop,
    deleteShopsByOrganization,
    countByProject,
};
