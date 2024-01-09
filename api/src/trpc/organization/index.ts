import { router, publicProcedure } from '..';
import { z } from 'zod';
import {
    loggedInUserMiddleware,
    organizationAdminMiddleware,
    organizationMiddleware,
} from '../middleware';
import Organizations from '../../repository/organization';
import { shopRouter } from './shop';

export const organizationRouter = router({
    shop: shopRouter,
    create: publicProcedure
        .input(z.string())
        .use(loggedInUserMiddleware)
        .mutation(async ({ input, ctx }) => {
            return await Organizations.create(ctx.drizzle, input, ctx.user);
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
            await Organizations.update(ctx.drizzle, input.orgId, input.name);
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
            await Organizations.deleteById(ctx.drizzle, input.orgId);
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
            return await Organizations.listMembers(ctx.drizzle, input.orgId);
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
            await Organizations.addMember(ctx.drizzle, input.orgId, input.email);
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
            await Organizations.removeMember(ctx.drizzle, input.orgId, input.userId);
        }),
});
