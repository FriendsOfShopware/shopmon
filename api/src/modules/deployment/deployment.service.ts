import { eq } from "drizzle-orm";
import { z } from "zod";
import { deployment, deploymentToken, organization, shop } from "#src/db.ts";
import type { Drizzle } from "#src/db.ts";

export const deploymentSchema = z.object({
  command: z.string(),
  output: z.string(),
  return_code: z.number(),
  start_date: z.string(),
  end_date: z.string(),
  execution_time: z.number(),
  composer: z.record(z.string(), z.string()).optional(),
  reference: z.string().optional(),
});

export type DeploymentInput = z.infer<typeof deploymentSchema>;

// Simple random name generator
function generateRandomName(): string {
  const adjectives = [
    "happy",
    "clever",
    "brave",
    "swift",
    "calm",
    "bright",
    "kind",
    "wise",
  ];
  const nouns = [
    "panda",
    "tiger",
    "eagle",
    "dolphin",
    "wolf",
    "falcon",
    "bear",
    "fox",
  ];
  const adj = adjectives[Math.floor(Math.random() * adjectives.length)];
  const noun = nouns[Math.floor(Math.random() * nouns.length)];
  return `${adj}-${noun}`;
}

export async function handleDeploymentSubmission(
  db: Drizzle,
  token: string,
  body: DeploymentInput,
) {
  // Find the deployment token
  const tokenResult = await db
    .select()
    .from(deploymentToken)
    .where(eq(deploymentToken.token, token));

  const deploymentTokenRecord = tokenResult[0];

  if (!deploymentTokenRecord) {
    throw new Error("Invalid token");
  }

  const name = generateRandomName();

  // Create the deployment record
  const result = await db
    .insert(deployment)
    .values({
      shopId: deploymentTokenRecord.shopId,
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
    .returning({ id: deployment.id });

  const deploymentId = result[0].id;

  // Update last used timestamp on token
  await db
    .update(deploymentToken)
    .set({ lastUsedAt: new Date() })
    .where(eq(deploymentToken.id, deploymentTokenRecord.id));

  // Get shop information to build the URL
  const shopResult = await db
    .select({
      organizationSlug: organization.slug,
    })
    .from(shop)
    .innerJoin(organization, eq(shop.organizationId, organization.id))
    .where(eq(shop.id, deploymentTokenRecord.shopId));

  const shopData = shopResult[0];

  const frontendUrl = process.env.FRONTEND_URL || "http://localhost:3000";
  const deploymentUrl = `${frontendUrl}/app/organizations/${shopData?.organizationSlug}/shops/${deploymentTokenRecord.shopId}/deployments/${deploymentId}`;

  return {
    success: true,
    name,
    deployment_id: deploymentId,
    url: deploymentUrl,
  };
}
