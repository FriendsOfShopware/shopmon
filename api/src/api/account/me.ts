import { getConnection, getKv } from "../../db";
import Users from "../../repository/users";
import bcryptjs from "bcryptjs";
import { ErrorResponse, NoContentResponse } from "../common/response";
import { validateEmail } from "../auth/register";

const revokeTokens = async (userId: string) => {
    const result = await getKv().list({prefix: `u-${userId}-`});
    
    for (const key of result.keys) {
        await getKv().delete(key.name);
    }
}

export async function accountMe(req: Request): Promise<Response> {
    const result = await getConnection().execute('SELECT id, username, email, created_at, MD5(email) as avatar FROM user WHERE id = ?', [req.userId]);

    const json = result.rows[0];

    json.avatar = `https://www.gravatar.com/avatar/${json.avatar}?d=identicon`;

    const teamResult = await getConnection().execute('SELECT team.id, team.name, team.created_at, (team.owner_id = user_to_team.user_id) as is_owner  FROM team INNER JOIN user_to_team ON user_to_team.team_id = team.id WHERE user_to_team.user_id = ?', [req.userId]);

    json.teams = teamResult.rows;
    
    return new Response(JSON.stringify(json), {
        status: 200,
        headers: {
            'Content-Type': 'application/json'
        }
    });
}

export async function accountDelete(req: Request): Promise<Response> {
    try {
        await Users.delete(req.userId);
    } catch (e: any) {
        return new ErrorResponse(e?.message || 'Unknown error');
    }

    await getKv().delete(req.headers.get('token') as string)

    return new NoContentResponse();
}

export async function accountUpdate(req: Request): Promise<Response> {
    const {currentPassword, email, newPassword, username } = await req.json();

    const result = await getConnection().execute('SELECT id, password FROM user WHERE id = ?', [req.userId]);

    if (!result.rows.length) {
        return new ErrorResponse('User not found', 404);
    }

    const user = result.rows[0];

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

        await revokeTokens(req.userId);

        await getConnection().execute('UPDATE user SET password = ? WHERE id = ?', [hash, req.userId]);
    }

    if (email !== undefined) {
        await getConnection().execute('UPDATE user SET email = ? WHERE id = ?', [email, req.userId]);
    }

    if (username !== undefined) {
        await getConnection().execute('UPDATE user SET username = ? WHERE id = ?', [username, req.userId]);
    }

    return new NoContentResponse();
}

export async function listUserShops(req: Request): Promise<Response> {
    const res = await getConnection().execute(`
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