import { Drizzle, schema } from '../db';
import { eq, and, inArray } from 'drizzle-orm';
import Users from './users';

interface TeamMember {
    id: number;
    email: string;
}

async function create(
    con: Drizzle,
    name: string,
    ownerId: number,
): Promise<string> {
    const teamInsertResult = await con
        .insert(schema.organization)
        .values({
            name,
            ownerId: ownerId,
            createdAt: new Date(),
        })
        .execute();

    await con.insert(schema.userToOrganization).values({
        organizationId: teamInsertResult.meta.last_row_id,
        userId: ownerId,
    });

    return teamInsertResult.meta.last_row_id.toString();
}

async function listMembers(
    con: Drizzle,
    teamId: number,
): Promise<TeamMember[]> {
    const result = await con
        .select({
            id: schema.user.id,
            displayName: schema.user.displayName,
            email: schema.user.email,
        })
        .from(schema.user)
        .innerJoin(
            schema.userToOrganization,
            eq(schema.userToOrganization.userId, schema.user.id),
        )
        .where(eq(schema.userToOrganization.organizationId, teamId))
        .all();

    return result as TeamMember[];
}

async function addMember(
    con: Drizzle,
    teamId: number,
    email: string,
): Promise<void> {
    const exists = await Users.existsByEmail(con, email);

    if (exists === null) {
        throw new Error('User not found');
    }

    await con
        .insert(schema.userToOrganization)
        .values({
            organizationId: teamId,
            userId: exists,
        })
        .execute();
}

async function removeMember(
    con: Drizzle,
    teamId: number,
    userId: number,
): Promise<void> {
    const ownerTeam = await con.query.organization.findFirst({
        columns: {
            ownerId: true,
        },
        where: eq(schema.organization.id, userId),
    });

    if (ownerTeam === undefined) {
        throw new Error('Organization not found');
    }

    if (ownerTeam.ownerId === userId) {
        throw new Error('Cannot remove owner from team');
    }

    await con
        .delete(schema.userToOrganization)
        .where(
            and(
                eq(schema.userToOrganization.organizationId, teamId),
                eq(schema.userToOrganization.userId, userId),
            ),
        )
        .execute();
}

async function deleteById(con: Drizzle, teamId: number): Promise<void> {
    const shops = await con.query.shop.findMany({
        columns: {
            id: true,
        },
        where: eq(schema.shop.organizationId, teamId),
    });

    const shopIds = shops.map((shop) => shop.id);

    // Delete shops associated with team
    await con
        .delete(schema.shop)
        .where(eq(schema.shop.organizationId, teamId))
        .execute();

    if (shopIds.length > 0) {
        await con
            .delete(schema.shopScrapeInfo)
            .where(inArray(schema.shopScrapeInfo.shopId, shopIds))
            .execute();
    }

    // Delete team
    await con
        .delete(schema.organization)
        .where(eq(schema.organization.id, teamId))
        .execute();

    await con
        .delete(schema.userToOrganization)
        .where(eq(schema.userToOrganization.organizationId, teamId))
        .execute();
}

async function update(
    con: Drizzle,
    teamId: number,
    name: string,
): Promise<void> {
    await con
        .update(schema.organization)
        .set({
            name,
        })
        .where(eq(schema.organization.id, teamId))
        .execute();
}

export default {
    create,
    listMembers,
    addMember,
    removeMember,
    deleteById,
    update,
};
