import { promises as fs } from "node:fs";
import path from "node:path";
import { HttpClient, type HttpClientResponse, SimpleShop } from "@shopware-ag/app-server-sdk";
import { TRPCError } from "@trpc/server";
import { and, desc, eq, sql } from "drizzle-orm";
import {
  type Drizzle,
  schema,
  shopCache,
  shopCheck,
  shopExtension,
  shopQueue,
  shopScheduledTask,
} from "#src/db.ts";
import * as LockRepository from "#src/modules/lock/lock.repository.ts";
import { scrapeSingleShop } from "#src/modules/shop/jobs/shop-scrape.job.ts";
import { scrapeSingleSitespeedShop } from "#src/modules/shop/jobs/sitespeed-scrape.job.ts";
import { sendAlert } from "#src/modules/shop/mail/mail.service.ts";
import { deleteDeploymentOutputsByShopId } from "#src/modules/deployment/deployment.storage.ts";
import { deleteShopScrapeInfo } from "#src/modules/shop/scrape-info.repository.ts";
import { deleteSitespeedReport, getReportUrl } from "#src/modules/shop/sitespeed.service.ts";
import Users from "#src/modules/user/user.repository.ts";
import { decrypt, encrypt } from "./crypto.ts";
import Shops from "./shop.repository.ts";

export interface CreateShopInput {
  name: string;
  shopUrl: string;
  clientId: string;
  clientSecret: string;
  projectId: number;
}

export interface UpdateShopInput {
  shopId: number;
  name?: string;
  shopUrl?: string;
  clientId?: string;
  clientSecret?: string;
  ignores?: string[];
  projectId?: number;
}

export interface ShopAlert {
  shopId: string;
  key: string;
  subject: string;
  message: string;
}

export const listShops = async (db: Drizzle, orgId: string) => {
  const result = await db
    .select({
      id: schema.shop.id,
      name: schema.shop.name,
      url: schema.shop.url,
      favicon: schema.shop.favicon,
      createdAt: schema.shop.createdAt,
      lastScrapedAt: schema.shop.lastScrapedAt,
      status: schema.shop.status,
      lastScrapedError: schema.shop.lastScrapedError,
      shopwareVersion: schema.shop.shopwareVersion,
      projectId: schema.shop.projectId,
      project: schema.project
        ? {
            id: schema.project.id,
            name: schema.project.name,
            description: schema.project.description,
          }
        : null,
    })
    .from(schema.shop)
    .leftJoin(schema.project, eq(schema.project.id, schema.shop.projectId))
    .where(eq(schema.shop.organizationId, orgId));

  return result === undefined ? [] : result;
};

