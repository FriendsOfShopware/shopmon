import { and, desc, eq, inArray, sql } from 'drizzle-orm';
import { getShopScrapeInfo } from '#src/repository/scrapeInfo';
import { schema } from '../../db.ts';
import { publicProcedure, router } from '../index.ts';
import { loggedInUserMiddleware } from '../middleware.ts';
import { notificationRouter } from './notification.ts';

interface Extension {
    name: string;
    label: string;
    active: boolean;
    version: string;
    latestVersion: string | null;
    installed: boolean;
    ratingAverage: number | null;
    storeLink: string | null;
    changelog: ExtensionChangelog[] | null;
    installedAt: string | null;
}

interface ExtensionChangelog {
    version: string;
    text: string;
    creationDate: string;
    isCompatible: boolean;
}

export interface UserExtension extends Extension {
    shops: {
        [key: string]: {
            id: number;
            name: string;
            organizationId: string;
            organizationSlug: string;
            shopwareVersion: string;
            installed: boolean;
            active: boolean;
            version: string;
        };
    };
}

export const accountRouter = router({
    notification: notificationRouter,
    currentUser: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const user = await ctx.drizzle
                .select({
                    id: schema.user.id,
                    displayName: schema.user.name,
                    email: schema.user.email,
                    createdAt: schema.user.createdAt,
                })
                .from(schema.user)
                .where(eq(schema.user.id, ctx.user.id))
                .get();

            return user;
        }),
    currentUserExtensions: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    organizationId: schema.shop.organizationId,
                    shopwareVersion: schema.shop.shopwareVersion,
                    organizationSlug:
                        sql<string>`${schema.organization.slug}`.as(
                            'organization_slug',
                        ),
                })
                .from(schema.shop)
                .innerJoin(
                    schema.member,
                    eq(
                        schema.member.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .where(eq(schema.member.userId, ctx.user.id))
                .orderBy(schema.shop.name)
                .all();

            const json = {} as { [key: string]: UserExtension };

            for (const row of result) {
                const scrapeResult = await getShopScrapeInfo(row.id);

                if (!scrapeResult) {
                    continue;
                }

                for (const extension of scrapeResult.extensions) {
                    if (json[extension.name] === undefined) {
                        json[extension.name] = extension as UserExtension;
                        json[extension.name].shops = {};
                    }

                    json[extension.name].shops[row.id] = {
                        id: row.id,
                        name: row.name,
                        organizationId: row.organizationId,
                        organizationSlug: row.organizationSlug,
                        shopwareVersion: row.shopwareVersion,
                        installed: extension.installed,
                        active: extension.active,
                        version: extension.version,
                    };
                }
            }

            return Object.values(json);
        }),
    listOrganizations: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            return ctx.drizzle
                .select({
                    id: schema.organization.id,
                    name: schema.organization.name,
                    createdAt: schema.organization.createdAt,
                    slug: schema.organization.slug,
                    shopCount: sql<number>`(SELECT COUNT(1) FROM ${schema.shop} WHERE ${schema.shop.organizationId} = ${schema.organization.id})`,
                    memberCount: sql<number>`(SELECT COUNT(1) FROM ${schema.member} WHERE ${schema.member.organizationId} = ${schema.organization.id})`,
                })
                .from(schema.organization)
                .innerJoin(
                    schema.member,
                    eq(schema.member.organizationId, schema.organization.id),
                )
                .where(eq(schema.member.userId, ctx.user.id))
                .all();
        }),
    currentUserShops: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            return await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    nameCombined: sql<string>`CONCAT(${schema.project.name}, ' / ', ${schema.shop.name})`,
                    status: schema.shop.status,
                    url: schema.shop.url,
                    favicon: schema.shop.favicon,
                    createdAt: schema.shop.createdAt,
                    lastScrapedAt: schema.shop.lastScrapedAt,
                    shopwareVersion: schema.shop.shopwareVersion,
                    lastChangelog: schema.shop.lastChangelog,
                    organizationId: schema.shop.organizationId,
                    organizationName:
                        sql<string>`${schema.organization.name}`.as(
                            'organization_name',
                        ),
                    organizationSlug:
                        sql<string>`${schema.organization.slug}`.as(
                            'organization_slug',
                        ),
                    projectId: schema.shop.projectId,
                    projectName: sql<string>`${schema.project.name}`.as(
                        'project_name',
                    ),
                })
                .from(schema.shop)
                .innerJoin(
                    schema.member,
                    eq(
                        schema.member.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .leftJoin(
                    schema.project,
                    eq(schema.project.id, schema.shop.projectId),
                )
                .where(eq(schema.member.userId, ctx.user.id))
                .orderBy(schema.shop.name)
                .all();
        }),
    currentUserProjects: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            return await ctx.drizzle
                .select({
                    id: schema.project.id,
                    name: sql<string>`CONCAT(${schema.organization.name}, ' / ', ${schema.project.name})`,
                })
                .from(schema.project)
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.project.organizationId),
                )
                .innerJoin(
                    schema.member,
                    and(
                        eq(
                            schema.member.organizationId,
                            schema.organization.id,
                        ),
                        eq(schema.member.userId, ctx.user.id),
                    ),
                );
        }),
    currentUserChangelogs: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle
                .select({
                    id: schema.shopChangelog.id,
                    shopId: schema.shopChangelog.shopId,
                    shopOrganizationId: schema.shop.organizationId,
                    organizationSlug:
                        sql<string>`${schema.organization.slug}`.as(
                            'organization_slug',
                        ),
                    shopName: schema.shop.name,
                    shopFavicon: schema.shop.favicon,
                    extensions: schema.shopChangelog.extensions,
                    oldShopwareVersion: schema.shopChangelog.oldShopwareVersion,
                    newShopwareVersion: schema.shopChangelog.newShopwareVersion,
                    date: schema.shopChangelog.date,
                })
                .from(schema.shop)
                .innerJoin(
                    schema.member,
                    eq(
                        schema.member.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .innerJoin(
                    schema.shopChangelog,
                    eq(schema.shopChangelog.shopId, schema.shop.id),
                )
                .where(eq(schema.member.userId, ctx.user.id))
                .orderBy(desc(schema.shopChangelog.date))
                .limit(10)
                .all();

            return result;
        }),
    subscribedShops: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            if (ctx.user.notifications.length === 0) {
                return [];
            }

            // Extract shop IDs from notifications array
            const shopIds = ctx.user.notifications
                .filter((n) => n.startsWith('shop-'))
                .map((n) => Number.parseInt(n.replace('shop-', ''), 10))
                .filter((id) => !Number.isNaN(id));

            if (shopIds.length === 0) {
                return [];
            }

            // Get shop details for subscribed shops
            const shops = await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    url: schema.shop.url,
                    favicon: schema.shop.favicon,
                    shopwareVersion: schema.shop.shopwareVersion,
                    organizationId: schema.shop.organizationId,
                    organizationName:
                        sql<string>`${schema.organization.name}`.as(
                            'organization_name',
                        ),
                    organizationSlug:
                        sql<string>`${schema.organization.slug}`.as(
                            'organization_slug',
                        ),
                })
                .from(schema.shop)
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .innerJoin(
                    schema.member,
                    eq(
                        schema.member.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .where(
                    and(
                        inArray(schema.shop.id, shopIds),
                        eq(schema.member.userId, ctx.user.id),
                    ),
                )
                .orderBy(schema.shop.name)
                .all();

            return shops;
        }),
});
