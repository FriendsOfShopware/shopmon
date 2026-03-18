import { TRPCError } from "@trpc/server";
import type { Drizzle } from "#src/db.ts";
import Projects from "./project.repository.ts";

interface PackagesTokenResponse {
  id: number;
  source: string;
  lastSyncedAt: number | null;
}

function getConfig() {
  const apiUrl = process.env.PACKAGES_API_URL;
  const apiToken = process.env.PACKAGES_API_TOKEN;

  if (!apiUrl || !apiToken) {
    throw new TRPCError({
      code: "PRECONDITION_FAILED",
      message: "Packages mirror is not configured",
    });
  }

  return { apiUrl, apiToken };
}

function projectSource(projectId: number): string {
  return `shopmon-project-${projectId}`;
}

async function assertProjectAccess(db: Drizzle, orgId: string, projectId: number) {
  const project = await Projects.findById(db, projectId);

  if (!project || project.organizationId !== orgId) {
    throw new TRPCError({
      code: "NOT_FOUND",
      message: "Project not found",
    });
  }

  return project;
}

export async function listTokens(
  db: Drizzle,
  input: { orgId: string; projectId: number },
): Promise<PackagesTokenResponse[]> {
  await assertProjectAccess(db, input.orgId, input.projectId);

  const { apiUrl, apiToken } = getConfig();
  const source = projectSource(input.projectId);

  const response = await fetch(`${apiUrl}/api/tokens?source=${encodeURIComponent(source)}`, {
    headers: { Authorization: `Bearer ${apiToken}` },
  });

  if (!response.ok) {
    throw new TRPCError({
      code: "INTERNAL_SERVER_ERROR",
      message: "Failed to fetch packages tokens",
    });
  }

  return (await response.json()) as PackagesTokenResponse[];
}

async function validateShopwareToken(token: string): Promise<void> {
  const response = await fetch("https://packages.shopware.com/packages.json", {
    headers: { Authorization: `Bearer ${token}` },
  });

  if (!response.ok) {
    throw new TRPCError({
      code: "BAD_REQUEST",
      message: "Invalid Shopware store token",
    });
  }
}

export async function createToken(
  db: Drizzle,
  input: { orgId: string; projectId: number; token: string },
): Promise<PackagesTokenResponse> {
  await assertProjectAccess(db, input.orgId, input.projectId);

  await validateShopwareToken(input.token);

  const { apiUrl, apiToken } = getConfig();
  const source = projectSource(input.projectId);

  const response = await fetch(`${apiUrl}/api/tokens`, {
    method: "POST",
    headers: {
      Authorization: `Bearer ${apiToken}`,
      "Content-Type": "application/json",
    },
    body: JSON.stringify({ token: input.token, source }),
  });

  if (!response.ok) {
    throw new TRPCError({
      code: "INTERNAL_SERVER_ERROR",
      message: "Failed to create packages token",
    });
  }

  return (await response.json()) as PackagesTokenResponse;
}

export async function deleteToken(
  db: Drizzle,
  input: { orgId: string; projectId: number; tokenId: number },
): Promise<void> {
  await assertProjectAccess(db, input.orgId, input.projectId);

  const { apiUrl, apiToken } = getConfig();

  // Verify the token belongs to this project by listing and checking
  const tokens = await listTokens(db, {
    orgId: input.orgId,
    projectId: input.projectId,
  });

  const tokenExists = tokens.some((t) => t.id === input.tokenId);
  if (!tokenExists) {
    throw new TRPCError({
      code: "NOT_FOUND",
      message: "Packages token not found for this project",
    });
  }

  const response = await fetch(`${apiUrl}/api/tokens/${input.tokenId}`, {
    method: "DELETE",
    headers: { Authorization: `Bearer ${apiToken}` },
  });

  if (!response.ok) {
    throw new TRPCError({
      code: "INTERNAL_SERVER_ERROR",
      message: "Failed to delete packages token",
    });
  }
}

export async function syncToken(
  db: Drizzle,
  input: { orgId: string; projectId: number; tokenId: number },
): Promise<void> {
  await assertProjectAccess(db, input.orgId, input.projectId);

  const { apiUrl, apiToken } = getConfig();

  // Verify the token belongs to this project
  const tokens = await listTokens(db, {
    orgId: input.orgId,
    projectId: input.projectId,
  });

  const tokenExists = tokens.some((t) => t.id === input.tokenId);
  if (!tokenExists) {
    throw new TRPCError({
      code: "NOT_FOUND",
      message: "Packages token not found for this project",
    });
  }

  const response = await fetch(`${apiUrl}/api/tokens/${input.tokenId}/sync`, {
    method: "POST",
    headers: { Authorization: `Bearer ${apiToken}` },
  });

  if (!response.ok) {
    throw new TRPCError({
      code: "INTERNAL_SERVER_ERROR",
      message: "Failed to sync packages token",
    });
  }
}

export function getConfiguration(): { configured: boolean; composerUrl: string | null } {
  const apiUrl = process.env.PACKAGES_API_URL;
  const configured = !!(apiUrl && process.env.PACKAGES_API_TOKEN);

  return {
    configured,
    composerUrl: configured ? apiUrl! : null,
  };
}
