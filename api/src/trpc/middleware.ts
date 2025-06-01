import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { t } from '.';
import { auth } from '../auth';
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
        input: { orgId: string };
        ctx: context & { user: number };
    };

    const result = ctx.drizzle
        .select({
            id: schema.member.id,
        })
        .from(schema.member)
        .where(
            and(
                eq(schema.member.organizationId, input.orgId),
                eq(schema.member.userId, ctx.user.id),
            ),
        )
        .get();

    if (!result) {
        throw new TRPCError({
            code: 'FORBIDDEN',
            message: 'You are not a member of this organization',
        });
    }

    return next();
});

export const hasPermissionMiddleware = (
    permissions: Parameters<
        typeof auth.api.hasPermission
    >[0]['body']['permissions'],
) => {
    return t.middleware(async (opts) => {
        const { ctx, input, next } = opts as typeof opts & {
            input: { orgId: string };
            ctx: context & { user: number };
        };

        const hasPermission = await auth.api.hasPermission({
            body: {
                organizationId: input.orgId,
                permissions,
            },
            headers: ctx.headers,
        });

        if (!hasPermission) {
            throw new TRPCError({
                code: 'FORBIDDEN',
                message: 'You do not have permission to perform this action',
            });
        }

        return next();
    });
};

export const shopMiddleware = t.middleware(async (opts) => {
    const { ctx, input, next } = opts as typeof opts & {
        input: { orgId: string; shopId: number };
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
