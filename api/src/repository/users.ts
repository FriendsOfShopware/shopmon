import { getConnection } from "../db";

export default class Users {
    static async delete(id: string): Promise<void> {
        const ownerTeams = await getConnection().execute(`SELECT id FROM team WHERE owner_id = ?`, [id]);

        if (ownerTeams.rows.length > 0) {
            throw new Error(`Cannot delete user that is the owner of a team. Delete all teams before.`);
        }

        await getConnection().execute(`DELETE FROM user WHERE id = ?`, [id]);
        await getConnection().execute(`DELETE FROM user_to_team WHERE user_id = ?`, [id]);
    }
}