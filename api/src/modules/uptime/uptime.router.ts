import { z } from "zod";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { loggedInUserMiddleware, shopMiddleware } from "#src/trpc/middleware.ts";
import * as UptimeService from "./uptime.service.ts";

export const uptimeRouter = router({
  getData: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .query(async ({ input, ctx }) => {
      return await UptimeService.getUptimeData(ctx.drizzle, input.shopId);
    }),
  updateSettings: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        enabled: z.boolean(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ input, ctx }) => {
      await UptimeService.updateUptimeSettings(ctx.drizzle, input.shopId, input.enabled);
      return true;
    }),
});
