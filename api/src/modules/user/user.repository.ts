import { TRPCError } from "@trpc/server";
import { and, desc, eq, inArray, sql } from "drizzle-orm";
import { type Drizzle, getConnection, schema } from "#src/db.ts";

async function existsByEmail(con: Drizzle, email: string): Promise<string | null> {
  const result = await con.query.user.findFirst({
    columns: {
      id: true,
    },
    where: eq(schema.user.email, email),
  });

  return result ? result.id : null;
}

async function existsById(con: Drizzle, id: string): Promise<boolean> {
  const result = await con.query.user.findFirst({
    columns: {
      id: true,
    },
    where: eq(schema.user.id, id),
  });

  return result !== undefined;
}

async function findById(con: Drizzle, id: string) {
  const result = await con
    .select({
      id: schema.user.id,
      displayName: schema.user.name,
      email: schema.user.email,
      createdAt: schema.user.createdAt,
    })
    .from(schema.user)
    .where(eq(schema.user.id, id));
  return result[0];
}

async function deleteById(con: Drizzle, id: string): Promise<void> {
  const orgIds = await con
    .select({
      id: schema.member.id,
    })
    .from(schema.member)
    .where(eq(schema.member.userId, id))
    .execute();

  if (orgIds.length > 0) {
    throw new TRPCError({
      code: "CONFLICT",
      message: "To delete an user, you must leave or delete all organizations first.",
    });
  }

  await con.delete(schema.userNotification).where(eq(schema.userNotification.userId, id)).execute();
}

async function createNotification(
  con: Drizzle,
  userId: string,
  key: string,
  notification: Omit<typeof schema.userNotification.$inferInsert, "createdAt" | "key" | "userId">,
) {
  const result = await con
    .insert(schema.userNotification)
    .values({
      userId: userId,
      key,
      level: notification.level,
      title: notification.title,
      message: notification.message,
      link: notification.link,
      createdAt: new Date(),
    })
    .onConflictDoUpdate({
      target: [schema.userNotification.userId, schema.userNotification.key],
      set: {
        read: false,
      },
    })
    .returning();

  return result[0];
}

async function hasAccessToProject(userId: string, projectId: number) {
  const result = await getConnection()
    .select({
      id: schema.project.id,
      organizationId: schema.project.organizationId,
    })
    .from(schema.project)
    .innerJoin(schema.organization, eq(schema.project.organizationId, schema.organization.id))
    .innerJoin(schema.member, eq(schema.organization.id, schema.member.organizationId))
    .where(and(eq(schema.project.id, projectId), eq(schema.member.userId, userId)));
  return result[0];
}

async function hasAccessToShop(userId: string, shopId: number): Promise<boolean> {
  const results = await getConnection()
    .select({
      id: schema.shop.id,
    })
    .from(schema.shop)
    .innerJoin(schema.member, eq(schema.shop.organizationId, schema.member.organizationId))
    .where(and(eq(schema.shop.id, shopId), eq(schema.member.userId, userId)));

  return results[0] !== undefined;
}

async function findOrganizations(con: Drizzle, userId: string) {
  return con
    .select({
      id: schema.organization.id,
      name: schema.organization.name,
      createdAt: schema.organization.createdAt,
      slug: schema.organization.slug,
      shopCount: sql<number>`(SELECT COUNT(1) FROM ${schema.shop} WHERE ${schema.shop.organizationId} = ${schema.organization.id})`,
      memberCount: sql<number>`(SELECT COUNT(1) FROM ${schema.member} WHERE ${schema.member.organizationId} = ${schema.organization.id})`,
    })
    .from(schema.organization)
    .innerJoin(schema.member, eq(schema.member.organizationId, schema.organization.id))
    .where(eq(schema.member.userId, userId));
}

