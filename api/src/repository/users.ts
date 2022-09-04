import { Connection } from "@planetscale/database/dist";
import { getConnection } from "../db";
import Teams from "./teams";

export default class Users {
    static async delete(con: Connection, id: string): Promise<void> {
        const ownerTeams = await con.execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);

        for (let row of ownerTeams.rows) {
            await Teams.deleteTeam(con, row.id);
        }

        await con.execute(`DELETE FROM user WHERE id = ?`, [id]);
        await con.execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }
}