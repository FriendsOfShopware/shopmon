import { init } from "@mmyoji/object-validator";
import { getConnection } from "../../db";
import bcryptjs from "bcryptjs";
import { ErrorResponse, JsonResponse } from "../common/response";
import { randomString } from "../../util";

const REFRESH_TOKEN_TTL = 60 * 60 * 6; // 6 hours
const ACCESS_TOKEN_TTL = 60 * 30; // 30 minutes


export interface Token {
    id: number;
}

type OAuthRequest = {
    client_id: string;
    grant_type: string;
    refresh_token: string;
    password: string;
    username: string;
    scopes: string;
}

const validate = init<OAuthRequest>({
    client_id: {
        type: "string",
        required: true
    },
    refresh_token: {
        type: "string",
    },
    grant_type: {
        type: "string",
        required: true
    },
    username: {
        type: "string"
    },
    password: {
        type: "string"
    }
});

export default async function oauth(req: Request, env: Env): Promise<Response> {
    const params = await req.json() as OAuthRequest;

    const errors = validate(params);

    if (errors.length) {
        return new JsonResponse(errors, 400);
    }

    if (params.client_id !== "shopmon") {
        return new ErrorResponse("Invalid client id", 400);
    }

    if (params.grant_type !== "password" && params.grant_type !== "refresh_token") {
        return new JsonResponse({ error: "unsupported_grant_type" }, 400);
    }

    if (params.grant_type === "password") {
        return await handlePasswordGrant(params, env);
    }

    return await handleRefreshTokenGrant(params, env);
}

async function handlePasswordGrant(params: OAuthRequest, env: Env): Promise<Response> {
    const con = getConnection(env);

    const loginError = new ErrorResponse('The user credentials were incorrect.', 401);

    const result = await con.execute("SELECT id, password FROM user WHERE email = ? AND verify_code IS NULL", [params.username.toLowerCase()]);

    if (!result.rows.length) {
        return loginError;
    }

    const user = result.rows[0];

    const passwordIsValid = await bcryptjs.compare(params.password, user.password as string);

    if (!passwordIsValid) {
        return loginError;
    }

    const refreshToken = `r-${user.id}-${randomString(32)}`;
    await env.kvStorage.put(
        refreshToken, 
        JSON.stringify({
            id: user.id
        } as Token),
        {
            expirationTtl: REFRESH_TOKEN_TTL,
        }
    );

    try {
        const accessToken = await getAuthentificationToken(env, refreshToken);

        return new JsonResponse({
            "access_token": accessToken,
            "expires_in": 3600,
            "refresh_token": refreshToken,
            "token_type": "Bearer"
        });
    } catch (e) {
        if (e instanceof Error) {
            return new ErrorResponse(e.message, 400);
        }

        return new ErrorResponse('An error occured', 500);
    }
}

async function getAuthentificationToken(env: Env, refreshToken: string) {
    const token = await env.kvStorage.get(refreshToken);
    const con = getConnection(env);

    if (token === null) {
        throw new Error('Invalid refresh token');
    }

    const data = JSON.parse(token) as Token;

    const result = await con.execute("SELECT id, password FROM user WHERE id = ?", [data.id]);

    if (!result.rows.length) {
        throw new Error('Invalid refresh token');
    }

    const accessToken = `u-${data.id}-${randomString(32)}`;
    await env.kvStorage.put(
        accessToken, 
        JSON.stringify({
            id: data.id
        } as Token),
        {
            expirationTtl: ACCESS_TOKEN_TTL,
        }
    );

    return accessToken;
}

async function handleRefreshTokenGrant(params: OAuthRequest, env: Env): Promise<Response> {
    try {
        const accessToken = await getAuthentificationToken(env, params.refresh_token);

        return new JsonResponse({
            "access_token": accessToken,
            "expires_in": 3600,
            "refresh_token": null,
            "token_type": "Bearer"
        });
    } catch (e) {
        if (e instanceof Error) {
            return new ErrorResponse(e.message, 400);
        }

        return new ErrorResponse('An error occured', 500);
    }
}

