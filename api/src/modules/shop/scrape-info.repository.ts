import { eq } from 'drizzle-orm';
import {
    getConnection,
    shopCache,
    shopCheck,
    shopExtension,
    shopQueue,
    shopScheduledTask,
} from '#src/db.ts';

export async function deleteShopScrapeInfo(shopId: number) {
    const db = getConnection();

    await Promise.all([
        db.delete(shopExtension).where(eq(shopExtension.shopId, shopId)),
        db
            .delete(shopScheduledTask)
            .where(eq(shopScheduledTask.shopId, shopId)),
        db.delete(shopQueue).where(eq(shopQueue.shopId, shopId)),
        db.delete(shopCache).where(eq(shopCache.shopId, shopId)),
        db.delete(shopCheck).where(eq(shopCheck.shopId, shopId)),
    ]);
}
