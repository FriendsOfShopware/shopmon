import { and, asc, desc, eq, gte, lt, sql } from "drizzle-orm";
import { type Drizzle, getConnection, schema, shop, shopUptime, shopUptimeDaily } from "#src/db.ts";

export interface UptimeCheckResult {
  statusCode: number | null;
  ttfb: number | null;
  responseTime: number | null;
  isUp: boolean;
  error: string | null;
}

export async function createCheck(db: Drizzle, shopId: number, result: UptimeCheckResult) {
  await db
    .insert(shopUptime)
    .values({
      shopId,
      checkedAt: new Date(),
      statusCode: result.statusCode,
      ttfb: result.ttfb,
      responseTime: result.responseTime,
      isUp: result.isUp,
      error: result.error,
    })
    .execute();
}

export async function getRecentChecks(db: Drizzle, shopId: number, limit = 100) {
  return db
    .select()
    .from(shopUptime)
    .where(eq(shopUptime.shopId, shopId))
    .orderBy(desc(shopUptime.checkedAt))
    .limit(limit);
}

export async function getDailyStats(db: Drizzle, shopId: number, sinceDays = 365) {
  const since = new Date(Date.now() - sinceDays * 24 * 60 * 60 * 1000);

  return db
    .select()
    .from(shopUptimeDaily)
    .where(and(eq(shopUptimeDaily.shopId, shopId), gte(shopUptimeDaily.date, since)))
    .orderBy(asc(shopUptimeDaily.date));
}

export async function getUptimeStats(db: Drizzle, shopId: number, sinceDays = 30) {
  const since = new Date(Date.now() - sinceDays * 24 * 60 * 60 * 1000);

  // Combine granular checks (recent) with daily aggregates (older) for accurate stats
  const recentResult = await db
    .select({
      totalChecks: sql<number>`count(*)::int`,
      upChecks: sql<number>`count(*) filter (where ${shopUptime.isUp} = true)::int`,
      avgTtfb: sql<number>`round(avg(${shopUptime.ttfb}) filter (where ${shopUptime.ttfb} is not null))::int`,
    })
    .from(shopUptime)
    .where(
      sql`${shopUptime.shopId} = ${shopId} and ${shopUptime.checkedAt} >= ${since}`,
    );

  const dailyResult = await db
    .select({
      totalChecks: sql<number>`coalesce(sum(${shopUptimeDaily.totalChecks}), 0)::int`,
      upChecks: sql<number>`coalesce(sum(${shopUptimeDaily.upChecks}), 0)::int`,
      avgTtfb: sql<number>`round(avg(${shopUptimeDaily.avgTtfb}) filter (where ${shopUptimeDaily.avgTtfb} is not null))::int`,
    })
    .from(shopUptimeDaily)
    .where(
      sql`${shopUptimeDaily.shopId} = ${shopId} and ${shopUptimeDaily.date} >= ${since}`,
    );

  const recent = recentResult[0];
  const daily = dailyResult[0];
  const totalChecks = (recent?.totalChecks ?? 0) + (daily?.totalChecks ?? 0);
  const upChecks = (recent?.upChecks ?? 0) + (daily?.upChecks ?? 0);

  // Weighted average of TTFBs
  let avgTtfb: number | null = null;
  if (recent?.avgTtfb && daily?.avgTtfb) {
    const recentWeight = recent.totalChecks;
    const dailyWeight = daily.totalChecks;
    avgTtfb = Math.round(
      (recent.avgTtfb * recentWeight + daily.avgTtfb * dailyWeight) / (recentWeight + dailyWeight),
    );
  } else {
    avgTtfb = recent?.avgTtfb ?? daily?.avgTtfb ?? null;
  }

  return {
    totalChecks,
    upChecks,
    uptimePercentage:
      totalChecks > 0 ? Number(((upChecks / totalChecks) * 100).toFixed(2)) : null,
    avgTtfb,
  };
}

export async function getShopsWithUptimeEnabled() {
  const db = getConnection();
  return db
    .select({
      id: shop.id,
      url: shop.url,
      uptimeStatus: shop.uptimeStatus,
      uptimeConsecutiveFails: shop.uptimeConsecutiveFails,
    })
    .from(shop)
    .where(eq(shop.uptimeEnabled, true));
}

export async function updateShopUptimeStatus(
  db: Drizzle,
  shopId: number,
  updates: {
    uptimeStatus: string;
    uptimeLastCheckedAt: Date;
    uptimeConsecutiveFails: number;
    uptimeDownSince?: Date | null;
  },
) {
  await db.update(shop).set(updates).where(eq(shop.id, shopId)).execute();
}

/**
 * Aggregate yesterday's granular checks into daily summaries, then delete the granular rows.
 * Called daily by the maintenance worker.
 */
export async function aggregateAndCleanup() {
  const db = getConnection();

  // Aggregate all granular checks older than 7 days into daily rows
  const cutoff = new Date(Date.now() - 7 * 24 * 60 * 60 * 1000);
  cutoff.setHours(0, 0, 0, 0);

  await db.execute(sql`
    INSERT INTO shop_uptime_daily (shop_id, date, total_checks, up_checks, avg_ttfb, min_ttfb, max_ttfb, avg_response_time)
    SELECT
      shop_id,
      date_trunc('day', checked_at) AS date,
      count(*)::int AS total_checks,
      count(*) FILTER (WHERE is_up = true)::int AS up_checks,
      round(avg(ttfb) FILTER (WHERE ttfb IS NOT NULL))::int AS avg_ttfb,
      min(ttfb) FILTER (WHERE ttfb IS NOT NULL) AS min_ttfb,
      max(ttfb) FILTER (WHERE ttfb IS NOT NULL) AS max_ttfb,
      round(avg(response_time) FILTER (WHERE response_time IS NOT NULL))::int AS avg_response_time
    FROM shop_uptime
    WHERE checked_at < ${cutoff}
    GROUP BY shop_id, date_trunc('day', checked_at)
    ON CONFLICT (shop_id, date) DO UPDATE SET
      total_checks = EXCLUDED.total_checks,
      up_checks = EXCLUDED.up_checks,
      avg_ttfb = EXCLUDED.avg_ttfb,
      min_ttfb = EXCLUDED.min_ttfb,
      max_ttfb = EXCLUDED.max_ttfb,
      avg_response_time = EXCLUDED.avg_response_time
  `);

  // Delete aggregated granular rows
  await db.delete(shopUptime).where(lt(shopUptime.checkedAt, cutoff)).execute();

  // Keep daily summaries for 1 year
  const yearAgo = new Date(Date.now() - 365 * 24 * 60 * 60 * 1000);
  await db.delete(shopUptimeDaily).where(lt(shopUptimeDaily.date, yearAgo)).execute();
}
