import { getConnection } from "../db";
import Teams from "./teams";

export default class Users {
    static async delete(id: string): Promise<void> {
        const ownerTeams = await getConnection().execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);

        for (let row of ownerTeams.rows) {
            await Teams.deleteTeam(row.id);
        }

        await getConnection().execute(`DELETE FROM user WHERE id = ?`, [id]);
        await getConnection().execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }
}