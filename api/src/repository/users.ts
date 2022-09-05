import { Connection } from "@planetscale/database/dist";
import Teams from "./teams";

export default class Users {
    static async delete(con: Connection, id: string): Promise<void> {
        const ownerTeams = await con.execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);

        const deletePromises = ownerTeams.rows.map(async row => 
            Teams.deleteTeam(con, row.id) 
        );
        
        await Promise.all(deletePromises);

        await con.execute(`DELETE FROM user WHERE id = ?`, [id]);
        await con.execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }
}