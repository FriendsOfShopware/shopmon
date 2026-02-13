import { notificationRouter } from "#src/modules/notification/notification.router.ts";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { loggedInUserMiddleware } from "#src/trpc/middleware.ts";
import * as AccountService from "./user.service.ts";

export const accountRouter = router({
  notification: notificationRouter,
  currentUser: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getCurrentUser(ctx.drizzle, ctx.user.id);
  }),
  currentUserExtensions: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getCurrentUserExtensions(ctx.drizzle, ctx.user.id);
  }),
  listOrganizations: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.listOrganizations(ctx.drizzle, ctx.user.id);
  }),
  currentUserShops: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getCurrentUserShops(ctx.drizzle, ctx.user.id);
  }),
  currentUserProjects: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getCurrentUserProjects(ctx.drizzle, ctx.user.id);
  }),
  currentUserChangelogs: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getCurrentUserChangelogs(ctx.drizzle, ctx.user.id);
  }),
  subscribedShops: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
    return await AccountService.getSubscribedShops(
      ctx.drizzle,
      ctx.user.id,
      ctx.user.notifications,
    );
  }),
});
