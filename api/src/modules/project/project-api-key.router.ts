import { z } from 'zod';
import type { ApiKeyScope } from '#src/db.ts';
import { publicProcedure, router } from '#src/trpc/index.ts';
import {
    loggedInUserMiddleware,
    organizationMiddleware,
} from '#src/trpc/middleware.ts';
import * as ApiKeyService from './project-api-key.service.ts';

const apiKeyScopeSchema = z.enum([
    'deployments',
]) satisfies z.ZodType<ApiKeyScope>;

export const apiKeyRouter = router({
    list: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                projectId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .query(async ({ ctx, input }) => {
            return await ApiKeyService.listApiKeys(ctx.drizzle, input);
        }),

    create: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                projectId: z.number(),
                name: z.string().min(1).max(100),
                scopes: z.array(apiKeyScopeSchema).min(1),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .mutation(async ({ ctx, input }) => {
            return await ApiKeyService.createApiKey(ctx.drizzle, input);
        }),

    delete: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                projectId: z.number(),
                apiKeyId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .mutation(async ({ ctx, input }) => {
            return await ApiKeyService.deleteApiKey(ctx.drizzle, input);
        }),

    scopes: publicProcedure.query(() => {
        return ApiKeyService.getAvailableScopes();
    }),
});
