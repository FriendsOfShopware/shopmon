import "#src/sentry.ts";
import { captureException } from "@sentry/bun";
import { auth } from "#src/auth.ts";
import { getConnection } from "#src/db.ts";
import {
  deploymentSchema,
  handleDeploymentSubmission,
} from "#src/modules/deployment/deployment.service.ts";
import { createContext } from "#src/trpc/context.ts";
import { appRouter } from "#src/trpc/router.ts";
import { fetchRequestHandler } from "@trpc/server/adapters/fetch";

export default {
  async fetch(request: Request): Promise<Response> {
    const pathName = new URL(request.url).pathname;

    if (pathName.startsWith("/auth")) {
      return auth.handler(request);
    }

    if (pathName.startsWith("/trpc")) {
      return fetchRequestHandler({
        req: request,
        router: appRouter,
        endpoint: "/trpc",
        createContext: createContext(),
        onError: (err) => {
          if (err.error.code === "INTERNAL_SERVER_ERROR") {
            console.error(`[tRPC] Error on path: ${err.path}`, err.error);
            captureException(err.error, {
              user: {
                id: err.ctx?.user?.id,
              },
              extra: {
                type: err.type,
                path: err.path,
              },
            });
          }
        },
      });
    }

    // CLI endpoint for deployment data
    if (pathName === "/api/cli/deployments" && request.method === "POST") {
      try {
        const authHeader = request.headers.get("Authorization");

        if (!authHeader || !authHeader.startsWith("Bearer ")) {
          return Response.json(
            { error: "Missing or invalid authorization header" },
            { status: 401 },
          );
        }

        const token = authHeader.substring(7);
        const body = await request.json();

        // Validate the body
        const validationResult = deploymentSchema.safeParse(body);
        if (!validationResult.success) {
          return Response.json(
            { error: "Invalid request body", details: validationResult.error },
            { status: 400 },
          );
        }

        const db = getConnection();
        const result = await handleDeploymentSubmission(
          db,
          token,
          validationResult.data,
        );

        return Response.json(result);
      } catch (error) {
        if (error instanceof Error && error.message === "Invalid token") {
          return Response.json({ error: "Invalid token" }, { status: 401 });
        }
        captureException(error);
        return Response.json(
          { error: "Internal server error" },
          { status: 500 },
        );
      }
    }

    return new Response(null, { status: 404 });
  },
};
