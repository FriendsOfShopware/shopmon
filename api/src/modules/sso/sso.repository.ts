import { and, eq } from 'drizzle-orm';
import { type Drizzle, ssoProvider } from '#src/db.ts';

export interface UpdateSSOProviderInput {
    organizationId: string;
    providerId: string;
    domain: string;
    issuer: string;
    oidcConfig: string;
}

async function listByOrganization(con: Drizzle, organizationId: string) {
    return await con
        .select()
        .from(ssoProvider)
        .where(eq(ssoProvider.organizationId, organizationId));
}

async function findById(
    con: Drizzle,
    organizationId: string,
    providerId: string,
) {
    return await con
        .select()
        .from(ssoProvider)
        .where(
            and(
                eq(ssoProvider.providerId, providerId),
                eq(ssoProvider.organizationId, organizationId),
            ),
        )
        .get();
}

async function update(con: Drizzle, input: UpdateSSOProviderInput) {
    return await con
        .update(ssoProvider)
        .set({
            domain: input.domain,
            issuer: input.issuer,
            oidcConfig: input.oidcConfig,
        })
        .where(
            and(
                eq(ssoProvider.providerId, input.providerId),
                eq(ssoProvider.organizationId, input.organizationId),
            ),
        )
        .returning();
}

async function deleteById(
    con: Drizzle,
    organizationId: string,
    providerId: string,
) {
    return await con
        .delete(ssoProvider)
        .where(
            and(
                eq(ssoProvider.providerId, providerId),
                eq(ssoProvider.organizationId, organizationId),
            ),
        )
        .returning();
}

export default {
    listByOrganization,
    findById,
    update,
    deleteById,
};
