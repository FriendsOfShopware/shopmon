import { z } from "zod";
import { publicProcedure, router } from "#src/trpc/index.ts";
import {
  loggedInUserMiddleware,
  organizationMiddleware,
  shopMiddleware,
} from "#src/trpc/middleware.ts";
import * as ShopService from "./shop.service.ts";

export const shopRouter = router({
  list: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .query(async ({ ctx, input }) => {
      return await ShopService.listShops(ctx.drizzle, input.orgId);
    }),
  get: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .query(async ({ input, ctx }) => {
      return await ShopService.getShopDetails(ctx.drizzle, input.shopId);
    }),
  create: publicProcedure
    .input(
      z.object({
        name: z.string(),
        shopUrl: z.string().url(),
        clientId: z.string(),
        clientSecret: z.string(),
        projectId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .mutation(async ({ input, ctx }) => {
      return await ShopService.create(ctx.drizzle, ctx.user.id, input);
    }),
  delete: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.deleteShop(ctx.drizzle, input.shopId);
      return true;
    }),
  update: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        shopId: z.number(),
        name: z.string().optional(),
        shopUrl: z.string().url().optional(),
        clientId: z.string().optional(),
        clientSecret: z.string().optional(),
        ignores: z.array(z.string()).optional(),
        projectId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.update(ctx.drizzle, ctx.user.id, input);
    }),
  refreshShop: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        sitespeed: z.boolean().optional(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input }) => {
      await ShopService.refresh(input.shopId, input.sitespeed);
      return true;
    }),
  clearShopCache: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.clearCache(ctx.drizzle, input.shopId);
      return true;
    }),
  rescheduleTask: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        taskId: z.string(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.rescheduleTask(ctx.drizzle, input.shopId, input.taskId);
    }),
  subscribeToNotifications: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.subscribeToNotifications(ctx.drizzle, ctx.user.id, input.shopId);
      return true;
    }),
  unsubscribeFromNotifications: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.unsubscribeFromNotifications(ctx.drizzle, ctx.user.id, input.shopId);
      return true;
    }),
  isSubscribedToNotifications: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .query(async ({ input, ctx }) => {
      return await ShopService.isSubscribedToNotifications(ctx.drizzle, ctx.user.id, input.shopId);
    }),
  updateSitespeedSettings: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        enabled: z.boolean(),
        urls: z.array(z.string()).max(5).optional(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await ShopService.updateSitespeedSettings(
        ctx.drizzle,
        input.shopId,
        input.enabled,
        input.urls,
      );
      return true;
    }),
});
