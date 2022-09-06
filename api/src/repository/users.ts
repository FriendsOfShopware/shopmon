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

    static async revokeUserSessions(session: KVNamespace, id: string): Promise<void> {
        const accessToken = await session.list({ prefix: `u-${id}-` });
    
        for (const key of accessToken.keys) {
            await session.delete(key.name);
        }

        const refreshToken = await session.list({ prefix: `r-${id}-` });
    
        for (const key of refreshToken.keys) {
            await session.delete(key.name);
        }
    }
}