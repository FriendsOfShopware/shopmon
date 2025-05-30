import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { t } from '.';
import { schema } from '../db';
import type { context } from './context';

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

export const organizationMiddleware = t.middleware(async (opts) => {
    const { ctx, input, next } = opts as typeof opts & {
        input: { orgId: number };
        ctx: context & { user: number };
    };

    const result = await ctx.drizzle.query.userToOrganization.findFirst({
        columns: {
            userId: true,
        },
        where: and(
            eq(schema.userToOrganization.organizationId, input.orgId),
            eq(schema.userToOrganization.userId, ctx.user.id),
        ),
    });

    if (!result) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});

export const organizationAdminMiddleware = t.middleware(async (opts) => {
    const { ctx, input, next } = opts as typeof opts & {
        input: { orgId: number };
        ctx: context & { user: number };
    };

    const result = await ctx.drizzle
        .select({
            ownerId: schema.organization.ownerId,
        })
        .from(schema.organization)
        .where(eq(schema.organization.id, input.orgId))
        .get();

    if (!result || result.ownerId !== ctx.user.id) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Your are not the owner of the organization',
        });
    }

    return next();
});

export const shopMiddleware = t.middleware(async (opts) => {
    const { ctx, input, next } = opts as typeof opts & {
        input: { orgId: number; shopId: number };
        ctx: context & { user: number };
    };

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
