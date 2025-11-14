import { TRPCError } from '@trpc/server';
import { desc, eq } from 'drizzle-orm';
import { z } from 'zod';
import { schema } from '../../db.ts';
import { publicProcedure, router } from '../index.ts';
import { loggedInUserMiddleware, shopMiddleware } from '../middleware.ts';

export const deploymentRouter = router({
    list: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                limit: z.number().min(1).max(100).default(50),
                offset: z.number().min(0).default(0),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .query(async ({ ctx, input }) => {
            const deployments = await ctx.drizzle
                .select({
                    id: schema.deployment.id,
                    name: schema.deployment.name,
                    command: schema.deployment.command,
                    returnCode: schema.deployment.returnCode,
                    startDate: schema.deployment.startDate,
                    endDate: schema.deployment.endDate,
                    executionTime: schema.deployment.executionTime,
                    reference: schema.deployment.reference,
                    createdAt: schema.deployment.createdAt,
                })
                .from(schema.deployment)
                .where(eq(schema.deployment.shopId, input.shopId))
                .orderBy(desc(schema.deployment.createdAt))
                .limit(input.limit)
                .offset(input.offset)
                .all();

            return deployments;
        }),

    get: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                deploymentId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .query(async ({ ctx, input }) => {
            const deployment = await ctx.drizzle
                .select()
                .from(schema.deployment)
                .where(eq(schema.deployment.id, input.deploymentId))
                .get();

            if (!deployment) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'Deployment not found',
                });
            }

            // Verify deployment belongs to the shop
            if (deployment.shopId !== input.shopId) {
                throw new TRPCError({
                    code: 'FORBIDDEN',
                    message: 'Deployment does not belong to this shop',
                });
            }

            return deployment;
        }),

    listTokens: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .query(async ({ ctx, input }) => {
            const tokens = await ctx.drizzle
                .select({
                    id: schema.deploymentToken.id,
                    name: schema.deploymentToken.name,
                    createdAt: schema.deploymentToken.createdAt,
                    lastUsedAt: schema.deploymentToken.lastUsedAt,
                })
                .from(schema.deploymentToken)
                .where(eq(schema.deploymentToken.shopId, input.shopId))
                .orderBy(desc(schema.deploymentToken.createdAt))
                .all();

            return tokens;
        }),

    createToken: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                name: z.string().min(1).max(100),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ ctx, input }) => {
            // Generate a random token (using crypto)
            const token = crypto.randomUUID().replace(/-/g, '');
            const id = crypto.randomUUID();

            await ctx.drizzle.insert(schema.deploymentToken).values({
                id,
                shopId: input.shopId,
                token,
                name: input.name,
                createdAt: new Date(),
            });

            // Return token only on creation (won't be shown again)
            return {
                id,
                token,
                name: input.name,
                createdAt: new Date(),
            };
        }),

    deleteToken: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                tokenId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ ctx, input }) => {
            // Verify token belongs to shop
            const token = await ctx.drizzle
                .select()
                .from(schema.deploymentToken)
                .where(eq(schema.deploymentToken.id, input.tokenId))
                .get();

            if (!token) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'Token not found',
                });
            }

            if (token.shopId !== input.shopId) {
                throw new TRPCError({
                    code: 'FORBIDDEN',
                    message: 'Token does not belong to this shop',
                });
            }

            await ctx.drizzle
                .delete(schema.deploymentToken)
                .where(eq(schema.deploymentToken.id, input.tokenId));

            return { success: true };
        }),

    delete: publicProcedure
        .input(
            z.object({
                shopId: z.number(),
                deploymentId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(shopMiddleware)
        .mutation(async ({ ctx, input }) => {
            // Verify deployment belongs to shop
            const deployment = await ctx.drizzle
                .select()
                .from(schema.deployment)
                .where(eq(schema.deployment.id, input.deploymentId))
                .get();

            if (!deployment) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'Deployment not found',
                });
            }

            if (deployment.shopId !== input.shopId) {
                throw new TRPCError({
                    code: 'FORBIDDEN',
                    message: 'Deployment does not belong to this shop',
                });
            }

            await ctx.drizzle
                .delete(schema.deployment)
                .where(eq(schema.deployment.id, input.deploymentId));

            return { success: true };
        }),
});
