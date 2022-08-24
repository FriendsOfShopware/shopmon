import { getConnection } from "../../db";
import bcryptjs from "bcryptjs";

const validateEmail = (email: string) => {
    return String(email)
      .toLowerCase()
      .match(
        /^(([^<>()[\]\\.,;:\s@"]+(\.[^<>()[\]\\.,;:\s@"]+)*)|(".+"))@((\[[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\.[0-9]{1,3}\])|(([a-zA-Z\-0-9]+\.)+[a-zA-Z]{2,}))$/
      );
  };

export default async function (req: Request) : Promise<Response> {
    const json = await req.json();

    if (json.email === undefined || json.password === undefined) {
        return new Response('Missing email or password', {
            status: 400
        });
    }

    if (!validateEmail(json.email)) {
        return new Response('Invalid email', {
            status: 400
        });
    }

    if (json.password.length < 8) {
        return new Response('Password must be at least 8 characters', {
            status: 400
        });
    }

    const result = await getConnection().execute("SELECT 1 FROM user WHERE email = ?", [json.email]);

    if (result.rows.length) {
        return new Response('User already exists', {
            status: 400
        });
    }

    const salt = bcryptjs.genSaltSync(10)
    const hashedPassword = bcryptjs.hashSync(json.password, salt)

    const userInsertResult = await getConnection().execute("INSERT INTO user (email, password, salt, created_at) VALUES 0(?, ?, ?, NOW())", [json.email, hashedPassword, salt]);

    const teamInsertResult = await getConnection().execute("INSERT INTO team (name, owner_id, created_at) VALUES (?, ?, NOW())", [`${json.email}'s Team`, userInsertResult.insertId]);

    await getConnection().execute("INSERT INTO user_to_team (user_id, team_id) VALUES (?, ?)", [userInsertResult.insertId, teamInsertResult.insertId]);

    return new Response("OK", {
        status: 200
    });
}