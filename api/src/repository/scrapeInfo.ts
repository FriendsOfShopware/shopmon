import { and, eq } from 'drizzle-orm';
import { getConnection, schema } from '#src/db.ts';
import type { ShopScrapeInfo } from '#src/types/index.ts';

export async function getShopScrapeInfo(
    shopId: number,
): Promise<ShopScrapeInfo | null> {
    const con = getConnection();

    // Get scrape info
    const scrapeInfo = await con
        .select()
        .from(schema.shopScrapeInfo)
        .where(eq(schema.shopScrapeInfo.shopId, shopId))
        .limit(1);

    if (scrapeInfo.length === 0) {
        return null;
    }

    // Get extensions
    const extensions = await con
        .select()
        .from(schema.shopExtension)
        .where(eq(schema.shopExtension.shopId, shopId));

    // Get scheduled tasks
    const scheduledTasks = await con
        .select()
        .from(schema.shopScheduledTask)
        .where(eq(schema.shopScheduledTask.shopId, shopId));

    // Get queue info
    const queueInfo = await con
        .select()
        .from(schema.shopQueueInfo)
        .where(eq(schema.shopQueueInfo.shopId, shopId));

    // Get checks
    const checks = await con
        .select()
        .from(schema.shopCheck)
        .where(eq(schema.shopCheck.shopId, shopId));

    const info = scrapeInfo[0];

    return {
        extensions: extensions.map((ext) => ({
            name: ext.name,
            label: ext.label,
            active: ext.active,
            version: ext.version,
            latestVersion: ext.latestVersion,
            installed: ext.installed,
            ratingAverage: ext.ratingAverage,
            storeLink: ext.storeLink,
            changelog: ext.changelog,
            installedAt: ext.installedAt?.toISOString() ?? null,
        })),
        scheduledTask: scheduledTasks.map((task) => ({
            id: task.taskId,
            name: task.name,
            status: task.status,
            interval: task.interval,
            overdue: task.overdue,
            lastExecutionTime: task.lastExecutionTime?.toISOString() ?? '',
            nextExecutionTime: task.nextExecutionTime?.toISOString() ?? '',
        })),
        queueInfo: queueInfo.map((q) => ({
            name: q.name,
            size: q.size,
        })),
        cacheInfo: {
            environment: info.cacheEnvironment ?? '',
            httpCache: info.cacheHttpCache ?? false,
            cacheAdapter: info.cacheAdapter ?? '',
        },
        checks: checks.map((check) => ({
            id: check.checkId,
            level: check.level as 'green' | 'yellow' | 'red',
            message: check.message,
            source: check.source,
            link: check.link,
        })),
        createdAt: info.createdAt,
    };
}

