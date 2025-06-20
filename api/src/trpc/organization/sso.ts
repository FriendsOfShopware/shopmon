import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { z } from 'zod';
import { ssoProvider } from '../../db.ts';
import { publicProcedure, router } from '../index.ts';
import {
    hasPermissionMiddleware,
    loggedInUserMiddleware,
    organizationMiddleware,
} from '../middleware.ts';

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
            const providers = await ctx.drizzle
                .select()
                .from(ssoProvider)
                .where(eq(ssoProvider.organizationId, input.orgId));

            return providers.map((provider) => {
                if (provider.oidcConfig) {
                    const oidcConfig = JSON.parse(provider.oidcConfig);
                    const { ...safeConfig } = oidcConfig;
                    return {
                        ...provider,
                        oidcConfig: safeConfig,
                    };
                }
                return provider;
            });
        }),

    discoverOpenIdConfig: publicProcedure
        .input(
            z.object({
                issuer: z.string().url('Must be a valid URL'),
            }),
        )
        .use(loggedInUserMiddleware)
        .query(async ({ input }) => {
            try {
                const discoveryUrl = `${input.issuer}/.well-known/openid-configuration`;

                const response = await fetch(discoveryUrl);

                if (!response.ok) {
                    throw new TRPCError({
                        code: 'BAD_REQUEST',
                        message: `Failed to fetch OpenID configuration from ${discoveryUrl}`,
                    });
                }

                const config = (await response.json()) as Record<
                    string,
                    unknown
                >;

                // Extract the required endpoints
                return {
                    issuer: (config.issuer as string) || input.issuer,
                    authorizationEndpoint:
                        config.authorization_endpoint as string,
                    tokenEndpoint: config.token_endpoint as string,
                    jwksEndpoint: config.jwks_uri as string,
                    userInfoEndpoint: config.userinfo_endpoint as string,
                    scopes: (config.scopes_supported as string[]) || [
                        'openid',
                        'profile',
                        'email',
                    ],
                };
            } catch (error) {
                if (error instanceof TRPCError) {
                    throw error;
                }

                throw new TRPCError({
                    code: 'INTERNAL_SERVER_ERROR',
                    message:
                        error instanceof Error
                            ? error.message
                            : 'Failed to discover OpenID configuration',
                });
            }
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
            // Get existing provider to preserve client secret if not provided
            const existingProvider = await ctx.drizzle
                .select()
                .from(ssoProvider)
                .where(
                    and(
                        eq(ssoProvider.providerId, input.providerId),
                        eq(ssoProvider.organizationId, input.orgId),
                    ),
                )
                .get();

            if (!existingProvider) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'SSO provider not found',
                });
            }

            let existingOidcConfig: Record<string, unknown> = {};
            if (existingProvider.oidcConfig) {
                existingOidcConfig = JSON.parse(existingProvider.oidcConfig);
            }

            // Merge existing config with new values, preserving all existing fields
            const updatedOidcConfig: Record<string, unknown> = {
                ...existingOidcConfig,
                clientId: input.clientId,
                authorizationEndpoint: input.authorizationEndpoint,
                tokenEndpoint: input.tokenEndpoint,
                jwksEndpoint: input.jwksEndpoint,
            };

            // Only update client secret if provided
            if (input.clientSecret) {
                updatedOidcConfig.clientSecret = input.clientSecret;
            }

            const oidcConfig = JSON.stringify(updatedOidcConfig);

            await ctx.drizzle
                .update(ssoProvider)
                .set({
                    domain: input.domain,
                    issuer: input.issuer,
                    oidcConfig,
                })
                .where(
                    and(
                        eq(ssoProvider.providerId, input.providerId),
                        eq(ssoProvider.organizationId, input.orgId),
                    ),
                )
                .returning();

            return true;
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
            const deleted = await ctx.drizzle
                .delete(ssoProvider)
                .where(
                    and(
                        eq(ssoProvider.providerId, input.providerId),
                        eq(ssoProvider.organizationId, input.orgId),
                    ),
                )
                .returning();

            if (deleted.length === 0) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'SSO provider not found',
                });
            }

            return { success: true };
        }),
});
