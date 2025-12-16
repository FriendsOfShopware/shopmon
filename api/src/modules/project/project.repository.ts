import { eq } from 'drizzle-orm';
import { type Drizzle, schema } from '#src/db.ts';

export interface CreateProjectInput {
    organizationId: string;
    name: string;
    description?: string;
}

export interface UpdateProjectInput {
    name?: string;
    description?: string;
}

async function findAllByOrganization(con: Drizzle, organizationId: string) {
    return await con
        .select({
            id: schema.project.id,
            name: schema.project.name,
            description: schema.project.description,
            createdAt: schema.project.createdAt,
            updatedAt: schema.project.updatedAt,
            organizationId: schema.project.organizationId,
        })
        .from(schema.project)
        .where(eq(schema.project.organizationId, organizationId))
        ;
}

async function create(con: Drizzle, input: CreateProjectInput) {
    const result = await con
        .insert(schema.project)
        .values({
            organizationId: input.organizationId,
            name: input.name,
            description: input.description,
            createdAt: new Date(),
            updatedAt: new Date(),
        })
        .returning({ id: schema.project.id });

    return result[0].id;
}

async function findById(con: Drizzle, id: number) {
    return await con.query.project.findFirst({
        where: eq(schema.project.id, id),
    });
}

async function update(con: Drizzle, id: number, input: UpdateProjectInput) {
    const updateData: {
        name?: string;
        description?: string;
        updatedAt: Date;
    } = { updatedAt: new Date() };

    if (input.name !== undefined) {
        updateData.name = input.name;
    }
    if (input.description !== undefined) {
        updateData.description = input.description;
    }

    await con
        .update(schema.project)
        .set(updateData)
        .where(eq(schema.project.id, id))
        .execute();
}

async function deleteById(con: Drizzle, id: number) {
    await con.delete(schema.project).where(eq(schema.project.id, id)).execute();
}

export default {
    findAllByOrganization,
    create,
    findById,
    update,
    deleteById,
};
