import { count, desc, eq, ilike, sql, asc } from "drizzle-orm";
import { type Drizzle, schema } from "#src/db.ts";

export interface OrganizationListParams {
  limit: number;
  offset: number;
  sortBy: "name" | "slug" | "createdAt" | "shopCount" | "memberCount";
  sortDirection: "asc" | "desc";
  searchField?: "name" | "slug";
  searchOperator?: "contains";
  searchValue?: string;
  filterField?: "name" | "slug";
  filterOperator?: "eq";
  filterValue?: string;
}

export interface ShopListParams {
  limit: number;
  offset: number;
  sortBy:
    | "name"
    | "url"
    | "status"
    | "shopwareVersion"
    | "lastScrapedAt"
    | "organizationName"
    | "organizationSlug"
    | "createdAt";
  sortDirection: "asc" | "desc";
  searchField?: "name" | "url";
  searchOperator?: "contains";
  searchValue?: string;
  filterField?: "name" | "url" | "status";
  filterOperator?: "eq";
  filterValue?: string;
}

async function listOrganizations(con: Drizzle, params?: OrganizationListParams) {
  if (!params) {
    const organizations = await con
      .select({
        id: schema.organization.id,
        name: schema.organization.name,
        slug: schema.organization.slug,
        logo: schema.organization.logo,
        createdAt: schema.organization.createdAt,
        shopCount: sql<number>`(SELECT COUNT(*) FROM shop WHERE shop.organization_id = "organization"."id")`,
        memberCount: sql<number>`(SELECT COUNT(*) FROM member WHERE member.organization_id = "organization"."id")`,
      })
      .from(schema.organization)
      .orderBy(desc(schema.organization.createdAt));
    return { organizations, total: organizations.length };
  }

  const {
    limit,
    offset,
    sortBy,
    sortDirection,
    searchField,
    searchOperator,
    searchValue,
    filterField,
    filterOperator,
    filterValue,
  } = params;

  // Build where conditions
  const conditions: Array<ReturnType<typeof eq> | ReturnType<typeof ilike>> = [];

  if (searchField && searchValue && searchOperator === "contains") {
    if (searchField === "name") {
      conditions.push(ilike(schema.organization.name, `%${searchValue}%`));
    } else if (searchField === "slug") {
      conditions.push(ilike(schema.organization.slug, `%${searchValue}%`));
    }
  }

  if (filterField && filterValue && filterOperator === "eq") {
    if (filterField === "name") {
      conditions.push(eq(schema.organization.name, filterValue));
    } else if (filterField === "slug") {
      conditions.push(eq(schema.organization.slug, filterValue));
    }
  }

  // Build order by
  // eslint-disable-next-line @typescript-eslint/no-explicit-any Complex union type for Drizzle orderBy expressions
  let orderBy: any;
  if (sortBy === "name") {
    orderBy = sortDirection === "desc" ? desc(schema.organization.name) : schema.organization.name;
  } else if (sortBy === "slug") {
    orderBy = sortDirection === "desc" ? desc(schema.organization.slug) : schema.organization.slug;
  } else if (sortBy === "createdAt") {
    orderBy =
      sortDirection === "desc"
        ? desc(schema.organization.createdAt)
        : schema.organization.createdAt;
  } else if (sortBy === "shopCount") {
    orderBy = sortDirection === "desc" ? desc(sql`shopCount`) : sql`shopCount`;
  } else if (sortBy === "memberCount") {
    orderBy = sortDirection === "desc" ? desc(sql`memberCount`) : sql`memberCount`;
  } else {
    orderBy = desc(schema.organization.createdAt);
  }

  // Execute queries
  const [organizations, totalResult] = await Promise.all([
    con
      .select({
        id: schema.organization.id,
        name: schema.organization.name,
        slug: schema.organization.slug,
        logo: schema.organization.logo,
        createdAt: schema.organization.createdAt,
        shopCount: sql<number>`(SELECT COUNT(*) FROM shop WHERE shop.organization_id = "organization"."id")`,
        memberCount: sql<number>`(SELECT COUNT(*) FROM member WHERE member.organization_id = "organization"."id")`,
      })
      .from(schema.organization)
      .where(
        conditions.length > 0 ? conditions.reduce((acc, condition) => acc && condition) : undefined,
      )
      .orderBy(orderBy)
      .limit(limit)
      .offset(offset),
    con
      .select({ count: count() })
      .from(schema.organization)
      .where(
        conditions.length > 0 ? conditions.reduce((acc, condition) => acc && condition) : undefined,
      )
      .then((rows) => rows[0]),
  ]);

  return {
    organizations,
    total: totalResult?.count || 0,
  };
}

