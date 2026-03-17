import { eq } from "drizzle-orm";
import { type Drizzle, getConnection, schema } from "#src/db.ts";
import * as LockRepository from "#src/modules/lock/lock.repository.ts";
import * as UptimeRepository from "./uptime.repository.ts";
import * as ShopService from "#src/modules/shop/shop.service.ts";

const CONSECUTIVE_FAILS_THRESHOLD = 3;

export async function performUptimeCheck(shopId: number) {
  const db = getConnection();

  const shopData = await db.query.shop.findFirst({
    columns: {
      id: true,
      name: true,
      url: true,
      uptimeStatus: true,
      uptimeConsecutiveFails: true,
      organizationId: true,
    },
    where: eq(schema.shop.id, shopId),
  });

  if (!shopData) {
    return;
  }

  const result = await checkUrl(shopData.url);

  await UptimeRepository.createCheck(db, shopId, result);

  const previousStatus = shopData.uptimeStatus;
  const newConsecutiveFails = result.isUp ? 0 : shopData.uptimeConsecutiveFails + 1;
  const isDown = newConsecutiveFails >= CONSECUTIVE_FAILS_THRESHOLD;
  const newStatus = result.isUp
    ? "up"
    : isDown
      ? "down"
      : previousStatus === "down"
        ? "down"
        : "unknown";

  const updates: Parameters<typeof UptimeRepository.updateShopUptimeStatus>[2] = {
    uptimeStatus: newStatus,
    uptimeLastCheckedAt: new Date(),
    uptimeConsecutiveFails: newConsecutiveFails,
  };

  // Shop just went down
  if (newStatus === "down" && previousStatus !== "down") {
    updates.uptimeDownSince = new Date();
    await notifyShopDown(db, shopData);
  }

  // Shop recovered
  if (result.isUp && previousStatus === "down") {
    updates.uptimeDownSince = null;
    await notifyShopRecovered(db, shopData);
  }

  await UptimeRepository.updateShopUptimeStatus(db, shopId, updates);
}

async function checkUrl(url: string): Promise<UptimeRepository.UptimeCheckResult> {
  const startTime = performance.now();

  try {
    const controller = new AbortController();
    const timeout = setTimeout(() => controller.abort(), 10000);

    const response = await fetch(url, {
      method: "GET",
      signal: controller.signal,
      redirect: "follow",
      headers: {
        "User-Agent": "Shopmon Uptime Monitor",
      },
    });

    clearTimeout(timeout);

    const ttfb = Math.round(performance.now() - startTime);
    const responseTime = Math.round(performance.now() - startTime);

    return {
      statusCode: response.status,
      ttfb,
      responseTime,
      isUp: response.status >= 200 && response.status < 500,
      error: response.status >= 500 ? `HTTP ${response.status}` : null,
    };
  } catch (e) {
    const responseTime = Math.round(performance.now() - startTime);

    return {
      statusCode: null,
      ttfb: null,
      responseTime,
      isUp: false,
      error: e instanceof Error ? e.message : String(e),
    };
  }
}

async function notifyShopDown(db: Drizzle, shop: { id: number; name: string }) {
  const alertKey = `uptime-down.${shop.id}`;

  if (await LockRepository.isLocked(alertKey)) {
    return;
  }

  await LockRepository.createLock(alertKey, 3600);

  await ShopService.notify(db, shop.id, `shop.uptime-down.${shop.id}`, {
    level: "error",
    title: `Shop ${shop.name} is down`,
    message: `The storefront of ${shop.name} is not reachable. Please check your server.`,
    link: {
      name: "account.shops.detail.uptime",
      params: { shopId: shop.id.toString() },
    },
  });

  await ShopService.alert(db, {
    key: alertKey,
    shopId: shop.id.toString(),
    subject: `Shop ${shop.name} is down`,
    message: `The storefront of ${shop.name} is not reachable. Please check your server immediately.`,
  });
}

async function notifyShopRecovered(db: Drizzle, shop: { id: number; name: string }) {
  await ShopService.notify(db, shop.id, `shop.uptime-recovered.${shop.id}`, {
    level: "info",
    title: `Shop ${shop.name} is back up`,
    message: `The storefront of ${shop.name} is reachable again.`,
    link: {
      name: "account.shops.detail.uptime",
      params: { shopId: shop.id.toString() },
    },
  });
}

const BATCH_CONCURRENCY = 50;

export async function checkAllShops() {
  const shops = await UptimeRepository.getShopsWithUptimeEnabled();

  console.log(`Uptime check: checking ${shops.length} shops directly`);

  // Process in batches to avoid overwhelming outbound connections
  for (let i = 0; i < shops.length; i += BATCH_CONCURRENCY) {
    const batch = shops.slice(i, i + BATCH_CONCURRENCY);
    await Promise.allSettled(batch.map((shop) => performUptimeCheck(shop.id)));
  }
}

export async function getUptimeData(db: Drizzle, shopId: number) {
  const [checks, stats, dailyStats, shopData] = await Promise.all([
    UptimeRepository.getRecentChecks(db, shopId),
    UptimeRepository.getUptimeStats(db, shopId),
    UptimeRepository.getDailyStats(db, shopId),
    db.query.shop.findFirst({
      columns: {
        uptimeEnabled: true,
        uptimeStatus: true,
        uptimeLastCheckedAt: true,
        uptimeDownSince: true,
      },
      where: eq(schema.shop.id, shopId),
    }),
  ]);

  return {
    enabled: shopData?.uptimeEnabled ?? false,
    status: shopData?.uptimeStatus ?? "unknown",
    lastCheckedAt: shopData?.uptimeLastCheckedAt ?? null,
    downSince: shopData?.uptimeDownSince ?? null,
    stats,
    checks,
    dailyStats,
  };
}

export async function updateUptimeSettings(db: Drizzle, shopId: number, enabled: boolean) {
  await db
    .update(schema.shop)
    .set({ uptimeEnabled: enabled })
    .where(eq(schema.shop.id, shopId))
    .execute();

  if (!enabled) {
    await db.delete(schema.shopUptime).where(eq(schema.shopUptime.shopId, shopId)).execute();

    await db
      .delete(schema.shopUptimeDaily)
      .where(eq(schema.shopUptimeDaily.shopId, shopId))
      .execute();

    await db
      .update(schema.shop)
      .set({
        uptimeStatus: "unknown",
        uptimeConsecutiveFails: 0,
        uptimeDownSince: null,
        uptimeLastCheckedAt: null,
      })
      .where(eq(schema.shop.id, shopId))
      .execute();
  }
}
