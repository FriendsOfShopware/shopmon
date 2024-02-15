import { TRPCError, experimental_standaloneMiddleware } from '@trpc/server';
import { t } from '.';
import { and, eq } from 'drizzle-orm';
import type { context } from './context';
import { schema } from '../db';

export const loggedInUserMiddleware = t.middleware(({ ctx, next }) => {
    if (!ctx.user) {
        throw new TRPCError({ code: 'UNAUTHORIZED', message: 'Token expired' });
    }

    return next({
        ctx: {
            user: ctx.user,
        },
    });
});

export const organizationMiddleware = experimental_standaloneMiddleware<{
    input: { orgId: number };
    ctx: context;
}>().create(async ({ ctx, input, next }) => {
    const result = await ctx.drizzle.query.userToOrganization.findFirst({
        columns: {
            userId: true,
        },
        where: and(
            eq(schema.userToOrganization.organizationId, input.orgId),
            eq(schema.userToOrganization.userId, ctx.user as number)
        ),
    });

    if (!result) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});

export const organizationAdminMiddleware = experimental_standaloneMiddleware<{
    input: { orgId: number };
    ctx: context;
}>().create(async ({ ctx, input, next }) => {
    const result = await ctx.drizzle
        .select({
            ownerId: schema.organization.ownerId,
        })
        .from(schema.organization)
        .where(eq(schema.organization.id, input.orgId))
        .get();

    if (!result || result.ownerId !== ctx.user) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});

export const shopMiddleware = experimental_standaloneMiddleware<{
    input: { orgId: number; shopId: number };
    ctx: context;
}>().create(async ({ ctx, input, next }) => {
    const result = await ctx.drizzle
        .select({
            orgId: schema.shop.organizationId,
        })
        .from(schema.shop)
        .where(eq(schema.shop.id, input.shopId))
        .get();

    if (!result || result.orgId !== input.orgId) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});