export const getShopDetails = async (db: Drizzle, shopId: number) => {
  const shopQuery = db
    .select({
      id: schema.shop.id,
      name: schema.shop.name,
      nameCombined: sql<string>`CONCAT(${schema.project.name}, ' / ', ${schema.shop.name})`,
      url: schema.shop.url,
      status: schema.shop.status,
      createdAt: schema.shop.createdAt,
      shopwareVersion: schema.shop.shopwareVersion,
      lastScrapedAt: schema.shop.lastScrapedAt,
      lastScrapedError: schema.shop.lastScrapedError,
      lastChangelog: schema.shop.lastChangelog,
      ignores: schema.shop.ignores,
      shopImage: schema.shop.shopImage,
      connectionIssueCount: schema.shop.connectionIssueCount,
      organizationId: schema.shop.organizationId,
      organizationName: sql<string>`${schema.organization.name}`.as("organization_name"),
      projectId: schema.shop.projectId,
      projectName: sql<string>`${schema.project.name}`.as("project_name"),
      projectDescription: schema.project.description,
      sitespeedEnabled: schema.shop.sitespeedEnabled,
      sitespeedUrls: schema.shop.sitespeedUrls,
    })
    .from(schema.shop)
    .innerJoin(schema.organization, eq(schema.organization.id, schema.shop.organizationId))
    .leftJoin(schema.project, eq(schema.project.id, schema.shop.projectId))
    .where(eq(schema.shop.id, shopId))
    .then((rows) => rows[0]);

  const sitespeedQuery = db.query.shopSitespeed.findMany({
    where: eq(schema.shopSitespeed.shopId, shopId),
    orderBy: [desc(schema.shopSitespeed.createdAt)],
  });

  const shopChangelogQuery = db.query.shopChangelog.findMany({
    where: eq(schema.shopChangelog.shopId, shopId),
    orderBy: [desc(schema.shopChangelog.date)],
  });

  const extensionsQuery = db.select().from(shopExtension).where(eq(shopExtension.shopId, shopId));

  const scheduledTasksQuery = db
    .select()
    .from(shopScheduledTask)
    .where(eq(shopScheduledTask.shopId, shopId));

  const queuesQuery = db.select().from(shopQueue).where(eq(shopQueue.shopId, shopId));

  const cacheQuery = db.select().from(shopCache).where(eq(shopCache.shopId, shopId)).limit(1);

  const checksQuery = db.select().from(shopCheck).where(eq(shopCheck.shopId, shopId));

  const [shop, sitespeed, shopChangelog, extensions, scheduledTasks, queues, cache, checks] =
    await Promise.all([
      shopQuery,
      sitespeedQuery,
      shopChangelogQuery,
      extensionsQuery,
      scheduledTasksQuery,
      queuesQuery,
      cacheQuery,
      checksQuery,
    ]);

  if (shop === undefined) {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "Not Found.",
    });
  }

  return {
    ...shop,
    extensions: extensions.map((ext) => ({
      name: ext.name,
      label: ext.label,
      active: ext.active,
      version: ext.version,
      latestVersion: ext.latestVersion,
      installed: ext.installed,
      ratingAverage: ext.ratingAverage,
      storeLink: ext.storeLink,
      changelog: ext.changelog ?? null,
      installedAt: ext.installedAt,
    })),
    scheduledTask: scheduledTasks.map((task) => ({
      id: task.taskId,
      name: task.name,
      status: task.status,
      interval: task.interval,
      overdue: task.overdue,
      lastExecutionTime: task.lastExecutionTime,
      nextExecutionTime: task.nextExecutionTime,
    })),
    queueInfo: queues.map((queue) => ({
      name: queue.name,
      size: queue.size,
    })),
    cacheInfo: cache[0]
      ? {
          environment: cache[0].environment,
          httpCache: cache[0].httpCache,
          cacheAdapter: cache[0].cacheAdapter,
        }
      : null,
    checks: checks.map((check) => ({
      id: check.checkId,
      level: check.level as "green" | "yellow" | "red",
      message: check.message,
      source: check.source,
      link: check.link,
    })),
    sitespeed: sitespeed,
    sitespeedReportUrl: getReportUrl(shopId),
    changelog: shopChangelog,
  };
};

export const isSubscribedToNotifications = async (db: Drizzle, userId: string, shopId: number) => {
  const user = await db.query.user.findFirst({
    columns: {
      notifications: true,
    },
    where: eq(schema.user.id, userId),
  });

  if (!user) {
    return false;
  }

  const shopKey = `shop-${shopId}`;
  return (user.notifications || []).includes(shopKey);
};

export const create = async (db: Drizzle, userId: string, input: CreateShopInput) => {
  const project = await Users.hasAccessToProject(userId, input.projectId);

  if (project === null) {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "You do not have access to this project",
    });
  }

  const shop = new SimpleShop("", input.shopUrl, "");
  shop.setShopCredentials(input.clientId, input.clientSecret);

  const client = new HttpClient(shop);

  let resp: HttpClientResponse<{ version: string }>;
  try {
    resp = await client.get("/_info/config");
  } catch {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "Cannot reach shop. Check your credentials and shop URL.",
    });
  }

  const clientSecret = await encrypt(process.env.APP_SECRET, input.clientSecret);

  const id = await Shops.createShop(db, {
    organizationId: project.organizationId,
    name: input.name,
    clientId: input.clientId,
    clientSecret: clientSecret,
    shopUrl: input.shopUrl,
    version: resp.body.version,
    projectId: input.projectId,
  });

  await scrapeSingleShop(id);

  return id;
};

