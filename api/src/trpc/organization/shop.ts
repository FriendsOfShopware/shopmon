import {
    HttpClient,
    type HttpClientResponse,
    SimpleShop,
} from '@shopware-ag/app-server-sdk';
import { TRPCError } from '@trpc/server';
import { and, desc, eq, sql } from 'drizzle-orm';
import { z } from 'zod';
import { scrapeSinglePagespeedShop } from '../../cron/jobs/pagespeedScrape.ts';
import { scrapeSingleShop } from '../../cron/jobs/shopScrape.ts';
import { decrypt, encrypt } from '../../crypto/index.ts';
import { schema } from '../../db.ts';
import Shops from '../../repository/shops.ts';
import { publicProcedure, router } from '../index.ts';
import {
    loggedInUserMiddleware,
    organizationMiddleware,
    shopMiddleware,
} from '../middleware.ts';

export const shopRouter = router({
    list: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .query(async ({ ctx, input }) => {
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
                where: eq(schema.shop.organizationId, input.orgId),
            });

            return result === undefined ? [] : result;
        }),
    get: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .query(async ({ input, ctx }) => {
            const shopQuery = ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    url: schema.shop.url,
                    status: schema.shop.status,
                    createdAt: schema.shop.createdAt,
                    shopwareVersion: schema.shop.shopwareVersion,
                    lastScrapedAt: schema.shop.lastScrapedAt,
                    lastScrapedError: schema.shop.lastScrapedError,
                    lastChangelog: schema.shop.lastChangelog,
                    ignores: schema.shop.ignores,
                    shopImage: schema.shop.shopImage,
                    extensions: schema.shopScrapeInfo.extensions,
                    scheduledTask: schema.shopScrapeInfo.scheduledTask,
                    queueInfo: schema.shopScrapeInfo.queueInfo,
                    cacheInfo: schema.shopScrapeInfo.cacheInfo,
                    checks: schema.shopScrapeInfo.checks,
                    connectionIssueCount: schema.shop.connectionIssueCount,
                    organizationId: schema.shop.organizationId,
                    organizationName:
                        sql<string>`${schema.organization.name}`.as(
                            'organization_name',
                        ),
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

            const pageSpeedQuery = ctx.drizzle.query.shopPageSpeed.findMany({
                where: eq(schema.shopPageSpeed.shopId, input.shopId),
                orderBy: [desc(schema.shopPageSpeed.createdAt)],
            });

            const shopChangelogQuery = ctx.drizzle.query.shopChangelog.findMany(
                {
                    where: eq(schema.shopChangelog.shopId, input.shopId),
                    orderBy: [desc(schema.shopChangelog.date)],
                },
            );

            const [shop, pageSpeed, shopChangelog] = await Promise.all([
                shopQuery,
                pageSpeedQuery,
                shopChangelogQuery,
            ]);

            if (shop === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Not Found.',
                });
            }

            return { ...shop, pageSpeed: pageSpeed, changelog: shopChangelog };
        }),
    create: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                name: z.string(),
                shopUrl: z.string().url(),
                clientId: z.string(),
                clientSecret: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .mutation(async ({ input, ctx }) => {
            const shop = new SimpleShop('', input.shopUrl, '');
            shop.setShopCredentials(input.clientId, input.clientSecret);

            const client = new HttpClient(shop);

            let resp: HttpClientResponse<{ version: string }>;
            try {
                resp = await client.get('/_info/config');
            } catch (e) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message:
                        'Cannot reach shop. Check your credentials and shop URL.',
                });
            }

            const clientSecret = await encrypt(
                process.env.APP_SECRET,
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

            await scrapeSingleShop(id);

            return id;
        }),
    delete: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Shops.deleteShop(ctx.drizzle, input.shopId);

            // Cleanup is handled by the delete operation
            console.log(`Shop ${input.shopId} deleted`);

            return true;
        }),
    update: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                shopId: z.number(),
                name: z.string().optional(),
                shopUrl: z.string().url().optional(),
                clientId: z.string().optional(),
                clientSecret: z.string().optional(),
                ignores: z.array(z.string()).optional(),
                newOrgId: z.string().optional(),
            }),
        )
        .use(loggedInUserMiddleware)
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
                const organization = await ctx.drizzle.query.member.findFirst({
                    columns: {
                        id: true,
                    },
                    where: and(
                        eq(schema.member.organizationId, input.newOrgId),
                        eq(schema.member.userId, ctx.user.id),
                    ),
                });

                if (organization === undefined) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'You are not a member of this organization',
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
                        message:
                            'Cannot reach shop. Check your credentials and shop URL.',
                    });
                }

                const clientSecret = await encrypt(
                    process.env.APP_SECRET,
                    input.clientSecret,
                );

                await ctx.drizzle
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
        }),
    refreshShop: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                pageSpeed: z.boolean(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ input, ctx }) => {
            await scrapeSingleShop(input.shopId);

            if (input.pageSpeed) {
                await scrapeSinglePagespeedShop(input.shopId);
            }

            return true;
        }),
    clearShopCache: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
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
                process.env.APP_SECRET,
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
                shopId: z.number(),
                taskId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
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
                process.env.APP_SECRET,
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
