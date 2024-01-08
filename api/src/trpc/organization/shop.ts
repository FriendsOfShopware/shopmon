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
import { HttpClient, HttpClientResponse, SimpleShop } from '@friendsofshopware/app-server-sdk';

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
                    created_at: true,
                    last_scraped_at: true,
                    status: true,
                    last_scraped_error: true,
                    shopware_version: true,
                },
                where: eq(schema.shop.team_id, ctx.user),
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
                    created_at: schema.shop.created_at,
                    shopware_version: schema.shop.shopware_version,
                    last_scraped_at: schema.shop.last_scraped_at,
                    last_scraped_error: schema.shop.last_scraped_error,
                    last_updated: schema.shop.last_updated,
                    ignores: schema.shop.ignores,
                    shop_image: schema.shop.shop_image,
                    extensions: schema.shopScrapeInfo.extensions,
                    scheduled_task: schema.shopScrapeInfo.scheduled_task,
                    queue_info: schema.shopScrapeInfo.queue_info,
                    cache_info: schema.shopScrapeInfo.cache_info,
                    checks: schema.shopScrapeInfo.checks,
                    team_id: schema.shop.team_id,
                    team_name: schema.team.name,
                })
                .from(schema.shop)
                .innerJoin(schema.team, eq(schema.team.id, schema.shop.team_id))
                .leftJoin(
                    schema.shopScrapeInfo,
                    eq(schema.shopScrapeInfo.shop, schema.shop.id),
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
                where: eq(schema.shopPageSpeed.shop_id, input.shopId),
                orderBy: [desc(schema.shopPageSpeed.created_at)],
            });

            const shopChangelog =
                await ctx.drizzle.query.shopChangelog.findMany({
                    where: eq(schema.shopChangelog.shop_id, input.shopId),
                    orderBy: [desc(schema.shopChangelog.date)],
                });

            for (const row of shopChangelog) {
                row.extensions = JSON.parse(row.extensions);
            }

            shop.extensions = JSON.parse(shop.extensions || '{}');
            shop.scheduled_task = JSON.parse(shop.scheduled_task || '{}');
            shop.queue_info = JSON.parse(shop.queue_info || '{}');
            shop.cache_info = JSON.parse(shop.cache_info || '{}');
            shop.checks = JSON.parse(shop.checks || '{}');
            shop.ignores = JSON.parse(shop.ignores || '[]');

            type AdditionalShopInfo = {
                pagespeed: typeof pageSpeed;
                changelog: typeof shopChangelog;
            };

            const result: typeof shop & AdditionalShopInfo = {
                ...shop,
                pagespeed: pageSpeed,
                changelog: shopChangelog,
            };

            return result;
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
                return new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Cannot reach shop',
                });
            }

            const clientSecret = await encrypt(
                ctx.env.APP_SECRET,
                input.clientSecret,
            );

            const id = await Shops.createShop(ctx.drizzle, {
                team_id: input.orgId,
                name: input.name,
                client_id: input.clientId,
                client_secret: clientSecret,
                shop_url: input.shopUrl,
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
                    .set({ ignores: JSON.stringify(input.ignores) })
                    .where(eq(schema.shop.id, input.shopId))
                    .execute();
            }

            if (input.newOrgId && input.newOrgId !== input.orgId) {
                const team = await ctx.drizzle.query.team.findFirst({
                    columns: {
                        id: true,
                        owner_id: true,
                    },
                    where: and(
                        eq(schema.team.id, input.newOrgId),
                        eq(schema.team.owner_id, ctx.user),
                    ),
                });

                if (team === undefined) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'You are not a member of this team',
                    });
                }

                if (team.owner_id !== ctx.user) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'You are not the owner of this team',
                    });
                }

                await ctx.drizzle
                    .update(schema.shop)
                    .set({ team_id: input.newOrgId })
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
                    return new TRPCError({
                        code: 'BAD_REQUEST',
                        message: 'Cannot reach shop',
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
                        client_id: input.clientId,
                        client_secret: clientSecret,
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
                    client_id: true,
                    client_secret: true,
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
                shopData.client_secret,
            );
            const shop = new SimpleShop('', shopData.url, '');
            shop.setShopCredentials(shopData.client_id, clientSecret);
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
                    client_id: true,
                    client_secret: true,
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
                shopData.client_secret,
            );
            const shop = new SimpleShop('', shopData.url, '');
            shop.setShopCredentials(shopData.client_id, clientSecret);
            const client = new HttpClient(shop);

            const nextExecutionTime: string = new Date().toISOString();
            await client.patch(`/scheduled-task/${input.taskId}`, {
                status: 'scheduled',
                nextExecutionTime: nextExecutionTime,
            });

            const scrapeResult =
                await ctx.drizzle.query.shopScrapeInfo.findFirst({
                    columns: {
                        scheduled_task: true,
                    },
                    where: eq(schema.shopScrapeInfo.shop, input.shopId),
                });

            // If there is no scrape result, we don't need to update the scheduled task
            if (scrapeResult === undefined) {
                return true;
            }

            const scheduledTasks = JSON.parse(
                scrapeResult.scheduled_task || '{}',
            );

            for (const task of scheduledTasks) {
                if (task.id === input.taskId) {
                    task.status = 'scheduled';
                    task.nextExecutionTime = nextExecutionTime;
                    task.overdue = false;
                }
            }

            await ctx.drizzle
                .update(schema.shopScrapeInfo)
                .set({ scheduled_task: JSON.stringify(scheduledTasks) })
                .where(eq(schema.shopScrapeInfo.shop, input.shopId))
                .execute();
        }),
});
