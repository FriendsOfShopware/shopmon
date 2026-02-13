import { adminRouter } from "#src/modules/admin/admin.router.ts";
import { cliRouter } from "#src/modules/deployment/cli.router.ts";
import { deploymentRouter } from "#src/modules/deployment/deployment.router.ts";
import { infoRouter } from "#src/modules/info/info.router.ts";
import { projectRouter } from "#src/modules/project/project.router.ts";
import { apiKeyRouter } from "#src/modules/project/project-api-key.router.ts";
import { shopRouter } from "#src/modules/shop/shop.router.ts";
import { ssoRouter } from "#src/modules/sso/sso.router.ts";
import { accountRouter } from "#src/modules/user/user.router.ts";
import { publicProcedure, router } from "./index.ts";

export const organizationRouter = router({
  shop: shopRouter,
  sso: ssoRouter,
  project: projectRouter,
  apiKey: apiKeyRouter,
  deployment: deploymentRouter,
});

export const appRouter = router({
  health: publicProcedure.query(() => "ok"),
  account: accountRouter,
  organization: organizationRouter,
  shop: shopRouter,
  admin: adminRouter,
  info: infoRouter,
  cli: cliRouter,
});

export type AppRouter = typeof appRouter;
