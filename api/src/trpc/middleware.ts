import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { auth } from '../auth.ts';
import { schema } from '../db.ts';
import type { context } from './context.ts';
import { t } from './index.ts';

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

        if (!hasPermission.success) {
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
        ctx: context & { user: NonNullable<context['user']> };
    };

    const result = await ctx.drizzle
        .select({
            orgId: schema.shop.organizationId,
        })
        .from(schema.shop)
        .innerJoin(
            schema.member,
            eq(schema.shop.organizationId, schema.member.organizationId),
        )
        .where(
            and(
                eq(schema.shop.id, input.shopId),
                eq(schema.member.userId, ctx.user.id),
            ),
        )
        .get();

    if (!result) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});
