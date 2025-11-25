import { shopRouter } from '#src/modules/shop/shop.router.ts';
import { accountRouter } from './account/index.ts';
import { adminRouter } from './admin/index.ts';
import { publicProcedure, router } from './index.ts';
import { infoRouter } from './info/index.ts';
import { organizationRouter } from './organization/index.ts';

export const appRouter = router({
    health: publicProcedure.query(() => 'ok'),
    account: accountRouter,
    organization: organizationRouter,
    shop: shopRouter,
    admin: adminRouter,
    info: infoRouter,
});

export type AppRouter = typeof appRouter;
