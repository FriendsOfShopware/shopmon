import { z } from "zod";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { loggedInUserMiddleware, organizationMiddleware } from "#src/trpc/middleware.ts";
import * as PackagesTokenService from "./packages-token.service.ts";

export const packagesTokenRouter = router({
  configuration: publicProcedure
    .use(loggedInUserMiddleware)
    .query(() => {
      return PackagesTokenService.getConfiguration();
    }),

  list: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .query(async ({ ctx, input }) => {
      return await PackagesTokenService.listTokens(ctx.drizzle, input);
    }),

  create: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
        token: z.string().min(1, "Token is required"),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ ctx, input }) => {
      return await PackagesTokenService.createToken(ctx.drizzle, input);
    }),

  delete: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
        tokenId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ ctx, input }) => {
      await PackagesTokenService.deleteToken(ctx.drizzle, input);
      return true;
    }),

  sync: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
        tokenId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ ctx, input }) => {
      await PackagesTokenService.syncToken(ctx.drizzle, input);
      return true;
    }),
});