export const deleteShop = async (db: Drizzle, shopId: number) => {
  try {
    await deleteDeploymentOutputsByShopId(db, shopId);
  } catch (error) {
    console.error(`Failed to clean up deployment outputs for shop ${shopId}:`, error);
  }

  await Shops.deleteShop(db, shopId);
  await deleteShopScrapeInfo(shopId);

  console.log(`Shop ${shopId} deleted`);

  try {
    await deleteSitespeedReport(shopId);
  } catch (error) {
    console.error(`Failed to clean up sitespeed results for shop ${shopId}:`, error);
  }

};

export const update = async (db: Drizzle, userId: string, input: UpdateShopInput) => {
  if (input.name) {
    await db
      .update(schema.shop)
      .set({ name: input.name })
      .where(eq(schema.shop.id, input.shopId))
      .execute();
  }

  if (input.ignores) {
    await db
      .update(schema.shop)
      .set({ ignores: input.ignores })
      .where(eq(schema.shop.id, input.shopId))
      .execute();
  }

  if (input.projectId) {
    const project = await Users.hasAccessToProject(userId, input.projectId);

    if (project === null) {
      throw new TRPCError({
        code: "BAD_REQUEST",
        message: "You do not have access to this project",
      });
    }

    await db
      .update(schema.shop)
      .set({
        organizationId: project.organizationId,
        projectId: project.id,
      })
      .where(eq(schema.shop.id, input.shopId))
      .execute();
  }

  if (input.shopUrl && input.clientId && input.clientSecret) {
    // Try out the new credentials
    const shop = new SimpleShop("", input.shopUrl, "");
    shop.setShopCredentials(input.clientId, input.clientSecret);

    const client = new HttpClient(shop);

    try {
      await client.get("/_info/config");
    } catch {
      throw new TRPCError({
        code: "BAD_REQUEST",
        message: "Cannot reach shop. Check your credentials and shop URL.",
      });
    }

    const clientSecret = await encrypt(process.env.APP_SECRET, input.clientSecret);

    await db
      .update(schema.shop)
      .set({
        url: input.shopUrl,
        clientId: input.clientId,
        clientSecret: clientSecret,
        connectionIssueCount: 0,
      })
      .where(eq(schema.shop.id, input.shopId))
      .execute();
  }
};

export const refresh = async (shopId: number, sitespeed?: boolean) => {
  await scrapeSingleShop(shopId);

  if (sitespeed) {
    await scrapeSingleSitespeedShop(shopId);
  }
};

export const clearCache = async (db: Drizzle, shopId: number) => {
  const shopData = await db.query.shop.findFirst({
    columns: {
      url: true,
      clientId: true,
      clientSecret: true,
    },
    where: eq(schema.shop.id, shopId),
  });

  if (shopData === undefined) {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "Not Found.",
    });
  }

  const clientSecret = await decrypt(process.env.APP_SECRET, shopData.clientSecret);
  const shop = new SimpleShop("", shopData.url, "");
  shop.setShopCredentials(shopData.clientId, clientSecret);
  const client = new HttpClient(shop);

  await client.delete("/_action/cache");
};