async function listShops(con: Drizzle, params?: ShopListParams) {
  // If no input provided, return all shops without pagination
  if (!params) {
    const shops = await con
      .select({
        id: schema.shop.id,
        name: schema.shop.name,
        url: schema.shop.url,
        status: schema.shop.status,
        shopwareVersion: schema.shop.shopwareVersion,
        lastScrapedAt: schema.shop.lastScrapedAt,
        organizationId: schema.shop.organizationId,
        organizationName: sql<string>`${schema.organization.name}`,
        organizationSlug: sql<string>`${schema.organization.slug}`,
      })
      .from(schema.shop)
      .innerJoin(schema.organization, eq(schema.shop.organizationId, schema.organization.id))
      .orderBy(desc(schema.shop.createdAt));
    return { shops, total: shops.length };
  }

  const {
    limit,
    offset,
    sortBy,
    sortDirection,
    searchField,
    searchOperator,
    searchValue,
    filterField,
    filterOperator,
    filterValue,
  } = params;

  // Build where conditions
  const conditions: Array<ReturnType<typeof eq> | ReturnType<typeof ilike>> = [];

  if (searchField && searchValue && searchOperator === "contains") {
    if (searchField === "name") {
      conditions.push(ilike(schema.shop.name, `%${searchValue}%`));
    } else if (searchField === "url") {
      conditions.push(ilike(schema.shop.url, `%${searchValue}%`));
    }
  }

  if (filterField && filterValue && filterOperator === "eq") {
    if (filterField === "name") {
      conditions.push(eq(schema.shop.name, filterValue));
    } else if (filterField === "url") {
      conditions.push(eq(schema.shop.url, filterValue));
    } else if (filterField === "status") {
      conditions.push(eq(schema.shop.status, filterValue));
    }
  }

  // Build order by
  // eslint-disable-next-line @typescript-eslint/no-explicit-any Complex union type for Drizzle orderBy expressions
  let orderBy: any;
  if (sortBy === "name") {
    orderBy = sortDirection === "desc" ? desc(schema.shop.name) : schema.shop.name;
  } else if (sortBy === "url") {
    orderBy = sortDirection === "desc" ? desc(schema.shop.url) : schema.shop.url;
  } else if (sortBy === "status") {
    orderBy = sortDirection === "desc" ? desc(schema.shop.status) : schema.shop.status;
  } else if (sortBy === "shopwareVersion") {
    orderBy =
      sortDirection === "desc" ? desc(schema.shop.shopwareVersion) : schema.shop.shopwareVersion;
  } else if (sortBy === "lastScrapedAt") {
    orderBy =
      sortDirection === "desc" ? desc(schema.shop.lastScrapedAt) : schema.shop.lastScrapedAt;
  } else if (sortBy === "organizationName") {
    orderBy = sortDirection === "desc" ? desc(sql`organizationName`) : sql`organizationName`;
  } else if (sortBy === "organizationSlug") {
    orderBy = sortDirection === "desc" ? desc(sql`organizationSlug`) : sql`organizationSlug`;
  } else if (sortBy === "createdAt") {
    orderBy = sortDirection === "desc" ? desc(schema.shop.createdAt) : schema.shop.createdAt;
  } else {
    orderBy = desc(schema.shop.createdAt);
  }

  // Execute queries
  const [shops, totalResult] = await Promise.all([
    con
      .select({
        id: schema.shop.id,
        name: schema.shop.name,
        url: schema.shop.url,
        status: schema.shop.status,
        shopwareVersion: schema.shop.shopwareVersion,
        lastScrapedAt: schema.shop.lastScrapedAt,
        organizationId: schema.shop.organizationId,
        organizationName: sql<string>`${schema.organization.name}`,
        organizationSlug: sql<string>`${schema.organization.slug}`,
      })
      .from(schema.shop)
      .innerJoin(schema.organization, eq(schema.shop.organizationId, schema.organization.id))
      .where(
        conditions.length > 0 ? conditions.reduce((acc, condition) => acc && condition) : undefined,
      )
      .orderBy(orderBy)
      .limit(limit)
      .offset(offset),
    con
      .select({ count: count() })
      .from(schema.shop)
      .innerJoin(schema.organization, eq(schema.shop.organizationId, schema.organization.id))
      .where(
        conditions.length > 0 ? conditions.reduce((acc, condition) => acc && condition) : undefined,
      )
      .then((rows) => rows[0]),
  ]);

  return {
    shops,
    total: totalResult?.count || 0,
  };
}

