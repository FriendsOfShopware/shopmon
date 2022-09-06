import { Connection } from "@planetscale/database/dist";

interface TeamMember {
    id: number;
    email: string;
}

export default class Teams {
    static async createTeam(con: Connection, name: string, ownerId: string): Promise<string> {
        const teamInsertResult = await con.execute("INSERT INTO team (name, owner_id) VALUES (?, ?)", [name, ownerId]);

        await con.execute("INSERT INTO user_to_team (user_id, team_id) VALUES (?, ?)", [ownerId, teamInsertResult.insertId]);

        return teamInsertResult.insertId as string;
    }

    static async listMembers(con: Connection, teamId: string): Promise<TeamMember[]> {
        const result = await con.execute(`SELECT
        user.id,
        user.username,
        user.email
    FROM user_to_team 
    INNER JOIN user ON(user.id = user_to_team.user_id)
    WHERE user_to_team.team_id = ?`, [teamId]);

        return result.rows as TeamMember[];
    }

    static async addMember(con: Connection, teamId: string, email: string): Promise<void> {
        const member = await con.execute(`SELECT id FROM user WHERE email = ?`, [email]);

        if (member.rows.length === 0) {
            return;
        }

        await con.execute(`INSERT INTO user_to_team (team_id, user_id) VALUES(?, ?)`, [teamId, member.rows[0].id]);
    }

    static async removeMember(con: Connection, teamId: string, userId: string): Promise<void> {
        await con.execute(`DELETE FROM user_to_team WHERE team_id = ? AND user_id = ?`, [teamId, userId]);
    }

    static async deleteTeam(con: Connection, teamId: string): Promise<void> {
        // Fetch all shops in the team
        const shops = await con.execute(`SELECT id FROM shop WHERE team_id = ?`, [teamId]);

        const shopIds = shops.rows.map(shop => shop.id);

        // Delete shops
        await con.execute(`DELETE FROM shop WHERE team_id = ?`, [teamId]);
        await con.execute(`DELETE FROM shop_scrape_info WHERE shop_id IN (?)`, [shopIds]);

        // Delete team
        await con.execute(`DELETE FROM team WHERE id = ?`, [teamId]);
        await con.execute(`DELETE FROM user_to_team WHERE team_id = ?`, [teamId]);
    }
}