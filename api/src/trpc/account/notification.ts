import { z } from 'zod';
import * as NotificationService from '#src/service/notification.ts';
import { publicProcedure, router } from '#src/trpc/index.ts';
import { loggedInUserMiddleware } from '#src/trpc/middleware.ts';

export const notificationRouter = router({
    list: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
        return await NotificationService.listNotifications(
            ctx.drizzle,
            ctx.user.id,
        );
    }),
    delete: publicProcedure
        .use(loggedInUserMiddleware)
        .input(z.number().optional())
        .mutation(async ({ ctx, input }) => {
            return await NotificationService.deleteNotification(
                ctx.drizzle,
                ctx.user.id,
                input,
            );
        }),
    markAllRead: publicProcedure
        .use(loggedInUserMiddleware)
        .mutation(async ({ ctx }) => {
            await NotificationService.markAllRead(ctx.drizzle, ctx.user.id);
        }),
});
