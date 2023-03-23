import { getConnection } from "../../db";
import Users from "../../repository/users";
import bcryptjs from "bcryptjs";
import { ErrorResponse, NoContentResponse } from "../common/response";
import { validateEmail } from "../auth/register";

type TeamRow = {
    id: string;
    name: string;
    created_at: string;
    is_owner: boolean;
    shopCount: number;
    memberCount: number;
}

export async function accountMe(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const query = env.USE_LOCAL_DATABASE ?
        'SELECT id, username, email, created_at, email as avatar FROM user WHERE id = ?' :
        'SELECT id, username, email, created_at, MD5(email) as avatar FROM user WHERE id = ?';
    const result = await con.execute(query, [req.userId]);

    const json = result.rows[0] as { avatar: string, teams: TeamRow[] };

    json.avatar = `https://www.gravatar.com/avatar/${json.avatar}?d=identicon`;

    const teamResult = await con.execute(`
    SELECT 
	team.id, 
	team.name, 
	team.created_at, 
	(team.owner_id = user_to_team.user_id) as is_owner,
    team.owner_id,
	(SELECT COUNT(1) FROM shop WHERE team_id = team.id) AS shopCount,
	(SELECT COUNT(1) FROM user_to_team WHERE team_id = team.id) AS memberCount
FROM team 
	INNER JOIN user_to_team ON user_to_team.team_id = team.id 
WHERE 
	user_to_team.user_id = ?
    `, [req.userId]);

    json.teams = teamResult.rows as TeamRow[];
    
    return new Response(JSON.stringify(json), {
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
    const { currentPassword, email, newPassword, username } = await req.json() as {currentPassword: string, email: string, newPassword: string, username: string};

    const con = getConnection(env);

    const result = await con.execute('SELECT id, password FROM user WHERE id = ?', [req.userId]);

    if (!result.rows.length) {
        return new ErrorResponse('User not found', 404);
    }

    const user = result.rows[0] as { id: string, password: string };

    if (!bcryptjs.compareSync(currentPassword, user.password)) {
        return new ErrorResponse('Invalid password', 400);
    }

    if (email !== undefined && !validateEmail(email)) {
        return new ErrorResponse('Invalid email', 400);
    }

    if (newPassword !== undefined && newPassword.length < 8) {
        return new ErrorResponse('Password must be at least 8 characters', 400);
    }

    if (newPassword !== undefined) {
        const hash = bcryptjs.hashSync(newPassword, 10);

        await Users.revokeUserSessions(env.kvStorage, req.userId);

        await con.execute('UPDATE user SET password = ? WHERE id = ?', [hash, req.userId]);
    }

    if (email !== undefined) {
        await con.execute('UPDATE user SET email = ? WHERE id = ?', [email, req.userId]);
    }

    if (username !== undefined) {
        await con.execute('UPDATE user SET username = ? WHERE id = ?', [username, req.userId]);
    }

    return new NoContentResponse();
}

export async function listUserShops(req: Request, env: Env): Promise<Response> {
    const con = getConnection(env);

    const res = await con.execute(`
        SELECT 
            shop.id,
            shop.name,
            shop.status,
            shop.url,
            shop.favicon,
            shop.created_at,
            shop.last_scraped_at,
            shop.shopware_version,
            shop.team_id,
            team.name as team_name
        FROM shop 
            INNER JOIN user_to_team ON(user_to_team.team_id = shop.team_id)
            INNER JOIN team ON(team.id = shop.team_id)
        WHERE user_to_team.user_id = ?
        ORDER BY shop.team_id
    `, [req.userId]);

    return new Response(JSON.stringify(res.rows), { 
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}