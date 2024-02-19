import { Drizzle, getLastInsertId, schema } from "../db";
import { eq, and, inArray } from "drizzle-orm";
import Users from "./users";

interface OrganizationMember {
	id: number;
	email: string;
}

async function create(
	con: Drizzle,
	name: string,
	ownerId: number,
): Promise<string> {
	const organizationInsertResult = await con
		.insert(schema.organization)
		.values({
			name,
			ownerId: ownerId,
			createdAt: new Date(),
		})
		.execute();

	const lastId = getLastInsertId(organizationInsertResult);

	await con.insert(schema.userToOrganization).values({
		organizationId: lastId,
		userId: ownerId,
	});

	return lastId.toString();
}

async function listMembers(
	con: Drizzle,
	organizationId: number,
): Promise<OrganizationMember[]> {
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
		.where(eq(schema.userToOrganization.organizationId, organizationId))
		.all();

	return result as OrganizationMember[];
}

async function addMember(
	con: Drizzle,
	organizationId: number,
	email: string,
): Promise<void> {
	const exists = await Users.existsByEmail(con, email);

	if (exists === null) {
		throw new Error("User not found");
	}

	await con
		.insert(schema.userToOrganization)
		.values({
			organizationId,
			userId: exists,
		})
		.execute();
}

async function removeMember(
	con: Drizzle,
	organizationId: number,
	userId: number,
): Promise<void> {
	const ownerOrganization = await con.query.organization.findFirst({
		columns: {
			ownerId: true,
		},
		where: eq(schema.organization.id, userId),
	});

	if (ownerOrganization === undefined) {
		throw new Error("Organization not found");
	}

	if (ownerOrganization.ownerId === userId) {
		throw new Error("Cannot remove owner from organization");
	}

	await con
		.delete(schema.userToOrganization)
		.where(
			and(
				eq(schema.userToOrganization.organizationId, organizationId),
				eq(schema.userToOrganization.userId, userId),
			),
		)
		.execute();
}

async function deleteById(con: Drizzle, organizationId: number): Promise<void> {
	const shops = await con.query.shop.findMany({
		columns: {
			id: true,
		},
		where: eq(schema.shop.organizationId, organizationId),
	});

	const shopIds = shops.map((shop) => shop.id);

	// Delete shops associated with organization
	await con
		.delete(schema.shop)
		.where(eq(schema.shop.organizationId, organizationId))
		.execute();

	if (shopIds.length > 0) {
		await con
			.delete(schema.shopScrapeInfo)
			.where(inArray(schema.shopScrapeInfo.shopId, shopIds))
			.execute();
	}

	// Delete organization
	await con
		.delete(schema.userToOrganization)
		.where(eq(schema.userToOrganization.organizationId, organizationId))
		.execute();

	await con
		.delete(schema.organization)
		.where(eq(schema.organization.id, organizationId))
		.execute();
}

async function update(
	con: Drizzle,
	organizationId: number,
	name: string,
): Promise<void> {
	await con
		.update(schema.organization)
		.set({
			name,
		})
		.where(eq(schema.organization.id, organizationId))
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
