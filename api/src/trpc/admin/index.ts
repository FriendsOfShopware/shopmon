import { count, desc, eq, ilike, sql } from 'drizzle-orm';
import { z } from 'zod';
import { schema } from '#src/db';
import { publicProcedure, router } from '#src/trpc/index.ts';
import { isAdminMiddleware } from '#src/trpc/middleware.ts';

const organizationsPaginationSchema = z.object({
    limit: z.number().min(1).max(100).default(20),
    offset: z.number().min(0).default(0),
    sortBy: z
        .enum(['name', 'slug', 'createdAt', 'shopCount', 'memberCount'])
        .default('createdAt'),
    sortDirection: z.enum(['asc', 'desc']).default('desc'),
    searchField: z.enum(['name', 'slug']).optional(),
    searchOperator: z.enum(['contains']).optional(),
    searchValue: z.string().optional(),
    filterField: z.enum(['name', 'slug']).optional(),
    filterOperator: z.enum(['eq']).optional(),
    filterValue: z.string().optional(),
});

const shopsPaginationSchema = z.object({
    limit: z.number().min(1).max(100).default(20),
    offset: z.number().min(0).default(0),
    sortBy: z
        .enum([
            'name',
            'url',
            'status',
            'shopwareVersion',
            'lastScrapedAt',
            'organizationName',
            'organizationSlug',
            'createdAt',
        ])
        .default('name'),
    sortDirection: z.enum(['asc', 'desc']).default('desc'),
    searchField: z.enum(['name', 'url']).optional(),
    searchOperator: z.enum(['contains']).optional(),
    searchValue: z.string().optional(),
    filterField: z.enum(['name', 'url', 'status']).optional(),
    filterOperator: z.enum(['eq']).optional(),
    filterValue: z.string().optional(),
});

