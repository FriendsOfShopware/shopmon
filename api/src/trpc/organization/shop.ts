import { router, publicProcedure } from '..';
import { z } from 'zod';
import {
    loggedInUserMiddleware,
    organizationAdminMiddleware,
    organizationMiddleware,
    shopMiddleware,
} from '../middleware';
import { and, eq, desc } from 'drizzle-orm';
import { schema } from '../../db';
import { TRPCError } from '@trpc/server';
import Shops from '../../repository/shops';
import { decrypt, encrypt } from '../../crypto';
import {
    HttpClient,
    HttpClientResponse,
    SimpleShop,
} from '@friendsofshopware/app-server-sdk';

export const shopRouter = router({
    list: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle.query.shop.findMany({
                columns: {
                    id: true,
                    name: true,
                    url: true,
                    favicon: true,
                    createdAt: true,
                    lastScrapedAt: true,
                    status: true,
                    lastScrapedError: true,
                    shopwareVersion: true,
                },
                where: eq(schema.shop.organizationId, ctx.user),
            });

            return result === undefined ? [] : result;
        }),
    get: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(shopMiddleware)
        .query(async ({ input, ctx }) => {
            const shop = await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    url: schema.shop.url,
                    status: schema.shop.status,
                    createdAt: schema.shop.createdAt,
                    shopware_version: schema.shop.shopwareVersion,
                    lastScrapedAt: schema.shop.lastScrapedAt,
                    lastScrapedError: schema.shop.lastScrapedError,
                    lastUpdated: schema.shop.lastUpdated,
                    ignores: schema.shop.ignores,
                    shopImage: schema.shop.shopImage,
                    extensions: schema.shopScrapeInfo.extensions,
                    scheduledTask: schema.shopScrapeInfo.scheduledTask,
                    queueInfo: schema.shopScrapeInfo.queueInfo,
                    cacheInfo: schema.shopScrapeInfo.cacheInfo,
                    checks: schema.shopScrapeInfo.checks,
                    organizationId: schema.shop.organizationId,
                    organizationName: schema.organization.name,
                })
                .from(schema.shop)
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .leftJoin(
                    schema.shopScrapeInfo,
                    eq(schema.shopScrapeInfo.shopId, schema.shop.id),
                )
                .where(eq(schema.shop.id, input.shopId))
                .get();

            if (shop === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Not Found.',
                });
            }

            const pageSpeed = await ctx.drizzle.query.shopPageSpeed.findMany({
                where: eq(schema.shopPageSpeed.shopId, input.shopId),
                orderBy: [desc(schema.shopPageSpeed.createdAt)],
            });

            const shopChangelog =
                await ctx.drizzle.query.shopChangelog.findMany({
                    where: eq(schema.shopChangelog.shopId, input.shopId),
                    orderBy: [desc(schema.shopChangelog.date)],
                });

            return { ...shop, pageSpeed: pageSpeed, changelog: shopChangelog };
        }),
    create: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                name: z.string(),
                shopUrl: z.string().url(),
                clientId: z.string(),
                clientSecret: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .mutation(async ({ input, ctx }) => {
            const shop = new SimpleShop('', input.shopUrl, '');
            shop.setShopCredentials(input.clientId, input.clientSecret);

            const client = new HttpClient(shop);

            let resp: HttpClientResponse;
            try {
                resp = await client.get('/_info/config');
            } catch (e) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Cannot reach shop. Check your credentials and shop URL.',
                });
            }

            const clientSecret = await encrypt(
                ctx.env.APP_SECRET,
                input.clientSecret,
            );

            const id = await Shops.createShop(ctx.drizzle, {
                organizationId: input.orgId,
                name: input.name,
                clientId: input.clientId,
                clientSecret: clientSecret,
                shopUrl: input.shopUrl,
                version: resp.body.version,
            });

            const scrapeObject = ctx.env.SHOPS_SCRAPE.get(
                ctx.env.SHOPS_SCRAPE.idFromName(id.toString()),
            );

            await scrapeObject.fetch(
                `http://localhost/now?id=${id.toString()}`,
            );

            const pagespeedObject = ctx.env.PAGESPEED_SCRAPE.get(
                ctx.env.PAGESPEED_SCRAPE.idFromName(id.toString()),
            );

            await pagespeedObject.fetch(
                `http://localhost/now?id=${id.toString()}`,
            );

            return id;
        }),
    delete: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Shops.deleteShop(ctx.drizzle, input.shopId);

            const scrapeObject = ctx.env.SHOPS_SCRAPE.get(
                ctx.env.SHOPS_SCRAPE.idFromName(input.shopId.toString()),
            );

            await scrapeObject.fetch(
                `http://localhost/delete?id=${input.shopId.toString()}`,
            );

            const pageSpeedObject = ctx.env.PAGESPEED_SCRAPE.get(
                ctx.env.PAGESPEED_SCRAPE.idFromName(input.shopId.toString()),
            );

            await pageSpeedObject.fetch(
                `http://localhost/delete?id=${input.shopId.toString()}`,
            );

            return true;
        }),
    update: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
                name: z.string().optional(),
                shopUrl: z.string().url().optional(),
                clientId: z.string().optional(),
                clientSecret: z.string().optional(),
                ignores: z.array(z.string()).optional(),
                newOrgId: z.number().optional(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            if (input.name) {
                await ctx.drizzle
                    .update(schema.shop)
                    .set({ name: input.name })
                    .where(eq(schema.shop.id, input.shopId))
                    .execute();
            }

            if (input.ignores) {
                await ctx.drizzle
                    .update(schema.shop)
                    .set({ ignores: input.ignores })
                    .where(eq(schema.shop.id, input.shopId))
                    .execute();
            }

            if (input.newOrgId && input.newOrgId !== input.orgId) {
                const organization = await ctx.drizzle.query.organization.findFirst({
                    columns: {
                        id: true,
                        ownerId: true,
                    },
                    where: and(
                        eq(schema.organization.id, input.newOrgId),
                        eq(schema.organization.ownerId, ctx.user),
                    ),
                });

                if (organization === undefined) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'You are not a member of this organization',
                    });
                }

                if (organization.ownerId !== ctx.user) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'You are not the owner of this organization',
                    });
                }

                await ctx.drizzle
                    .update(schema.shop)
                    .set({ organizationId: input.newOrgId })
                    .where(eq(schema.shop.id, input.shopId))
                    .execute();
            }

            if (input.shopUrl && input.clientId && input.clientSecret) {
                // Try out the new credentials
                const shop = new SimpleShop('', input.shopUrl, '');
                shop.setShopCredentials(input.clientId, input.clientSecret);

                const client = new HttpClient(shop);

                try {
                    await client.get('/_info/config');
                } catch (e) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'Cannot reach shop. Check your credentials and shop URL.',
                    });
                }

                const clientSecret = await encrypt(
                    ctx.env.APP_SECRET,
                    input.clientSecret,
                );

                await ctx.drizzle
                    .update(schema.shop)
                    .set({
                        url: input.shopUrl,
                        clientId: input.clientId,
                        clientSecret: clientSecret,
                    })
                    .where(eq(schema.shop.id, input.shopId))
                    .execute();
            }
        }),
    refreshShop: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
                pageSpeed: z.boolean(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            const obj = ctx.env.SHOPS_SCRAPE.get(
                ctx.env.SHOPS_SCRAPE.idFromName(input.shopId.toString()),
            );

            await obj.fetch(
                `http://localhost/now?id=${input.shopId.toString()}&userId=${ctx.user?.toString()}`,
            );

            if (input.pageSpeed) {
                const pagespeed = ctx.env.PAGESPEED_SCRAPE.get(
                    ctx.env.PAGESPEED_SCRAPE.idFromName(
                        input.shopId.toString(),
                    ),
                );

                await pagespeed.fetch(
                    `http://localhost/now?id=${input.shopId.toString()}`,
                );
            }

            return true;
        }),
    clearShopCache: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            const shopData = await ctx.drizzle.query.shop.findFirst({
                columns: {
                    url: true,
                    clientId: true,
                    clientSecret: true,
                },
                where: eq(schema.shop.id, input.shopId),
            });

            if (shopData === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Not Found.',
                });
            }

            const clientSecret = await decrypt(
                ctx.env.APP_SECRET,
                shopData.clientSecret,
            );
            const shop = new SimpleShop('', shopData.url, '');
            shop.setShopCredentials(shopData.clientId, clientSecret);
            const client = new HttpClient(shop);

            await client.delete('/_action/cache');

            return true;
        }),
    rescheduleTask: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                shopId: z.number(),
                taskId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            const shopData = await ctx.drizzle.query.shop.findFirst({
                columns: {
                    url: true,
                    clientId: true,
                    clientSecret: true,
                },
                where: eq(schema.shop.id, input.shopId),
            });

            if (shopData === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Not Found.',
                });
            }

            const clientSecret = await decrypt(
                ctx.env.APP_SECRET,
                shopData.clientSecret,
            );
            const shop = new SimpleShop('', shopData.url, '');
            shop.setShopCredentials(shopData.clientId, clientSecret);
            const client = new HttpClient(shop);

            const nextExecutionTime: string = new Date().toISOString();
            await client.patch(`/scheduled-task/${input.taskId}`, {
                status: 'scheduled',
                nextExecutionTime: nextExecutionTime,
            });

            const scrapeResult =
                await ctx.drizzle.query.shopScrapeInfo.findFirst({
                    columns: {
                        scheduledTask: true,
                    },
                    where: eq(schema.shopScrapeInfo.shopId, input.shopId),
                });

            // If there is no scrape result, we don't need to update the scheduled task
            if (scrapeResult === undefined) {
                return true;
            }

            for (const task of scrapeResult.scheduledTask) {
                if (task.id === input.taskId) {
                    task.status = 'scheduled';
                    task.nextExecutionTime = nextExecutionTime;
                    task.overdue = false;
                }
            }

            await ctx.drizzle
                .update(schema.shopScrapeInfo)
                .set({ scheduledTask: scrapeResult.scheduledTask })
                .where(eq(schema.shopScrapeInfo.shopId, input.shopId))
                .execute();
        }),
});
