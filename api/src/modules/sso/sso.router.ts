import { z } from 'zod';
import { publicProcedure, router } from '#src/trpc/index.ts';
import {
    hasPermissionMiddleware,
    loggedInUserMiddleware,
    organizationMiddleware,
} from '#src/trpc/middleware.ts';
import * as SSOService from './sso.service.ts';

export const ssoRouter = router({
    list: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .query(async ({ ctx, input }) => {
            return await SSOService.listProviders(ctx.drizzle, input.orgId);
        }),

    discoverOpenIdConfig: publicProcedure
        .input(
            z.object({
                issuer: z.string().url('Must be a valid URL'),
            }),
        )
        .use(loggedInUserMiddleware)
        .query(async ({ input }) => {
            return await SSOService.discoverOpenIdConfig(input.issuer);
        }),

    update: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                providerId: z.string(),
                domain: z.string(),
                issuer: z.string().url(),
                clientId: z.string(),
                clientSecret: z.string().optional(),
                authorizationEndpoint: z.string().url(),
                tokenEndpoint: z.string().url(),
                jwksEndpoint: z.string().url(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(
            hasPermissionMiddleware({
                organization: ['update'],
            }),
        )
        .mutation(async ({ ctx, input }) => {
            return await SSOService.updateProvider(ctx.drizzle, input);
        }),

    delete: publicProcedure
        .input(
            z.object({
                orgId: z.string(),
                providerId: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(
            hasPermissionMiddleware({
                organization: ['update'],
            }),
        )
        .mutation(async ({ ctx, input }) => {
            return await SSOService.deleteProvider(
                ctx.drizzle,
                input.orgId,
                input.providerId,
            );
        }),
});
