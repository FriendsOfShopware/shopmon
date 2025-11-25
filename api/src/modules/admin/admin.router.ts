import { z } from 'zod';
import { publicProcedure, router } from '#src/trpc/index.ts';
import { isAdminMiddleware } from '#src/trpc/middleware.ts';
import * as AdminService from './admin.service.ts';

const organizationsPaginationSchema = z.object({
    limit: z.number().min(1).max(100).default(20),
    offset: z.number().min(0).default(0),
    sortBy: z
        .enum(['name', 'slug', 'createdAt', 'shopCount', 'memberCount'])
        .default('createdAt'),
    sortDirection: z.enum(['asc', 'desc']).default('desc'),
    searchField: z.enum(['name', 'slug']).optional(),
    searchOperator: z.enum(['contains']).optional(),
    searchValue: z.string().optional(),
    filterField: z.enum(['name', 'slug']).optional(),
    filterOperator: z.enum(['eq']).optional(),
    filterValue: z.string().optional(),
});

const shopsPaginationSchema = z.object({
    limit: z.number().min(1).max(100).default(20),
    offset: z.number().min(0).default(0),
    sortBy: z
        .enum([
            'name',
            'url',
            'status',
            'shopwareVersion',
            'lastScrapedAt',
            'organizationName',
            'organizationSlug',
            'createdAt',
        ])
        .default('name'),
    sortDirection: z.enum(['asc', 'desc']).default('desc'),
    searchField: z.enum(['name', 'url']).optional(),
    searchOperator: z.enum(['contains']).optional(),
    searchValue: z.string().optional(),
    filterField: z.enum(['name', 'url', 'status']).optional(),
    filterOperator: z.enum(['eq']).optional(),
    filterValue: z.string().optional(),
});

export const adminRouter = router({
    listOrganizations: publicProcedure
        .use(isAdminMiddleware)
        .input(organizationsPaginationSchema.optional())
        .query(async ({ ctx, input }) => {
            return await AdminService.listOrganizations(ctx.drizzle, input);
        }),
    listShops: publicProcedure
        .use(isAdminMiddleware)
        .input(shopsPaginationSchema.optional())
        .query(async ({ ctx, input }) => {
            return await AdminService.listShops(ctx.drizzle, input);
        }),
    getStats: publicProcedure.use(isAdminMiddleware).query(async ({ ctx }) => {
        return await AdminService.getStats(ctx.drizzle);
    }),
});
