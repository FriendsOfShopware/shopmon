import { and, eq } from 'drizzle-orm';
import { type ApiKeyScope, type Drizzle, schema } from '#src/db.ts';

export interface CreateApiKeyInput {
    id: string;
    projectId: number;
    name: string;
    token: string;
    scopes: ApiKeyScope[];
}

async function findAllByProject(con: Drizzle, projectId: number) {
    return await con
        .select({
            id: schema.projectApiKey.id,
            name: schema.projectApiKey.name,
            scopes: schema.projectApiKey.scopes,
            createdAt: schema.projectApiKey.createdAt,
            lastUsedAt: schema.projectApiKey.lastUsedAt,
        })
        .from(schema.projectApiKey)
        .where(eq(schema.projectApiKey.projectId, projectId))
        .all();
}

async function create(con: Drizzle, input: CreateApiKeyInput) {
    await con.insert(schema.projectApiKey).values({
        id: input.id,
        projectId: input.projectId,
        name: input.name,
        token: input.token,
        scopes: input.scopes,
        createdAt: new Date(),
    });

    return input.id;
}

async function findById(con: Drizzle, id: string) {
    return await con.query.projectApiKey.findFirst({
        where: eq(schema.projectApiKey.id, id),
    });
}

async function findByToken(con: Drizzle, token: string) {
    return await con.query.projectApiKey.findFirst({
        where: eq(schema.projectApiKey.token, token),
        with: {
            project: true,
        },
    });
}

async function deleteById(con: Drizzle, id: string) {
    await con
        .delete(schema.projectApiKey)
        .where(eq(schema.projectApiKey.id, id))
        .execute();
}

async function deleteByProjectAndId(
    con: Drizzle,
    projectId: number,
    id: string,
) {
    await con
        .delete(schema.projectApiKey)
        .where(
            and(
                eq(schema.projectApiKey.id, id),
                eq(schema.projectApiKey.projectId, projectId),
            ),
        )
        .execute();
}

async function updateLastUsedAt(con: Drizzle, id: string) {
    await con
        .update(schema.projectApiKey)
        .set({ lastUsedAt: new Date() })
        .where(eq(schema.projectApiKey.id, id))
        .execute();
}

export default {
    findAllByProject,
    create,
    findById,
    findByToken,
    deleteById,
    deleteByProjectAndId,
    updateLastUsedAt,
};