export async function saveShopScrapeInfo(shopId: number, info: ShopScrapeInfo) {
    const con = getConnection();
    const now = new Date();

    // Use a transaction to ensure consistency
    await con.transaction(async (tx) => {
        // Upsert scrape info
        const existingScrapeInfo = await tx
            .select({ id: schema.shopScrapeInfo.id })
            .from(schema.shopScrapeInfo)
            .where(eq(schema.shopScrapeInfo.shopId, shopId))
            .limit(1);

        if (existingScrapeInfo.length > 0) {
            await tx
                .update(schema.shopScrapeInfo)
                .set({
                    cacheEnvironment: info.cacheInfo.environment,
                    cacheHttpCache: info.cacheInfo.httpCache,
                    cacheAdapter: info.cacheInfo.cacheAdapter,
                    updatedAt: now,
                })
                .where(eq(schema.shopScrapeInfo.shopId, shopId));
        } else {
            await tx.insert(schema.shopScrapeInfo).values({
                shopId,
                cacheEnvironment: info.cacheInfo.environment,
                cacheHttpCache: info.cacheInfo.httpCache,
                cacheAdapter: info.cacheInfo.cacheAdapter,
                createdAt: now,
                updatedAt: now,
            });
        }

        // Delete existing related data
        await tx
            .delete(schema.shopExtension)
            .where(eq(schema.shopExtension.shopId, shopId));
        await tx
            .delete(schema.shopScheduledTask)
            .where(eq(schema.shopScheduledTask.shopId, shopId));
        await tx
            .delete(schema.shopQueueInfo)
            .where(eq(schema.shopQueueInfo.shopId, shopId));
        await tx
            .delete(schema.shopCheck)
            .where(eq(schema.shopCheck.shopId, shopId));

        // Insert extensions
        if (info.extensions.length > 0) {
            await tx.insert(schema.shopExtension).values(
                info.extensions.map((ext) => ({
                    shopId,
                    name: ext.name,
                    label: ext.label,
                    active: ext.active,
                    version: ext.version,
                    latestVersion: ext.latestVersion,
                    installed: ext.installed,
                    ratingAverage: ext.ratingAverage,
                    storeLink: ext.storeLink,
                    changelog: ext.changelog,
                    installedAt: ext.installedAt ? new Date(ext.installedAt) : null,
                })),
            );
        }

        // Insert scheduled tasks
        if (info.scheduledTask.length > 0) {
            await tx.insert(schema.shopScheduledTask).values(
                info.scheduledTask.map((task) => ({
                    shopId,
                    taskId: task.id,
                    name: task.name,
                    status: task.status,
                    interval: task.interval,
                    overdue: task.overdue,
                    lastExecutionTime: task.lastExecutionTime
                        ? new Date(task.lastExecutionTime)
                        : null,
                    nextExecutionTime: task.nextExecutionTime
                        ? new Date(task.nextExecutionTime)
                        : null,
                })),
            );
        }

        // Insert queue info
        if (info.queueInfo.length > 0) {
            await tx.insert(schema.shopQueueInfo).values(
                info.queueInfo.map((q) => ({
                    shopId,
                    name: q.name,
                    size: q.size,
                })),
            );
        }

        // Insert checks
        if (info.checks.length > 0) {
            await tx.insert(schema.shopCheck).values(
                info.checks.map((check) => ({
                    shopId,
                    checkId: check.id,
                    level: check.level,
                    message: check.message,
                    source: check.source,
                    link: check.link,
                })),
            );
        }
    });
}

export async function deleteShopScrapeInfo(shopId: number) {
    const con = getConnection();

    // With ON DELETE CASCADE, we only need to delete the main record
    // But let's be explicit for clarity and safety
    await con.transaction(async (tx) => {
        await tx
            .delete(schema.shopExtension)
            .where(eq(schema.shopExtension.shopId, shopId));
        await tx
            .delete(schema.shopScheduledTask)
            .where(eq(schema.shopScheduledTask.shopId, shopId));
        await tx
            .delete(schema.shopQueueInfo)
            .where(eq(schema.shopQueueInfo.shopId, shopId));
        await tx
            .delete(schema.shopCheck)
            .where(eq(schema.shopCheck.shopId, shopId));
        await tx
            .delete(schema.shopScrapeInfo)
            .where(eq(schema.shopScrapeInfo.shopId, shopId));
    });
}

// Helper function to update a single scheduled task (used for reschedule)
export async function updateScheduledTask(
    shopId: number,
    taskId: string,
    updates: {
        status?: string;
        nextExecutionTime?: string;
        overdue?: boolean;
    },
) {
    const con = getConnection();

    const updateData: {
        status?: string;
        nextExecutionTime?: Date;
        overdue?: boolean;
    } = {};

    if (updates.status !== undefined) {
        updateData.status = updates.status;
    }
    if (updates.nextExecutionTime !== undefined) {
        updateData.nextExecutionTime = new Date(updates.nextExecutionTime);
    }
    if (updates.overdue !== undefined) {
        updateData.overdue = updates.overdue;
    }

    await con
        .update(schema.shopScheduledTask)
        .set(updateData)
        .where(
            and(
                eq(schema.shopScheduledTask.shopId, shopId),
                eq(schema.shopScheduledTask.taskId, taskId),
            ),
        );
}
