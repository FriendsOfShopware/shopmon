import { TRPCError } from '@trpc/server';
import { router, publicProcedure } from '..';
import { md5 } from '../../crypto';
import { schema } from '../../db';
import { loggedInUserMiddleware } from '../middleware';
import { eq, sql } from 'drizzle-orm';
import { z } from 'zod';
import Users from '../../repository/users';
import bcryptjs from 'bcryptjs';
import { notificationRouter } from './notification';

interface Extension {
    name: string;
    label: string;
    active: boolean;
    version: string;
    latestVersion: string | null;
    installed: boolean;
    ratingAverage: number | null;
    storeLink: string | null;
    changelog: ExtensionChangelog[] | null;
    installedAt: string | null;
}

interface ExtensionChangelog {
    version: string;
    text: string;
    creationDate: string;
    isCompatible: boolean;
}

export interface UserExtension extends Extension {
    shops: {
        [key: string]: {
            id: number;
            name: string;
            team_id: number;
            shopware_version: string;
            installed: boolean;
            active: boolean;
            version: string;
        };
    };
}

export const accountRouter = router({
    notification: notificationRouter,
    currentUser: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const user = await ctx.drizzle
                .select({
                    id: schema.user.id,
                    displayName: schema.user.displayName,
                    email: schema.user.email,
                    createdAt: schema.user.createdAt,
                })
                .from(schema.user)
                .where(eq(schema.user.id, ctx.user))
                .get();

            if (user === undefined) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'User not found',
                });
            }

            const emailMd5 = await md5(user.email);

            const avatar = `https://seccdn.libravatar.org/avatar/${emailMd5}?d=identicon`;

            const organizations = await ctx.drizzle
                .select({
                    id: schema.organization.id,
                    name: schema.organization.name,
                    createdAt: schema.organization.createdAt,
                    ownerId: schema.organization.ownerId,
                    shopCount: sql<number>`(SELECT COUNT(1) FROM ${schema.shop} WHERE ${schema.shop.organizationId} = ${schema.organization.id})`,
                    memberCount: sql<number>`(SELECT COUNT(1) FROM ${schema.userToOrganization} WHERE ${schema.userToOrganization.organizationId} = ${schema.organization.id})`,
                })
                .from(schema.organization)
                .innerJoin(
                    schema.userToOrganization,
                    eq(
                        schema.userToOrganization.organizationId,
                        schema.organization.id,
                    ),
                )
                .where(eq(schema.userToOrganization.userId, ctx.user))
                .all();

            return { ...user, avatar, organizations };
        }),
    updateCurrentUser: publicProcedure
        .use(loggedInUserMiddleware)
        .input(
            z.object({
                currentPassword: z.string().min(8),
                newPassword: z.string().min(8).optional(),
                email: z.string().email().optional(),
                displayName: z.string().min(5).optional(),
            }),
        )
        .mutation(async ({ ctx, input }) => {
            const user = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                    password: true,
                },
                where: eq(schema.user.id, ctx.user),
            });

            if (user === undefined) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'User not found',
                });
            }

            if (!bcryptjs.compareSync(input.currentPassword, user.password)) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid password',
                });
            }

            const updates: {
                password?: string;
                displayName?: string;
                email?: string;
            } = {};

            if (input.newPassword !== undefined) {
                const hash = bcryptjs.hashSync(input.newPassword, 10);

                await Users.revokeUserSessions(ctx.env.kvStorage, ctx.user);
                updates.password = hash;
            }

            if (input.email !== undefined) {
                updates.email = input.email;
            }

            if (input.displayName !== undefined) {
                updates.displayName = input.displayName;
            }

            if (Object.keys(updates).length !== 0) {
                await ctx.drizzle
                    .update(schema.user)
                    .set(updates)
                    .where(eq(schema.user.id, ctx.user))
                    .execute();
            }

            return true;
        }),
    deleteCurrentUser: publicProcedure
        .use(loggedInUserMiddleware)
        .mutation(async ({ ctx }) => {
            await Users.revokeUserSessions(ctx.env.kvStorage, ctx.user);
            await Users.deleteById(ctx.drizzle, ctx.user);

            return true;
        }),
    currentUserShops: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            return await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    status: schema.shop.status,
                    url: schema.shop.url,
                    favicon: schema.shop.favicon,
                    createdAt: schema.shop.createdAt,
                    lastScrapedAt: schema.shop.lastScrapedAt,
                    shopwareVersion: schema.shop.shopwareVersion,
                    lastUpdated: schema.shop.lastUpdated,
                    organizationId: schema.shop.organizationId,
                    organizationName: schema.organization.name,
                })
                .from(schema.shop)
                .innerJoin(
                    schema.userToOrganization,
                    eq(
                        schema.userToOrganization.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.organization,
                    eq(schema.organization.id, schema.shop.organizationId),
                )
                .where(eq(schema.userToOrganization.userId, ctx.user))
                .orderBy(schema.shop.name)
                .all();
        }),
    currentUserChangelogs: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle
                .select({
                    id: schema.shopChangelog.id,
                    shop_id: schema.shopChangelog.shopId,
                    shop_name: schema.shop.name,
                    shop_favicon: schema.shop.favicon,
                    extensions: schema.shopChangelog.extensions,
                    old_shopware_version:
                        schema.shopChangelog.oldShopwareVersion,
                    new_shopware_version:
                        schema.shopChangelog.newShopwareVersion,
                    date: schema.shopChangelog.date,
                })
                .from(schema.shop)
                .innerJoin(
                    schema.userToOrganization,
                    eq(
                        schema.userToOrganization.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.shopChangelog,
                    eq(schema.shopChangelog.shopId, schema.shop.id),
                )
                .where(eq(schema.userToOrganization.userId, ctx.user))
                .limit(10)
                .all();

            return result;
        }),
    currentUserApps: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle
                .select({
                    id: schema.shop.id,
                    name: schema.shop.name,
                    team_id: schema.shop.organizationId,
                    shopware_version: schema.shop.shopwareVersion,
                    extensions: schema.shopScrapeInfo.extensions,
                })
                .from(schema.shop)
                .innerJoin(
                    schema.userToOrganization,
                    eq(
                        schema.userToOrganization.organizationId,
                        schema.shop.organizationId,
                    ),
                )
                .innerJoin(
                    schema.shopScrapeInfo,
                    eq(schema.shopScrapeInfo.shopId, schema.shop.id),
                )
                .where(eq(schema.userToOrganization.userId, ctx.user))
                .orderBy(schema.shop.name)
                .all();

            const json = {} as { [key: string]: UserExtension };

            for (const row of result) {
                for (const extension of row.extensions) {
                    if (json[extension.name] === undefined) {
                        json[extension.name] = extension as UserExtension;
                        json[extension.name].shops = {};
                    }

                    json[extension.name].shops[row.id] = {
                        id: row.id,
                        name: row.name,
                        team_id: row.team_id,
                        shopware_version: row.shopware_version,
                        installed: extension.installed,
                        active: extension.active,
                        version: extension.version,
                    };
                }
            }

            return Object.values(json);
        }),
});
