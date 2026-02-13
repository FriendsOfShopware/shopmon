import { z } from "zod";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { loggedInUserMiddleware, organizationMiddleware } from "#src/trpc/middleware.ts";
import * as ProjectService from "./project.service.ts";

export const projectRouter = router({
  list: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .query(async ({ ctx, input }) => {
      return await ProjectService.listProjects(ctx.drizzle, input.orgId);
    }),
  create: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        name: z.string(),
        description: z.string().optional(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ input, ctx }) => {
      return await ProjectService.createProject(ctx.drizzle, input);
    }),
  update: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
        name: z.string().optional(),
        description: z.string().optional(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ input, ctx }) => {
      return await ProjectService.updateProject(ctx.drizzle, input);
    }),
  delete: publicProcedure
    .input(
      z.object({
        orgId: z.string(),
        projectId: z.number(),
      }),
    )
    .use(loggedInUserMiddleware)
    .use(organizationMiddleware)
    .mutation(async ({ input, ctx }) => {
      return await ProjectService.deleteProject(ctx.drizzle, input);
    }),
});
