import { adminRouter } from '#src/modules/admin/admin.router.ts';
import { infoRouter } from '#src/modules/info/info.router.ts';
import { projectRouter } from '#src/modules/project/project.router.ts';
import { shopRouter } from '#src/modules/shop/shop.router.ts';
import { ssoRouter } from '#src/modules/sso/sso.router.ts';
import { accountRouter } from '#src/modules/user/user.router.ts';
import { publicProcedure, router } from './index.ts';

export const organizationRouter = router({
    shop: shopRouter,
    sso: ssoRouter,
    project: projectRouter,
});

export const appRouter = router({
    health: publicProcedure.query(() => 'ok'),
    account: accountRouter,
    organization: organizationRouter,
    shop: shopRouter,
    admin: adminRouter,
    info: infoRouter,
});

export type AppRouter = typeof appRouter;
