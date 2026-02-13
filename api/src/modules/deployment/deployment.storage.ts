import { S3Client } from "bun";
import { eq } from "drizzle-orm";
import { type Drizzle, deployment } from "#src/db.ts";

const s3 = new S3Client({
  endpoint: process.env.APP_S3_ENDPOINT,
  accessKeyId: process.env.APP_S3_ACCESS_KEY_ID,
  secretAccessKey: process.env.APP_S3_SECRET_ACCESS_KEY,
  bucket: process.env.APP_S3_BUCKET,
  region: process.env.APP_S3_REGION,
});

function getS3Key(deploymentId: number): string {
  return `deployments/${deploymentId}/output.zst`;
}

export function presignDeploymentOutputUpload(deploymentId: number): string {
  return s3.presign(getS3Key(deploymentId), { method: "PUT", expiresIn: 3600 });
}

export async function getDeploymentOutput(deploymentId: number): Promise<string> {
  const data = await s3.file(getS3Key(deploymentId)).arrayBuffer();
  const decompressed = Bun.zstdDecompressSync(Buffer.from(data));
  return Buffer.from(decompressed).toString();
}

export async function deleteDeploymentOutput(deploymentId: number): Promise<void> {
  await s3.file(getS3Key(deploymentId)).delete();
}

export async function deleteDeploymentOutputsByShopId(db: Drizzle, shopId: number): Promise<void> {
  const deployments = await db
    .select({ id: deployment.id })
    .from(deployment)
    .where(eq(deployment.shopId, shopId));

  await Promise.all(deployments.map((d) => deleteDeploymentOutput(d.id)));
}
