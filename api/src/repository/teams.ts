import { Drizzle, schema } from '../db'
import { eq, and, inArray } from 'drizzle-orm';
import Users from "./users";

interface TeamMember {
    id: number;
    email: string;
}

export default class Teams {
    static async createTeam(con: Drizzle, name: string, ownerId: number): Promise<string> {
        const teamInsertResult = await con
            .insert(schema.team)
            .values({
                name,
                owner_id: ownerId,
                created_at: (new Date().toISOString()),
            })
            .execute();

        await con
            .insert(schema.userToTeam)
            .values({
                team_id: teamInsertResult.meta.last_row_id,
                user_id: ownerId,
            })

        return teamInsertResult.meta.last_row_id.toString();
    }

    static async listMembers(con: Drizzle, teamId: number): Promise<TeamMember[]> {
        const result = await con.select({
            id: schema.user.id,
            username: schema.user.username,
            email: schema.user.email,
        }).from(schema.user)
            .innerJoin(schema.userToTeam, eq(schema.userToTeam.user_id, schema.user.id))
            .where(eq(schema.userToTeam.team_id, teamId))
            .all();

        return result as TeamMember[];
    }

    static async addMember(con: Drizzle, teamId: string, email: string): Promise<void> {
        const exists = await Users.existsByEmail(con, email);

        if (exists === false) {
            throw new Error("User not found");
        }

        await con
            .insert(schema.userToTeam)
            .values({
                team_id: parseInt(teamId),
                user_id: parseInt(email),
            })
            .execute();
    }

    static async removeMember(con: Drizzle, teamId: string, userId: string): Promise<void> {
        await con
            .delete(schema.userToTeam)
            .where(and(eq(schema.userToTeam.team_id, parseInt(teamId)), eq(schema.userToTeam.user_id, parseInt(userId))))
            .execute();
    }

    static async deleteTeam(con: Drizzle, teamId: number): Promise<void> {
        const shops = await con.query.shop.findMany({
            columns: {
                id: true,
            },
            where: eq(schema.shop.team_id, teamId),
        })

        const shopIds = shops.map(shop => shop.id);

        // Delete shops associated with team
        await con
            .delete(schema.shop)
            .where(eq(schema.shop.team_id, teamId))
            .execute();

        if (shopIds.length > 0) {
            await con
                .delete(schema.shopScrapeInfo)
                .where(inArray(schema.shopScrapeInfo.shop, shopIds))
                .execute();
        }

        // Delete team
        await con
            .delete(schema.team)
            .where(eq(schema.team.id, teamId))
            .execute();

        await con
            .delete(schema.userToTeam)
            .where(eq(schema.userToTeam.team_id, teamId))
            .execute();
    }

    static async updateTeam(con: Drizzle, teamId: string, name: string): Promise<void> {
        await con
            .update(schema.team)
            .set({
                name
            })
            .where(eq(schema.team.id, parseInt(teamId)))
            .execute()
    }
}
