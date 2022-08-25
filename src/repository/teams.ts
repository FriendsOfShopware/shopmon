import { JsonResponse } from "../api/common/response";
import { getConnection } from "../db";

interface TeamMember {
    id: number;
    email: string;
}

export default class Teams {
    static async listMembers(teamId: string): Promise<TeamMember[]> {
        const result = await getConnection().execute(`SELECT
        user.id,
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

    static async removeMember(teamId: string, email: string): Promise<void> {
        const member = await getConnection().execute(`SELECT id FROM user WHERE email = ?`, [email]);

        if (member.rows.length === 0) {
            return;
        }

        await getConnection().execute(`DELETE FROM user_to_team WHERE team_id = ? AND user_id = ?`, [teamId, member.rows[0].id]);
    }
}