async function getStats(con: Drizzle) {
  const [usersResult, organizationsResult, shopsStatsResult] = await Promise.all([
    con
      .select({ count: count() })
      .from(schema.user)
      .then((rows) => rows[0]),
    con
      .select({ count: count() })
      .from(schema.organization)
      .then((rows) => rows[0]),
    con
      .select({
        total: count().as("total"),
        greenCount: count(sql`CASE WHEN ${schema.shop.status} = 'green' THEN 1 END`).as(
          "greenCount",
        ),
        yellowCount: count(sql`CASE WHEN ${schema.shop.status} = 'yellow' THEN 1 END`).as(
          "yellowCount",
        ),
        redCount: count(sql`CASE WHEN ${schema.shop.status} = 'red' THEN 1 END`).as("redCount"),
      })
      .from(schema.shop)
      .then((rows) => rows[0]),
  ]);

  const shopsByStatus = {
    green: shopsStatsResult?.greenCount || 0,
    yellow: shopsStatsResult?.yellowCount || 0,
    red: shopsStatsResult?.redCount || 0,
  };

  return {
    totalUsers: usersResult?.count || 0,
    totalOrganizations: organizationsResult?.count || 0,
    totalShops: shopsStatsResult?.total || 0,
    shopsByStatus,
  };
}

async function getGrowthData(con: Drizzle) {
  const [userGrowth, shopGrowth] = await Promise.all([
    con
      .select({
        month: sql<string>`to_char(${schema.user.createdAt}, 'YYYY-MM')`.as("month"),
        count: count(),
      })
      .from(schema.user)
      .groupBy(sql`to_char(${schema.user.createdAt}, 'YYYY-MM')`)
      .orderBy(asc(sql`to_char(${schema.user.createdAt}, 'YYYY-MM')`)),
    con
      .select({
        month: sql<string>`to_char(${schema.shop.createdAt}, 'YYYY-MM')`.as("month"),
        count: count(),
      })
      .from(schema.shop)
      .groupBy(sql`to_char(${schema.shop.createdAt}, 'YYYY-MM')`)
      .orderBy(asc(sql`to_char(${schema.shop.createdAt}, 'YYYY-MM')`)),
  ]);

  // Convert to cumulative counts
  let userTotal = 0;
  const cumulativeUsers = userGrowth.map((row) => {
    userTotal += row.count;
    return { month: row.month, count: userTotal };
  });

  let shopTotal = 0;
  const cumulativeShops = shopGrowth.map((row) => {
    shopTotal += row.count;
    return { month: row.month, count: shopTotal };
  });

  return { users: cumulativeUsers, shops: cumulativeShops };
}

async function getRecentActivity(con: Drizzle) {
  const [recentUsers, recentShops] = await Promise.all([
    con
      .select({
        id: schema.user.id,
        name: schema.user.name,
        email: schema.user.email,
        createdAt: schema.user.createdAt,
      })
      .from(schema.user)
      .orderBy(desc(schema.user.createdAt))
      .limit(10),
    con
      .select({
        id: schema.shop.id,
        name: schema.shop.name,
        url: schema.shop.url,
        organizationName: sql<string>`${schema.organization.name}`,
        createdAt: schema.shop.createdAt,
      })
      .from(schema.shop)
      .innerJoin(schema.organization, eq(schema.shop.organizationId, schema.organization.id))
      .orderBy(desc(schema.shop.createdAt))
      .limit(10),
  ]);

  return { recentUsers, recentShops };
}

async function getShopwareVersions(con: Drizzle) {
  return await con
    .select({
      version: schema.shop.shopwareVersion,
      count: count(),
    })
    .from(schema.shop)
    .groupBy(schema.shop.shopwareVersion)
    .orderBy(desc(count()));
}

export default {
  listOrganizations,
  listShops,
  getStats,
  getGrowthData,
  getRecentActivity,
  getShopwareVersions,
};
