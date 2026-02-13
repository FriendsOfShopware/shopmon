import { TRPCError } from "@trpc/server";
import { eq } from "drizzle-orm";
import { z } from "zod";
import { deployment, deploymentToken, organization, shop } from "#src/db.ts";
import { generateRandomName } from "#src/helpers/nameGenerator.ts";
import { publicProcedure, router } from "#src/trpc/index.ts";
import { presignDeploymentOutputUpload } from "./deployment.storage.ts";

const deploymentSchema = z.object({
  command: z.string(),
  return_code: z.number(),
  start_date: z.string(),
  end_date: z.string(),
  execution_time: z.number(),
  composer: z.record(z.string(), z.string()).optional(),
  reference: z.string().optional(),
});

export const cliRouter = router({
  createDeployment: publicProcedure.input(deploymentSchema).mutation(async ({ ctx, input }) => {
    const authHeader = ctx.headers.get("Authorization");

    if (!authHeader || !authHeader.startsWith("Bearer ")) {
      throw new TRPCError({
        code: "UNAUTHORIZED",
        message: "Missing or invalid authorization header",
      });
    }

    const token = authHeader.substring(7);

    const tokenResult = await ctx.drizzle
      .select()
      .from(deploymentToken)
      .where(eq(deploymentToken.token, token));

    const deploymentTokenRecord = tokenResult[0];

    if (!deploymentTokenRecord) {
      throw new TRPCError({
        code: "UNAUTHORIZED",
        message: "Invalid token",
      });
    }

    const name = generateRandomName();

    const result = await ctx.drizzle
      .insert(deployment)
      .values({
        shopId: deploymentTokenRecord.shopId,
        name,
        command: input.command,
        returnCode: input.return_code,
        startDate: new Date(input.start_date),
        endDate: new Date(input.end_date),
        executionTime: input.execution_time.toString(),
        composer: (input.composer || {}) as Record<string, string>,
        reference: input.reference,
        createdAt: new Date(),
      })
      .returning({ id: deployment.id });

    const deploymentId = result[0].id;

    const upload_url = presignDeploymentOutputUpload(deploymentId);

    await ctx.drizzle
      .update(deploymentToken)
      .set({ lastUsedAt: new Date() })
      .where(eq(deploymentToken.id, deploymentTokenRecord.id));

    const shopResult = await ctx.drizzle
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
      upload_url,
    };
  }),
});
