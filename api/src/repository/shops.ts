import { promises as fs } from 'node:fs';
import path from 'node:path';
import { eq } from 'drizzle-orm';
import { type Drizzle, getConnection, schema } from '../db.ts';
import { sendAlert } from '../mail/mail.ts';
import * as LockRepository from './lock.ts';
import { deleteShopScrapeInfo } from './scrapeInfo.ts';
import Users from './users.ts';

interface CreateShopRequest {
    organizationId: string;
    name: string;
    version: string;
    shopUrl: string;
    clientId: string;
    clientSecret: string;
    projectId: number;
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
        .execute();

    return Number(result.lastInsertRowid);
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

    await deleteShopScrapeInfo(id);

    // Clean up sitespeed results from filesystem
    try {
        const sitespeedDataFolder = './files/sitespeed';
        const shopResultsDir = path.join(sitespeedDataFolder, id.toString());

        // Check if the directory exists before trying to delete it
        try {
            await fs.access(shopResultsDir);
            await fs.rm(shopResultsDir, { recursive: true, force: true });
            console.log(`Cleaned up sitespeed results for shop ${id}`);
        } catch (_accessError) {
            // Directory doesn't exist, which is fine
            console.log(
                `No sitespeed results found for shop ${id} - nothing to clean up`,
            );
        }
    } catch (error) {
        // Log the error but don't fail the shop deletion
        console.error(
            `Failed to clean up sitespeed results for shop ${id}:`,
            error,
        );
    }
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
        .where(eq(schema.shop.id, shopId))
        .all();

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

    if (await LockRepository.isLocked(alertKey)) {
        return;
    }

    await LockRepository.createLock(alertKey);

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
    deleteShopsByOrganization,
    notify,
    alert,
};
