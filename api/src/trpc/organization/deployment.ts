import { TRPCError } from "@trpc/server";
import { desc, eq } from "drizzle-orm";
import { z } from "zod";
import { schema } from "../../db.ts";
import { publicProcedure, router } from "../index.ts";
import { loggedInUserMiddleware, shopMiddleware } from "../middleware.ts";

export const deploymentRouter = router({
  list: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        limit: z.number().min(1).max(100).default(50),
        offset: z.number().min(0).default(0),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .query(async ({ ctx, input }) => {
      const deployments = await ctx.drizzle
        .select({
          id: schema.deployment.id,
          name: schema.deployment.name,
          command: schema.deployment.command,
          returnCode: schema.deployment.returnCode,
          startDate: schema.deployment.startDate,
          endDate: schema.deployment.endDate,
          executionTime: schema.deployment.executionTime,
          reference: schema.deployment.reference,
          gitUrl: schema.project.gitUrl,
          createdAt: schema.deployment.createdAt,
        })
        .from(schema.deployment)
        .innerJoin(schema.shop, eq(schema.shop.id, schema.deployment.shopId))
        .innerJoin(schema.project, eq(schema.project.id, schema.shop.projectId))
        .where(eq(schema.deployment.shopId, input.shopId))
        .orderBy(desc(schema.deployment.createdAt))
        .limit(input.limit)
        .offset(input.offset);

      return deployments;
    }),

  get: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        deploymentId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .query(async ({ ctx, input }) => {
      const [deployment] = await ctx.drizzle
        .select({
          id: schema.deployment.id,
          shopId: schema.deployment.shopId,
          name: schema.deployment.name,
          command: schema.deployment.command,
          returnCode: schema.deployment.returnCode,
          startDate: schema.deployment.startDate,
          endDate: schema.deployment.endDate,
          executionTime: schema.deployment.executionTime,
          composer: schema.deployment.composer,
          reference: schema.deployment.reference,
          gitUrl: schema.project.gitUrl,
          createdAt: schema.deployment.createdAt,
        })
        .from(schema.deployment)
        .innerJoin(schema.shop, eq(schema.shop.id, schema.deployment.shopId))
        .innerJoin(schema.project, eq(schema.project.id, schema.shop.projectId))
        .where(eq(schema.deployment.id, input.deploymentId));

      if (!deployment) {
        throw new TRPCError({
          code: "NOT_FOUND",
          message: "Deployment not found",
        });
      }

      if (deployment.shopId !== input.shopId) {
        throw new TRPCError({
          code: "FORBIDDEN",
          message: "Deployment does not belong to this shop",
        });
      }

      return deployment;
    }),

  delete: publicProcedure
    .input(
      z.object({
        shopId: z.number(),
        deploymentId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(shopMiddleware)
    .mutation(async ({ ctx, input }) => {
      const [deployment] = await ctx.drizzle
        .select()
        .from(schema.deployment)
        .where(eq(schema.deployment.id, input.deploymentId));

      if (!deployment) {
        throw new TRPCError({
          code: "NOT_FOUND",
          message: "Deployment not found",
        });
      }

      if (deployment.shopId !== input.shopId) {
        throw new TRPCError({
          code: "FORBIDDEN",
          message: "Deployment does not belong to this shop",
        });
      }

      await ctx.drizzle
        .delete(schema.deployment)
        .where(eq(schema.deployment.id, input.deploymentId));

      return { success: true };
    }),
});
