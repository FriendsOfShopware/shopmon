import { getConnection } from "../db";

interface TeamMember {
    id: number;
    email: string;
}

export default class Teams {
    static async createTeam(name: string, ownerId: string): Promise<string> {
        const teamInsertResult = await getConnection().execute("INSERT INTO team (name, owner_id) VALUES (?, ?)", [name, ownerId]);

        await getConnection().execute("INSERT INTO user_to_team (user_id, team_id) VALUES (?, ?)", [ownerId, teamInsertResult.insertId]);

        return teamInsertResult.insertId as string;
    }

    static async listMembers(teamId: string): Promise<TeamMember[]> {
        const result = await getConnection().execute(`SELECT
        user.id,
        user.username,
        user.email
    FROM user_to_team 
    INNER JOIN user ON(user.id = user_to_team.user_id)
    WHERE user_to_team.team_id = ?`, [teamId]);

        return result.rows as TeamMember[];
    }

    static async addMember(teamId: string, email: string): Promise<void> {
        const member = await getConnection().execute(`SELECT id FROM user WHERE email = ?`, [email]);

        if (member.rows.length === 0) {
            return;
        }

        await getConnection().execute(`INSERT INTO user_to_team (team_id, user_id) VALUES(?, ?)`, [teamId, member.rows[0].id]);
    }

    static async removeMember(teamId: string, userId: string): Promise<void> {
        await getConnection().execute(`DELETE FROM user_to_team WHERE team_id = ? AND user_id = ?`, [teamId, userId]);
    }

    static async deleteTeam(teamId: string): Promise<void> {
        // Fetch all shops in the team
        const shops = await getConnection().execute(`SELECT id FROM shop WHERE team_id = ?`, [teamId]);

        const shopIds = shops.rows.map(shop => shop.id);

        // Delete shops
        await getConnection().execute(`DELETE FROM shop WHERE team_id = ?`, [teamId]);
        await getConnection().execute(`DELETE FROM shop_scrape_info WHERE shop_id IN (?)`, [shopIds]);

        // Delete team
        await getConnection().execute(`DELETE FROM team WHERE id = ?`, [teamId]);
        await getConnection().execute(`DELETE FROM user_to_team WHERE team_id = ?`, [teamId]);
    }
}