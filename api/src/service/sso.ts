import { TRPCError } from '@trpc/server';
import type { Drizzle } from '#src/db.ts';
import SSORepository from '#src/repository/sso.ts';

export interface UpdateSSOProviderDTO {
    orgId: string;
    providerId: string;
    domain: string;
    issuer: string;
    clientId: string;
    clientSecret?: string;
    authorizationEndpoint: string;
    tokenEndpoint: string;
    jwksEndpoint: string;
}

export const listProviders = async (db: Drizzle, orgId: string) => {
    const providers = await SSORepository.listByOrganization(db, orgId);

    return providers.map((provider) => {
        if (provider.oidcConfig) {
            const oidcConfig = JSON.parse(provider.oidcConfig);
            const { ...safeConfig } = oidcConfig;
            return {
                ...provider,
                oidcConfig: safeConfig,
            };
        }
        return provider;
    });
};

export const discoverOpenIdConfig = async (issuer: string) => {
    try {
        const discoveryUrl = `${issuer}/.well-known/openid-configuration`;
        const response = await fetch(discoveryUrl);

        if (!response.ok) {
            throw new TRPCError({
                code: 'BAD_REQUEST',
                message: `Failed to fetch OpenID configuration from ${discoveryUrl}`,
            });
        }

        const config = (await response.json()) as Record<string, unknown>;

        return {
            issuer: (config.issuer as string) || issuer,
            authorizationEndpoint: config.authorization_endpoint as string,
            tokenEndpoint: config.token_endpoint as string,
            jwksEndpoint: config.jwks_uri as string,
            userInfoEndpoint: config.userinfo_endpoint as string,
            scopes: (config.scopes_supported as string[]) || [
                'openid',
                'profile',
                'email',
            ],
        };
    } catch (error) {
        if (error instanceof TRPCError) {
            throw error;
        }

        throw new TRPCError({
            code: 'INTERNAL_SERVER_ERROR',
            message:
                error instanceof Error
                    ? error.message
                    : 'Failed to discover OpenID configuration',
        });
    }
};

export const updateProvider = async (
    db: Drizzle,
    input: UpdateSSOProviderDTO,
) => {
    const existingProvider = await SSORepository.findById(
        db,
        input.orgId,
        input.providerId,
    );

    if (!existingProvider) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'SSO provider not found',
        });
    }

    let existingOidcConfig: Record<string, unknown> = {};
    if (existingProvider.oidcConfig) {
        existingOidcConfig = JSON.parse(existingProvider.oidcConfig);
    }

    const updatedOidcConfig: Record<string, unknown> = {
        ...existingOidcConfig,
        clientId: input.clientId,
        authorizationEndpoint: input.authorizationEndpoint,
        tokenEndpoint: input.tokenEndpoint,
        jwksEndpoint: input.jwksEndpoint,
    };

    if (input.clientSecret) {
        updatedOidcConfig.clientSecret = input.clientSecret;
    }

    const oidcConfig = JSON.stringify(updatedOidcConfig);

    await SSORepository.update(db, {
        organizationId: input.orgId,
        providerId: input.providerId,
        domain: input.domain,
        issuer: input.issuer,
        oidcConfig,
    });

    return true;
};

export const deleteProvider = async (
    db: Drizzle,
    orgId: string,
    providerId: string,
) => {
    const deleted = await SSORepository.deleteById(db, orgId, providerId);

    if (deleted.length === 0) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'SSO provider not found',
        });
    }

    return { success: true };
};
