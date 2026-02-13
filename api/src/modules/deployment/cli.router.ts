import { TRPCError } from "@trpc/server";
import { eq } from "drizzle-orm";
import { z } from "zod";
import { deployment, productToken, organization, shop } from "#src/db.ts";
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
      .from(productToken)
      .where(eq(productToken.token, token));

    const productTokenRecord = tokenResult[0];

    if (!productTokenRecord) {
      throw new TRPCError({
        code: "UNAUTHORIZED",
        message: "Invalid token",
      });
    }

    if (productTokenRecord.scope !== "deployment") {
      throw new TRPCError({
        code: "FORBIDDEN",
        message: "Token does not have deployment scope",
      });
    }

    const name = generateRandomName();

    const result = await ctx.drizzle
      .insert(deployment)
      .values({
        shopId: productTokenRecord.shopId,
        name,
        command: input.command,
        returnCode: input.return_code,
        startDate: new Date(input.start_date),
        endDate: new Date(input.end_date),
        executionTime: input.execution_time,
        composer: (input.composer || {}) as Record<string, string>,
        reference: input.reference,
        createdAt: new Date(),
      })
      .returning({ id: deployment.id });

    const deploymentId = result[0].id;

    const upload_url = presignDeploymentOutputUpload(deploymentId);

    await ctx.drizzle
      .update(productToken)
      .set({ lastUsedAt: new Date() })
      .where(eq(productToken.id, productTokenRecord.id));

    const shopResult = await ctx.drizzle
      .select({
        organizationSlug: organization.slug,
      })
      .from(shop)
      .innerJoin(organization, eq(shop.organizationId, organization.id))
      .where(eq(shop.id, productTokenRecord.shopId));

    const shopData = shopResult[0];

    const frontendUrl = process.env.FRONTEND_URL || "http://localhost:3000";
    const deploymentUrl = `${frontendUrl}/app/organizations/${shopData?.organizationSlug}/shops/${productTokenRecord.shopId}/deployments/${deploymentId}`;

    return {
      success: true,
      name,
      deployment_id: deploymentId,
      url: deploymentUrl,
      upload_url,
    };
  }),
});
