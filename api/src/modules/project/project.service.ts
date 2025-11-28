import { TRPCError } from '@trpc/server';
import type { Drizzle } from '#src/db.ts';
import Shops from '#src/modules/shop/shop.repository.ts';
import Projects from './project.repository.ts';

export interface CreateProjectDTO {
    orgId: string;
    name: string;
    description?: string;
}

export interface UpdateProjectDTO {
    orgId: string;
    projectId: number;
    name?: string;
    description?: string;
}

export interface DeleteProjectDTO {
    orgId: string;
    projectId: number;
}

export const listProjects = async (db: Drizzle, orgId: string) => {
    const result = await Projects.findAllByOrganization(db, orgId);
    return result === undefined ? [] : result;
};

export const createProject = async (db: Drizzle, input: CreateProjectDTO) => {
    return await Projects.create(db, {
        organizationId: input.orgId,
        name: input.name,
        description: input.description,
    });
};

export const updateProject = async (db: Drizzle, input: UpdateProjectDTO) => {
    const project = await Projects.findById(db, input.projectId);

    if (!project || project.organizationId !== input.orgId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Project not found',
        });
    }

    await Projects.update(db, input.projectId, {
        name: input.name,
        description: input.description,
    });

    return true;
};

export const deleteProject = async (db: Drizzle, input: DeleteProjectDTO) => {
    const project = await Projects.findById(db, input.projectId);

    if (!project || project.organizationId !== input.orgId) {
        throw new TRPCError({
            code: 'NOT_FOUND',
            message: 'Project not found',
        });
    }

    const shopCount = await Shops.countByProject(db, input.projectId);

    if (shopCount > 0) {
        throw new TRPCError({
            code: 'BAD_REQUEST',
            message:
                'Cannot delete project with assigned shops. Please reassign or delete the shops first.',
        });
    }

    await Projects.deleteById(db, input.projectId);

    return true;
};
