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

type CurrentUser = {
    id: number;
    username: string;
    email: string;
    created_at: string;
    avatar: string;
    teams: {
        id: number;
        name: string;
        created_at: string;
        owner_id: number;
        shopCount: number;
        memberCount: number;
    }[];
};

interface Extension {
    name: string,
    label: string,
    active: boolean,
    version: string,
    latestVersion: string | null,
    installed: boolean,
    ratingAverage: number | null,
    storeLink: string | null,
    changelog: ExtensionChangelog[] | null,
    installedAt: string | null,
}

interface ExtensionChangelog {
    version: string
    text: string
    creationDate: string
    isCompatible: boolean;
}

export interface UserExtension extends Extension {
    shops: {
        [key: string]: {
            id: number,
            name: string,
            team_id: number,
            shopware_version: string,
            installed: boolean,
            active: boolean,
            version: string
        }
    }
}

export const accountRouter = router({
    notification: notificationRouter,
    currentUser: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const user = await ctx.drizzle
                .select({
                    id: schema.user.id,
                    username: schema.user.username,
                    email: schema.user.email,
                    created_at: schema.user.created_at,
                })
                .from(schema.user)
                .where(eq(schema.user.id, ctx.user!!))
                .get()

            if (user === undefined) {
                throw new TRPCError({
                    code: 'NOT_FOUND',
                    message: 'User not found',
                })
            }


            const emailMd5 = await md5(user.email);

            const avatar = `https://seccdn.libravatar.org/avatar/${emailMd5}?d=identicon`;

            const teamResult = await ctx.drizzle
                .select({
                    id: schema.team.id,
                    name: schema.team.name,
                    created_at: schema.team.created_at,
                    owner_id: schema.team.owner_id,
                    shopCount: sql<number>`(SELECT COUNT(1) FROM shop WHERE team_id = ${schema.team.id})`,
                    memberCount: sql<number>`(SELECT COUNT(1) FROM user_to_team WHERE team_id = ${schema.team.id})`,
                })
                .from(schema.team)
                .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.team.id))
                .where(eq(schema.userToTeam.user_id, ctx.user!!))
                .all();

            const result: CurrentUser = { ...user, avatar, teams: teamResult }

            return result;
        }),
    updateCurrentUser: publicProcedure.use(loggedInUserMiddleware).input(z.object({
        currentPassword: z.string().min(8),
        newPassword: z.string().min(8).optional(),
        email: z.string().email().optional(),
        username: z.string().min(5).optional(),
    })).mutation(async ({ ctx, input }) => {
        const user = await ctx.drizzle.query.user.findFirst({
            columns: {
                id: true,
                password: true
            },
            where: eq(schema.user.id, ctx.user!!)
        })

        if (user === undefined) {
            throw new TRPCError({
                code: 'NOT_FOUND',
                message: 'User not found',
            })
        }

        if (!bcryptjs.compareSync(input.currentPassword, user.password)) {
            throw new TRPCError({
                code: 'BAD_REQUEST',
                message: 'Invalid password',
            })
        }

        const updates: { password?: string, username?: string, email?: string } = {}

        if (input.newPassword !== undefined) {
            const hash = bcryptjs.hashSync(input.newPassword, 10);

            await Users.revokeUserSessions(ctx.env.kvStorage, ctx.user!!);
            updates['password'] = hash;
        }

        if (input.email !== undefined) {
            updates['email'] = input.email;
        }

        if (input.username !== undefined) {
            updates['username'] = input.username;
        }

        if (Object.keys(updates).length !== 0) {
            await ctx.drizzle.update(schema.user).set(updates).where(eq(schema.user.id, ctx.user!!)).execute();
        }

        return true;
    }),
    deleteCurrentUser: publicProcedure.use(loggedInUserMiddleware).mutation(async ({ ctx }) => {
        await Users.revokeUserSessions(ctx.env.kvStorage, ctx.user!!);
        await Users.delete(ctx.drizzle, ctx.user!!);

        return true;
    }),
    currentUserShops: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
        return await ctx.drizzle
            .select({
                id: schema.shop.id,
                name: schema.shop.name,
                status: schema.shop.status,
                url: schema.shop.url,
                favicon: schema.shop.favicon,
                created_at: schema.shop.created_at,
                last_scraped_at: schema.shop.last_scraped_at,
                shopware_version: schema.shop.shopware_version,
                last_updated: schema.shop.last_updated,
                team_id: schema.shop.team_id,
                team_name: schema.team.name
            })
            .from(schema.shop)
            .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.shop.team_id))
            .innerJoin(schema.team, eq(schema.team.id, schema.shop.team_id))
            .where(eq(schema.userToTeam.user_id, ctx.user!!))
            .orderBy(schema.shop.name)
            .all();
    }),
    currentUserChangelogs: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
        const result = await ctx.drizzle.select({
            id: schema.shopChangelog.id,
            shop_id: schema.shopChangelog.shop_id,
            shop_name: schema.shop.name,
            shop_favicon: schema.shop.favicon,
            extensions: schema.shopChangelog.extensions,
            old_shopware_version: schema.shopChangelog.old_shopware_version,
            new_shopware_version: schema.shopChangelog.new_shopware_version,
            date: schema.shopChangelog.date
        })
            .from(schema.shop)
            .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.shop.team_id))
            .innerJoin(schema.shopChangelog, eq(schema.shopChangelog.shop_id, schema.shop.id))
            .where(eq(schema.userToTeam.user_id, ctx.user!!))
            .limit(10)
            .all();

        for (const row of result) {
            row.extensions = JSON.parse(row.extensions || '{}');
        }

        return result;
    }),
    currentUserApps: publicProcedure.use(loggedInUserMiddleware).query(async ({ ctx }) => {
        const result = await ctx.drizzle.select({
            id: schema.shop.id,
            name: schema.shop.name,
            team_id: schema.shop.team_id,
            shopware_version: schema.shop.shopware_version,
            extensions: schema.shopScrapeInfo.extensions
        })
            .from(schema.shop)
            .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.shop.team_id))
            .innerJoin(schema.shopScrapeInfo, eq(schema.shopScrapeInfo.shop, schema.shop.id))
            .where(eq(schema.userToTeam.user_id, ctx.user!!))
            .orderBy(schema.shop.name)
            .all()

        const json = {} as { [key: string]: UserExtension };

        for (const row of result) {
            const extensions = JSON.parse(row.extensions || '{}') as Extension[];

            for (const extension of extensions) {
                if (json[extension.name] === undefined) {
                    json[extension.name] = extension as UserExtension;
                    json[extension.name].shops = {};
                }

                json[extension.name].shops[row.id] = {
                    'id': row.id,
                    'name': row.name,
                    'team_id': row.team_id,
                    'shopware_version': row.shopware_version,
                    'installed': extension.installed,
                    'active': extension.active,
                    'version': extension.version
                }
            }
        }

        return result;
    }),
});
