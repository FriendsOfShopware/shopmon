import { TRPCError } from "@trpc/server";
import { desc, eq } from "drizzle-orm";
import { z } from "zod";
import { deployment, shop, project } from "#src/db.ts";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { loggedInUserMiddleware, shopMiddleware } from "#src/trpc/middleware.ts";
import { deleteDeploymentOutput, getDeploymentOutput } from "./deployment.storage.ts";

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
          id: deployment.id,
          name: deployment.name,
          command: deployment.command,
          returnCode: deployment.returnCode,
          startDate: deployment.startDate,
          endDate: deployment.endDate,
          executionTime: deployment.executionTime,
          reference: deployment.reference,
          gitUrl: project.gitUrl,
          createdAt: deployment.createdAt,
        })
        .from(deployment)
        .innerJoin(shop, eq(shop.id, deployment.shopId))
        .innerJoin(project, eq(project.id, shop.projectId))
        .where(eq(deployment.shopId, input.shopId))
        .orderBy(desc(deployment.createdAt))
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
      const result = await ctx.drizzle
        .select({
          id: deployment.id,
          shopId: deployment.shopId,
          name: deployment.name,
          command: deployment.command,
          returnCode: deployment.returnCode,
          startDate: deployment.startDate,
          endDate: deployment.endDate,
          executionTime: deployment.executionTime,
          composer: deployment.composer,
          reference: deployment.reference,
          gitUrl: project.gitUrl,
          createdAt: deployment.createdAt,
        })
        .from(deployment)
        .innerJoin(shop, eq(shop.id, deployment.shopId))
        .innerJoin(project, eq(project.id, shop.projectId))
        .where(eq(deployment.id, input.deploymentId));

      const deploymentRecord = result[0];

      if (!deploymentRecord) {
        throw new TRPCError({
          code: "NOT_FOUND",
          message: "Deployment not found",
        });
      }

      if (deploymentRecord.shopId !== input.shopId) {
        throw new TRPCError({
          code: "FORBIDDEN",
          message: "Deployment does not belong to this shop",
        });
      }

      const output = await getDeploymentOutput(deploymentRecord.id);

      return { ...deploymentRecord, output };
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
      const result = await ctx.drizzle
        .select()
        .from(deployment)
        .where(eq(deployment.id, input.deploymentId));

      const deploymentRecord = result[0];

      if (!deploymentRecord) {
        throw new TRPCError({
          code: "NOT_FOUND",
          message: "Deployment not found",
        });
      }

      if (deploymentRecord.shopId !== input.shopId) {
        throw new TRPCError({
          code: "FORBIDDEN",
          message: "Deployment does not belong to this shop",
        });
      }

      await deleteDeploymentOutput(input.deploymentId);

      await ctx.drizzle.delete(deployment).where(eq(deployment.id, input.deploymentId));

      return { success: true };
    }),
});
