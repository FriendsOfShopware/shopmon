import { getConnection } from "../../db";
import bcryptjs from "bcryptjs";
import { ErrorResponse, NoContentResponse } from "../common/response";
import Teams from "../../repository/teams";
import { randomString } from "../../util";
import { sendMailConfirmToUser } from "../../mail/mail";

export const validateEmail = (email: string) => {
    return String(email)
        .toLowerCase()
        .match(
            /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
        );
};

export default async function (req: Request, env: Env) : Promise<Response> {
    const json = await req.json() as { email?: string, password?: string, username?: string };

    if (typeof json.email !== "string" || typeof json.password !== "string" || typeof json.username !== "string") {
        return new ErrorResponse('Missing user, email or password', 400);
    }

    if (!validateEmail(json.email)) {
        return new ErrorResponse('Invalid email', 400);
    }

    if (json.password.length < 8) {
        return new ErrorResponse('Password must be at least 8 characters', 400);
    }

    if (json.username.length < 5) {
        return new ErrorResponse('The username must be at least 5 characters long', 400);
    }

    const con = getConnection(env);

    const result = await con.execute("SELECT 1 FROM user WHERE email = ?", [json.email]);

    if (result.rows.length) {
        return new ErrorResponse('Given email address is already registered', 400);
    }

    const salt = bcryptjs.genSaltSync(10)
    const hashedPassword = await bcryptjs.hash(json.password, salt)

    const token = randomString(32)

    const userInsertResult = await con.execute("INSERT INTO user (email, username, password, verify_code) VALUES (?, ?, ?, ?)", [json.email, json.username, hashedPassword, token]);

    await Teams.createTeam(con, `${json.email}'s Team`, userInsertResult.insertId as string);

    await sendMailConfirmToUser(env, json.email, token);

    return new NoContentResponse();
}