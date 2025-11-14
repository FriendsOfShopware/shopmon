import { zValidator } from '@hono/zod-validator';
import { eq } from 'drizzle-orm';
import type { Context } from 'hono';
import { z } from 'zod';
import { getConnection, schema } from './db.ts';
import { generateRandomName } from './helpers/nameGenerator.ts';

const deploymentSchema = z.object({
    command: z.string(),
    output: z.string(),
    return_code: z.number(),
    start_date: z.string(),
    end_date: z.string(),
    execution_time: z.number(),
    composer: z.record(z.string(), z.string()).optional(),
    reference: z.string().optional(),
});

type DeploymentInput = z.infer<typeof deploymentSchema>;

export const deploymentValidator = zValidator('json', deploymentSchema);

export async function handleDeploymentSubmission(c: Context) {
    const authHeader = c.req.header('Authorization');

    if (!authHeader || !authHeader.startsWith('Bearer ')) {
        return c.json(
            { error: 'Missing or invalid authorization header' },
            401,
        );
    }

    const token = authHeader.substring(7);
    const db = getConnection();

    // Find the deployment token
    const deploymentToken = await db
        .select()
        .from(schema.deploymentToken)
        .where(eq(schema.deploymentToken.token, token))
        .get();

    if (!deploymentToken) {
        return c.json({ error: 'Invalid token' }, 401);
    }

    // Get validated body
    const body = (await c.req.json()) as DeploymentInput;

    const name = generateRandomName();

    // Create the deployment record
    const result = await db
        .insert(schema.deployment)
        .values({
            shopId: deploymentToken.shopId,
            name,
            command: body.command,
            output: body.output,
            returnCode: body.return_code,
            startDate: new Date(body.start_date),
            endDate: new Date(body.end_date),
            executionTime: body.execution_time.toString(),
            composer: (body.composer || {}) as Record<string, string>,
            reference: body.reference,
            createdAt: new Date(),
        })
        .returning({ id: schema.deployment.id });

    const deploymentId = result[0].id;

    // Update last used timestamp on token
    await db
        .update(schema.deploymentToken)
        .set({ lastUsedAt: new Date() })
        .where(eq(schema.deploymentToken.id, deploymentToken.id));

    // Get shop information to build the URL
    const shop = await db
        .select({
            organizationSlug: schema.organization.slug,
        })
        .from(schema.shop)
        .innerJoin(
            schema.organization,
            eq(schema.shop.organizationId, schema.organization.id),
        )
        .where(eq(schema.shop.id, deploymentToken.shopId))
        .get();

    const frontendUrl = process.env.FRONTEND_URL || 'http://localhost:3000';
    const deploymentUrl = `${frontendUrl}/app/organizations/${shop?.organizationSlug}/shops/${deploymentToken.shopId}/deployments/${deploymentId}`;

    return c.json({
        success: true,
        name,
        deployment_id: deploymentId,
        url: deploymentUrl,
    });
}
