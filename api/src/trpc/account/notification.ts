import { router, publicProcedure } from '..';
import { z } from 'zod';
import { loggedInUserMiddleware } from '../middleware';
import { eq, and } from 'drizzle-orm';
import { schema } from '../../db';

export const notificationRouter = router({
    list: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
        const results = await ctx.drizzle.query.userNotification.findMany({
            where: eq(schema.userNotification.userId, ctx.user),
        });

        return results;
    }),
    delete: publicProcedure
        .use(loggedInUserMiddleware)
        .input(z.number().optional())
        .mutation(async ({ ctx, input }) => {
            if (input) {
                await ctx.drizzle
                    .delete(schema.userNotification)
                    .where(
                        and(
                            eq(schema.userNotification.id, input),
                            eq(schema.userNotification.userId, ctx.user),
                        ),
                    )
                    .execute();

                return true;
            }

            await ctx.drizzle
                .delete(schema.userNotification)
                .where(eq(schema.userNotification.userId, ctx.user))
                .execute();

            return true;
        }),
    markAllRead: publicProcedure
        .use(loggedInUserMiddleware)
        .mutation(async ({ ctx }) => {
            await ctx.drizzle
                .update(schema.userNotification)
                .set({ read: true })
                .where(eq(schema.userNotification.userId, ctx.user))
                .execute();
        }),
});
