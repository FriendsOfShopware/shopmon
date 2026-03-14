import "#src/sentry.ts";
import { captureException } from "@sentry/bun";
import { auth } from "#src/auth.ts";
import { createContext } from "#src/trpc/context.ts";
import { appRouter } from "#src/trpc/router.ts";
import { fetchRequestHandler } from "@trpc/server/adapters/fetch";
import { createBullBoard } from "@bull-board/api";
import { BullMQAdapter } from "@bull-board/api/bullMQAdapter";
import { HonoAdapter } from "@bull-board/hono";
import { serveStatic } from "hono/bun";
import { Hono } from "hono";
import { getShopQueue, getSitespeedQueue, getMaintenanceQueue } from "#src/modules/queue/queues.ts";

const serverAdapter = new HonoAdapter(serveStatic);
serverAdapter.setBasePath("/admin/queues");

createBullBoard({
  queues: [
    new BullMQAdapter(getShopQueue()),
    new BullMQAdapter(getSitespeedQueue()),
    new BullMQAdapter(getMaintenanceQueue()),
  ],
  serverAdapter,
});

const bullBoardApp = new Hono();
bullBoardApp.route("/admin/queues", serverAdapter.registerPlugin());

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

    if (pathName.startsWith("/admin/queues")) {
      const session = await auth.api.getSession({ headers: request.headers });

      if (!session?.user || session.user.role !== "admin") {
        return new Response(null, { status: 403 });
      }

      return bullBoardApp.fetch(request);
    }

    return new Response(null, { status: 404 });
  },
};