export const rescheduleTask = async (db: Drizzle, shopId: number, taskId: string) => {
  const shopData = await db.query.shop.findFirst({
    columns: {
      url: true,
      clientId: true,
      clientSecret: true,
    },
    where: eq(schema.shop.id, shopId),
  });

  if (shopData === undefined) {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "Not Found.",
    });
  }

  const clientSecret = await decrypt(process.env.APP_SECRET, shopData.clientSecret);
  const shop = new SimpleShop("", shopData.url, "");
  shop.setShopCredentials(shopData.clientId, clientSecret);
  const client = new HttpClient(shop);

  const nextExecutionTime: string = new Date().toISOString();
  await client.patch(`/scheduled-task/${taskId}`, {
    status: "scheduled",
    nextExecutionTime: nextExecutionTime,
  });

  // Update the scheduled task directly in the database
  await db
    .update(shopScheduledTask)
    .set({
      status: "scheduled",
      nextExecutionTime: nextExecutionTime,
      overdue: false,
    })
    .where(and(eq(shopScheduledTask.shopId, shopId), eq(shopScheduledTask.taskId, taskId)))
    .execute();
};

export const subscribeToNotifications = async (db: Drizzle, userId: string, shopId: number) => {
  const user = await db.query.user.findFirst({
    columns: {
      notifications: true,
    },
    where: eq(schema.user.id, userId),
  });

  if (!user) {
    throw new TRPCError({
      code: "NOT_FOUND",
      message: "User not found",
    });
  }

  const shopKey = `shop-${shopId}`;
  const notifications = user.notifications || [];

  if (!notifications.includes(shopKey)) {
    notifications.push(shopKey);
    await db.update(schema.user).set({ notifications }).where(eq(schema.user.id, userId)).execute();
  }
};

export const unsubscribeFromNotifications = async (db: Drizzle, userId: string, shopId: number) => {
  const user = await db.query.user.findFirst({
    columns: {
      notifications: true,
    },
    where: eq(schema.user.id, userId),
  });

  if (!user) {
    throw new TRPCError({
      code: "NOT_FOUND",
      message: "User not found",
    });
  }

  const shopKey = `shop-${shopId}`;
  const notifications = user.notifications || [];
  const index = notifications.indexOf(shopKey);

  if (index > -1) {
    notifications.splice(index, 1);
    await db.update(schema.user).set({ notifications }).where(eq(schema.user.id, userId)).execute();
  }
};

export const updateSitespeedSettings = async (
  db: Drizzle,
  shopId: number,
  enabled: boolean,
  urls?: string[],
) => {
  const updateData: {
    sitespeedEnabled: boolean;
    sitespeedUrls?: string[];
  } = {
    sitespeedEnabled: enabled,
  };

  if (urls !== undefined) {
    updateData.sitespeedUrls = urls;
  }

  await db.update(schema.shop).set(updateData).where(eq(schema.shop.id, shopId)).execute();

  if (!enabled) {
    await db.delete(schema.shopSitespeed).where(eq(schema.shopSitespeed.shopId, shopId)).execute();

    await fs.rm(path.join("./files/sitespeed", shopId.toString()), {
      recursive: true,
      force: true,
    });
  }
};

export const alert = async (db: Drizzle, alertData: ShopAlert) => {
  const users = await Shops.getUsersOfShop(db, Number.parseInt(alertData.shopId, 10));
  const alertKey = `alert_${alertData.key}_${alertData.shopId}`;

  if (await LockRepository.isLocked(alertKey)) {
    return;
  }

  await LockRepository.createLock(alertKey);

  const result = await db.query.shop.findFirst({
    columns: {
      name: true,
    },
    where: eq(schema.shop.id, Number.parseInt(alertData.shopId, 10)),
  });

  if (result === undefined) {
    return;
  }

  for (const user of users) {
    await sendAlert(result, user, alertData);
  }
};

export const notify = async (
  db: Drizzle,
  shopId: number,
  key: string,
  notification: Omit<typeof schema.userNotification.$inferInsert, "createdAt" | "key" | "userId">,
) => {
  const users = await Shops.getUsersOfShop(db, shopId);

  for (const user of users) {
    await Users.createNotification(db, user.id, key, notification);
  }
};
