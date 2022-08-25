import { getConnection } from "../../db";
import bcryptjs from "bcryptjs";
import { ErrorResponse, NoContentResponse } from "../common/response";
import Teams from "../../repository/teams";

export const validateEmail = (email: string) => {
    return String(email)
      .toLowerCase()
      .match(
        /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
      );
  };

export default async function (req: Request) : Promise<Response> {
    const json = await req.json();

    if (json.email === undefined || json.password === undefined) {
        return new ErrorResponse('Missing email or password', 400);
    }

    if (!validateEmail(json.email)) {
        return new ErrorResponse('Invalid email', 400);
    }

    if (json.password.length < 8) {
        return new ErrorResponse('Password must be at least 8 characters', 400);
    }

    const result = await getConnection().execute("SELECT 1 FROM user WHERE email = ?", [json.email]);

    if (result.rows.length) {
        return new ErrorResponse('User already exists', 400);
    }

    const salt = bcryptjs.genSaltSync(10)
    const hashedPassword = await bcryptjs.hash(json.password, salt)

    const userInsertResult = await getConnection().execute("INSERT INTO user (email, password) VALUES (?, ?)", [json.email, hashedPassword]);

    await Teams.createTeam(`${json.email}'s Team`, userInsertResult.insertId as string);

    return new NoContentResponse();
}