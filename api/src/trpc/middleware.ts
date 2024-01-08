import { TRPCError, experimental_standaloneMiddleware } from '@trpc/server';
import { t } from '.';
import { eq, and } from 'drizzle-orm';
import type { context } from './context'
import { schema } from '../db';

export const loggedInUserMiddleware = t.middleware(({ ctx, next }) => {
    if (!ctx.user) {
        throw new TRPCError({ code: 'UNAUTHORIZED' });
    }

    return next();
});

export const organizationMiddleware = experimental_standaloneMiddleware<{ input: { orgId: number }, ctx: context }>().create(async ({ ctx, input, next }) => {
    const result = await ctx.drizzle.query.userToTeam.findFirst({
        columns: {
            user_id: true
        },
        where: eq(schema.userToTeam.team_id, input.orgId)
    });

    if (!result) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});

export const organizationAdminMiddleware = experimental_standaloneMiddleware<{ input: { orgId: number }, ctx: context }>().create(async ({ ctx, input, next }) => {
    const result = await ctx.drizzle
        .select({
            ownerId: schema.team.owner_id,
        })
        .from(schema.team)
        .where(eq(schema.team.id, input.orgId))
        .get()

    if (!result || result.ownerId !== ctx.user) {
        throw new TRPCError({ code: 'NOT_FOUND' });
    }

    return next();
});
