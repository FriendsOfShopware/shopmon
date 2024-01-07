import { getConnection, schema } from "../../db";
import Users from "../../repository/users";
import bcryptjs from "bcryptjs";
import { ErrorResponse, JsonResponse, NoContentResponse } from "../common/response";
import { Extension, UserExtension } from "../../../../shared/shop";
import { sql, eq } from "drizzle-orm";
import { md5 } from "../../crypto";

export async function accountMe(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con
        .select({
            id: schema.user.id,
            username: schema.user.username,
            email: schema.user.email,
            created_at: schema.user.created_at,
        })
        .from(schema.user)
        .where(eq(schema.user.id, req.userId))
        .get()

    if (result === undefined) {
        return new ErrorResponse('User not found', 404);
    }

    const emailMd5 = await md5(result.email);

    // @ts-ignore
    result.avatar = `https://www.gravatar.com/avatar/${emailMd5}?d=identicon`;

    const teamResult = await con
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
        .where(eq(schema.userToTeam.user_id, req.userId))
        .all();

    // @ts-ignore
    result.teams = teamResult;

    return new Response(JSON.stringify(result), {
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}

export async function accountDelete(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    try {
        await Users.delete(con, req.userId);
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
    } catch (e: any) {
        return new ErrorResponse(e?.message || 'Unknown error');
    }

    await env.kvStorage.delete(req.headers.get('token') as string)

    return new NoContentResponse();
}

export async function accountUpdate(req: Request, env: Env): Promise<Response> {
    const { currentPassword, email, newPassword, username } = await req.json() as { currentPassword: string, email: string, newPassword: string, username: string };

    const con = getConnection(env);

    const user = await con.query.user.findFirst({
        columns: {
            id: true,
            password: true
        },
        where: eq(schema.user.id, req.userId)
    })

    if (user === undefined) {
        return new ErrorResponse('User not found', 404);
    }

    if (!bcryptjs.compareSync(currentPassword, user.password)) {
        return new ErrorResponse('Invalid password', 400);
    }

    if (email !== undefined) {
        return new ErrorResponse('Invalid email', 400);
    }

    if (newPassword !== undefined && newPassword.length < 8) {
        return new ErrorResponse('Password must be at least 8 characters', 400);
    }

    const updates: { password?: string, email?: string, username?: string } = {};

    if (newPassword !== undefined) {
        const hash = bcryptjs.hashSync(newPassword, 10);

        await Users.revokeUserSessions(env.kvStorage, req.userId);
        updates['password'] = hash;
    }

    if (email !== undefined) {
        updates['email'] = email;
    }

    if (username !== undefined) {
        updates['username'] = username;
    }

    if (Object.keys(updates).length !== 0) {
        await con.update(schema.user).set(updates).where(eq(schema.user.id, req.userId)).execute();
    }

    return new NoContentResponse();
}

export async function listUserShops(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con
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
        .where(eq(schema.userToTeam.user_id, req.userId))
        .orderBy(schema.shop.name)
        .all();

    return new JsonResponse(result);
}


export async function listUserChangelogs(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con.select({
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
        .where(eq(schema.userToTeam.user_id, req.userId))
        .limit(10)
        .all();

    for (const row of result) {
        row.extensions = JSON.parse(row.extensions || '{}');
    }

    return new JsonResponse(result);
}

export async function listUserApps(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const result = await con.select({
        id: schema.shop.id,
        name: schema.shop.name,
        team_id: schema.shop.team_id,
        shopware_version: schema.shop.shopware_version,
        extensions: schema.shopScrapeInfo.extensions
    })
        .from(schema.shop)
        .innerJoin(schema.userToTeam, eq(schema.userToTeam.team_id, schema.shop.team_id))
        .innerJoin(schema.shopScrapeInfo, eq(schema.shopScrapeInfo.shop, schema.shop.id))
        .where(eq(schema.userToTeam.user_id, req.userId))
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

    return new JsonResponse(json);
}
