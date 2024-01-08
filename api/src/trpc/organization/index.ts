import { router, publicProcedure } from '..';
import { z } from 'zod';
import {
    loggedInUserMiddleware,
    organizationAdminMiddleware,
    organizationMiddleware,
} from '../middleware';
import Teams from '../../repository/teams';
import { shopRouter } from './shop';

export const organizationRouter = router({
    shop: shopRouter,
    create: publicProcedure
        .input(z.string())
        .use(loggedInUserMiddleware)
        .mutation(async ({ input, ctx }) => {
            return await Teams.create(ctx.drizzle, input, ctx.user!!);
        }),
    update: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                name: z.string(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Teams.update(ctx.drizzle, input.orgId, input.name);
        }),
    delete: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Teams.delete(ctx.drizzle, input.orgId);
        }),
    listMembers: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .query(async ({ input, ctx }) => {
            return await Teams.listMembers(ctx.drizzle, input.orgId);
        }),
    addMember: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                email: z.string().email(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Teams.addMember(ctx.drizzle, input.orgId, input.email);
        }),
    removeMember: publicProcedure
        .input(
            z.object({
                orgId: z.number(),
                userId: z.number(),
            }),
        )
        .use(loggedInUserMiddleware)
        .use(organizationMiddleware)
        .use(organizationAdminMiddleware)
        .mutation(async ({ input, ctx }) => {
            await Teams.removeMember(ctx.drizzle, input.orgId, input.userId);
        }),
});
