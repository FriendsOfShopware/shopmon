import { TRPCError } from '@trpc/server';
import { eq } from 'drizzle-orm';
import { z } from 'zod';
import { schema } from '../../db.ts';
import { publicProcedure, router } from '../index.ts';
import {
    loggedInUserMiddleware,
    organizationMiddleware,
} from '../middleware.ts';

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
            const result = await ctx.drizzle
                .select({
                    id: schema.project.id,
                    name: schema.project.name,
                    description: schema.project.description,
                    createdAt: schema.project.createdAt,
                    updatedAt: schema.project.updatedAt,
                    organizationId: schema.project.organizationId,
                })
                .from(schema.project)
                .where(eq(schema.project.organizationId, input.orgId))
                .all();

            return result === undefined ? [] : result;
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
            const result = await ctx.drizzle
                .insert(schema.project)
                .values({
                    organizationId: input.orgId,
                    name: input.name,
                    description: input.description,
                    createdAt: new Date(),
                    updatedAt: new Date(),
                })
                .execute();

            return Number(result.lastInsertRowid);
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
            // Check if project belongs to organization
            const project = await ctx.drizzle.query.project.findFirst({
                where: eq(schema.project.id, input.projectId),
            });

            if (!project || project.organizationId !== input.orgId) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'Project not found',
                });
            }

            const updateData: {
                name?: string;
                description?: string;
                updatedAt: Date;
            } = { updatedAt: new Date() };
            if (input.name !== undefined) {
                updateData.name = input.name;
            }
            if (input.description !== undefined) {
                updateData.description = input.description;
            }

            await ctx.drizzle
                .update(schema.project)
                .set(updateData)
                .where(eq(schema.project.id, input.projectId))
                .execute();

            return true;
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
            // Check if project belongs to organization
            const project = await ctx.drizzle.query.project.findFirst({
                where: eq(schema.project.id, input.projectId),
            });

            if (!project || project.organizationId !== input.orgId) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'Project not found',
                });
            }

            // Check if there are shops assigned to this project
            const shops = await ctx.drizzle.query.shop.findMany({
                where: eq(schema.shop.projectId, input.projectId),
            });

            if (shops.length > 0) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message:
                        'Cannot delete project with assigned shops. Please reassign or delete the shops first.',
                });
            }

            await ctx.drizzle
                .delete(schema.project)
                .where(eq(schema.project.id, input.projectId))
                .execute();

            return true;
        }),
});
