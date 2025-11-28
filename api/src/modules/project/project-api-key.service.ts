import { TRPCError } from '@trpc/server';
import type { ApiKeyScope, Drizzle } from '#src/db.ts';
import Projects from './project.repository.ts';
import ApiKeys from './project-api-key.repository.ts';

export const API_KEY_SCOPES = ['deployments'] as const satisfies ApiKeyScope[];

export interface CreateApiKeyDTO {
    orgId: string;
    projectId: number;
    name: string;
    scopes: ApiKeyScope[];
}

export interface DeleteApiKeyDTO {
    orgId: string;
    projectId: number;
    apiKeyId: string;
}

export interface ListApiKeysDTO {
    orgId: string;
    projectId: number;
}

function generateApiKey(): string {
    const prefix = 'shm_';
    const randomBytes = crypto.getRandomValues(new Uint8Array(32));
    const base64 = btoa(String.fromCharCode(...randomBytes))
        .replace(/\+/g, '-')
        .replace(/\//g, '_')
        .replace(/=/g, '');
    return prefix + base64;
}

function generateId(): string {
    return crypto.randomUUID();
}

export const listApiKeys = async (db: Drizzle, input: ListApiKeysDTO) => {
    const project = await Projects.findById(db, input.projectId);

    if (!project || project.organizationId !== input.orgId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Project not found',
        });
    }

    return await ApiKeys.findAllByProject(db, input.projectId);
};

export const createApiKey = async (db: Drizzle, input: CreateApiKeyDTO) => {
    const project = await Projects.findById(db, input.projectId);

    if (!project || project.organizationId !== input.orgId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Project not found',
        });
    }

    const invalidScopes = input.scopes.filter(
        (scope) => !API_KEY_SCOPES.includes(scope),
    );
    if (invalidScopes.length > 0) {
        throw new TRPCError({
            code: 'BAD_REQUEST',
            message: `Invalid scopes: ${invalidScopes.join(', ')}`,
        });
    }

    if (input.scopes.length === 0) {
        throw new TRPCError({
            code: 'BAD_REQUEST',
            message: 'At least one scope is required',
        });
    }

    const token = generateApiKey();
    const id = generateId();

    await ApiKeys.create(db, {
        id,
        projectId: input.projectId,
        name: input.name,
        token,
        scopes: input.scopes,
    });

    return {
        id,
        token,
        name: input.name,
        scopes: input.scopes,
    };
};

export const deleteApiKey = async (db: Drizzle, input: DeleteApiKeyDTO) => {
    const project = await Projects.findById(db, input.projectId);

    if (!project || project.organizationId !== input.orgId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Project not found',
        });
    }

    const apiKey = await ApiKeys.findById(db, input.apiKeyId);

    if (!apiKey || apiKey.projectId !== input.projectId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'API key not found',
        });
    }

    await ApiKeys.deleteByProjectAndId(db, input.projectId, input.apiKeyId);

    return true;
};

export const getAvailableScopes = () => {
    return API_KEY_SCOPES.map((scope) => ({
        value: scope,
        label: getScopeLabel(scope),
        description: getScopeDescription(scope),
    }));
};

function getScopeLabel(scope: ApiKeyScope): string {
    switch (scope) {
        case 'deployments':
            return 'Deployments';
        default:
            return scope;
    }
}

function getScopeDescription(scope: ApiKeyScope): string {
    switch (scope) {
        case 'deployments':
            return 'Create and view deployment records for shops in this project';
        default:
            return '';
    }
}