async function findShops(con: Drizzle, userId: string) {
  return await con
    .select({
      id: schema.shop.id,
      name: schema.shop.name,
      nameCombined: sql<string>`CONCAT(${schema.project.name}, ' / ', ${schema.shop.name})`,
      status: schema.shop.status,
      url: schema.shop.url,
      favicon: schema.shop.favicon,
      createdAt: schema.shop.createdAt,
      lastScrapedAt: schema.shop.lastScrapedAt,
      shopwareVersion: schema.shop.shopwareVersion,
      lastChangelog: schema.shop.lastChangelog,
      organizationId: schema.shop.organizationId,
      organizationName: sql<string>`${schema.organization.name}`.as("organization_name"),
      organizationSlug: sql<string>`${schema.organization.slug}`.as("organization_slug"),
      projectId: schema.shop.projectId,
      projectName: sql<string>`${schema.project.name}`.as("project_name"),
    })
    .from(schema.shop)
    .innerJoin(schema.member, eq(schema.member.organizationId, schema.shop.organizationId))
    .innerJoin(schema.organization, eq(schema.organization.id, schema.shop.organizationId))
    .leftJoin(schema.project, eq(schema.project.id, schema.shop.projectId))
    .where(eq(schema.member.userId, userId))
    .orderBy(schema.shop.name);
}

async function findShopsSimple(con: Drizzle, userId: string) {
  return await con
    .select({
      id: schema.shop.id,
      name: schema.shop.name,
      organizationId: schema.shop.organizationId,
      shopwareVersion: schema.shop.shopwareVersion,
      organizationSlug: sql<string>`${schema.organization.slug}`.as("organization_slug"),
    })
    .from(schema.shop)
    .innerJoin(schema.member, eq(schema.member.organizationId, schema.shop.organizationId))
    .innerJoin(schema.organization, eq(schema.organization.id, schema.shop.organizationId))
    .where(eq(schema.member.userId, userId))
    .orderBy(schema.shop.name);
}

async function findProjects(con: Drizzle, userId: string) {
  return await con
    .select({
      id: schema.project.id,
      name: schema.project.name,
      nameCombined: sql<string>`CONCAT(${schema.organization.name}, ' / ', ${schema.project.name})`,
      organizationId: schema.project.organizationId,
      organizationSlug: schema.organization.slug,
      description: schema.project.description,
      createdAt: schema.project.createdAt,
    })
    .from(schema.project)
    .innerJoin(schema.organization, eq(schema.organization.id, schema.project.organizationId))
    .innerJoin(
      schema.member,
      and(
        eq(schema.member.organizationId, schema.organization.id),
        eq(schema.member.userId, userId),
      ),
    );
}

async function findChangelogs(con: Drizzle, userId: string) {
  return await con
    .select({
      id: schema.shopChangelog.id,
      shopId: schema.shopChangelog.shopId,
      shopOrganizationId: schema.shop.organizationId,
      organizationSlug: sql<string>`${schema.organization.slug}`.as("organization_slug"),
      shopName: schema.shop.name,
      shopFavicon: schema.shop.favicon,
      extensions: schema.shopChangelog.extensions,
      oldShopwareVersion: schema.shopChangelog.oldShopwareVersion,
      newShopwareVersion: schema.shopChangelog.newShopwareVersion,
      date: schema.shopChangelog.date,
    })
    .from(schema.shop)
    .innerJoin(schema.member, eq(schema.member.organizationId, schema.shop.organizationId))
    .innerJoin(schema.organization, eq(schema.organization.id, schema.shop.organizationId))
    .innerJoin(schema.shopChangelog, eq(schema.shopChangelog.shopId, schema.shop.id))
    .where(eq(schema.member.userId, userId))
    .orderBy(desc(schema.shopChangelog.date))
    .limit(10);
}

async function findSubscribedShops(con: Drizzle, userId: string, shopIds: number[]) {
  return await con
    .select({
      id: schema.shop.id,
      name: schema.shop.name,
      url: schema.shop.url,
      favicon: schema.shop.favicon,
      shopwareVersion: schema.shop.shopwareVersion,
      organizationId: schema.shop.organizationId,
      organizationName: sql<string>`${schema.organization.name}`.as("organization_name"),
      organizationSlug: sql<string>`${schema.organization.slug}`.as("organization_slug"),
    })
    .from(schema.shop)
    .innerJoin(schema.organization, eq(schema.organization.id, schema.shop.organizationId))
    .innerJoin(schema.member, eq(schema.member.organizationId, schema.shop.organizationId))
    .where(and(inArray(schema.shop.id, shopIds), eq(schema.member.userId, userId)))
    .orderBy(schema.shop.name);
}

export default {
  existsByEmail,
  existsById,
  findById,
  deleteById,
  createNotification,
  hasAccessToProject,
  hasAccessToShop,
  findOrganizations,
  findShops,
  findShopsSimple,
  findProjects,
  findChangelogs,
  findSubscribedShops,
};