export const adminRouter = router({
    listOrganizations: publicProcedure
        .use(isAdminMiddleware)
        .input(organizationsPaginationSchema.optional())
        .query(async ({ ctx, input }) => {
            if (!input) {
                return ctx.drizzle
                    .select({
                        id: schema.organization.id,
                        name: schema.organization.name,
                        slug: schema.organization.slug,
                        logo: schema.organization.logo,
                        createdAt: schema.organization.createdAt,
                        shopCount: sql<number>`(SELECT COUNT(*) FROM shop WHERE organization_id = ${schema.organization.id})`,
                        memberCount: sql<number>`(SELECT COUNT(*) FROM member WHERE organization_id = ${schema.organization.id})`,
                    })
                    .from(schema.organization)
                    .orderBy(desc(schema.organization.createdAt))
                    .all();
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
            } = input;

            // Build where conditions
            const conditions: Array<
                ReturnType<typeof eq> | ReturnType<typeof ilike>
            > = [];

            if (searchField && searchValue && searchOperator === 'contains') {
                if (searchField === 'name') {
                    conditions.push(
                        ilike(schema.organization.name, `%${searchValue}%`),
                    );
                } else if (searchField === 'slug') {
                    conditions.push(
                        ilike(schema.organization.slug, `%${searchValue}%`),
                    );
                }
            }

            if (filterField && filterValue && filterOperator === 'eq') {
                if (filterField === 'name') {
                    conditions.push(eq(schema.organization.name, filterValue));
                } else if (filterField === 'slug') {
                    conditions.push(eq(schema.organization.slug, filterValue));
                }
            }

            // Build order by
            // biome-ignore lint/suspicious/noExplicitAny: Complex union type for Drizzle orderBy expressions
            let orderBy: any;
            if (sortBy === 'name') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.organization.name)
                        : schema.organization.name;
            } else if (sortBy === 'slug') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.organization.slug)
                        : schema.organization.slug;
            } else if (sortBy === 'createdAt') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.organization.createdAt)
                        : schema.organization.createdAt;
            } else if (sortBy === 'shopCount') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(sql`shopCount`)
                        : sql`shopCount`;
            } else if (sortBy === 'memberCount') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(sql`memberCount`)
                        : sql`memberCount`;
            } else {
                orderBy = desc(schema.organization.createdAt);
            }

            // Execute queries
            const [organizations, totalResult] = await Promise.all([
                ctx.drizzle
                    .select({
                        id: schema.organization.id,
                        name: schema.organization.name,
                        slug: schema.organization.slug,
                        logo: schema.organization.logo,
                        createdAt: schema.organization.createdAt,
                        shopCount: sql<number>`(SELECT COUNT(*) FROM shop WHERE organization_id = ${schema.organization.id})`,
                        memberCount: sql<number>`(SELECT COUNT(*) FROM member WHERE organization_id = ${schema.organization.id})`,
                    })
                    .from(schema.organization)
                    .where(
                        conditions.length > 0
                            ? conditions.reduce(
                                  (acc, condition) => acc && condition,
                              )
                            : undefined,
                    )
                    .orderBy(orderBy)
                    .limit(limit)
                    .offset(offset),
                ctx.drizzle
                    .select({ count: count() })
                    .from(schema.organization)
                    .where(
                        conditions.length > 0
                            ? conditions.reduce(
                                  (acc, condition) => acc && condition,
                              )
                            : undefined,
                    )
                    .get(),
            ]);

            return {
                organizations,
                total: totalResult?.count || 0,
            };
        }),
    listShops: publicProcedure
        .use(isAdminMiddleware)
        .input(shopsPaginationSchema.optional())
        .query(async ({ ctx, input }) => {
            // If no input provided, return all shops without pagination
            if (!input) {
                return ctx.drizzle
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
                    .innerJoin(
                        schema.organization,
                        eq(schema.shop.organizationId, schema.organization.id),
                    )
                    .orderBy(desc(schema.shop.createdAt))
                    .all();
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
            } = input;

            // Build where conditions
            const conditions: Array<
                ReturnType<typeof eq> | ReturnType<typeof ilike>
            > = [];

            if (searchField && searchValue && searchOperator === 'contains') {
                if (searchField === 'name') {
                    conditions.push(
                        ilike(schema.shop.name, `%${searchValue}%`),
                    );
                } else if (searchField === 'url') {
                    conditions.push(ilike(schema.shop.url, `%${searchValue}%`));
                }
            }

            if (filterField && filterValue && filterOperator === 'eq') {
                if (filterField === 'name') {
                    conditions.push(eq(schema.shop.name, filterValue));
                } else if (filterField === 'url') {
                    conditions.push(eq(schema.shop.url, filterValue));
                } else if (filterField === 'status') {
                    conditions.push(eq(schema.shop.status, filterValue));
                }
            }

            // Build order by
            // biome-ignore lint/suspicious/noExplicitAny: Complex union type for Drizzle orderBy expressions
            let orderBy: any;
            if (sortBy === 'name') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.name)
                        : schema.shop.name;
            } else if (sortBy === 'url') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.url)
                        : schema.shop.url;
            } else if (sortBy === 'status') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.status)
                        : schema.shop.status;
            } else if (sortBy === 'shopwareVersion') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.shopwareVersion)
                        : schema.shop.shopwareVersion;
            } else if (sortBy === 'lastScrapedAt') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.lastScrapedAt)
                        : schema.shop.lastScrapedAt;
            } else if (sortBy === 'organizationName') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(sql`organizationName`)
                        : sql`organizationName`;
            } else if (sortBy === 'organizationSlug') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(sql`organizationSlug`)
                        : sql`organizationSlug`;
            } else if (sortBy === 'createdAt') {
                orderBy =
                    sortDirection === 'desc'
                        ? desc(schema.shop.createdAt)
                        : schema.shop.createdAt;
            } else {
                orderBy = desc(schema.shop.createdAt);
            }

            // Execute queries
            const [shops, totalResult] = await Promise.all([
                ctx.drizzle
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
                    .innerJoin(
                        schema.organization,
                        eq(schema.shop.organizationId, schema.organization.id),
                    )
                    .where(
                        conditions.length > 0
                            ? conditions.reduce(
                                  (acc, condition) => acc && condition,
                              )
                            : undefined,
                    )
                    .orderBy(orderBy)
                    .limit(limit)
                    .offset(offset),
                ctx.drizzle
                    .select({ count: count() })
                    .from(schema.shop)
                    .innerJoin(
                        schema.organization,
                        eq(schema.shop.organizationId, schema.organization.id),
                    )
                    .where(
                        conditions.length > 0
                            ? conditions.reduce(
                                  (acc, condition) => acc && condition,
                              )
                            : undefined,
                    )
                    .get(),
            ]);

            return {
                shops,
                total: totalResult?.count || 0,
            };
        }),
    getStats: publicProcedure.use(isAdminMiddleware).query(async ({ ctx }) => {
        const [usersResult, organizationsResult, shopsStatsResult] =
            await Promise.all([
                ctx.drizzle.select({ count: count() }).from(schema.user).get(),
                ctx.drizzle
                    .select({ count: count() })
                    .from(schema.organization)
                    .get(),
                ctx.drizzle
                    .select({
                        total: count().as('total'),
                        greenCount: count(
                            sql`CASE WHEN ${schema.shop.status} = 'green' THEN 1 END`,
                        ).as('greenCount'),
                        yellowCount: count(
                            sql`CASE WHEN ${schema.shop.status} = 'yellow' THEN 1 END`,
                        ).as('yellowCount'),
                        redCount: count(
                            sql`CASE WHEN ${schema.shop.status} = 'red' THEN 1 END`,
                        ).as('redCount'),
                    })
                    .from(schema.shop)
                    .get(),
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
    }),
});
