import { router } from './index.ts';
import { accountRouter } from './account/index.ts';
import { authRouter } from './auth/index.ts';
import { infoRouter } from './info/index.ts';
import { organizationRouter } from './organization/index.ts';

export const appRouter = router({
    auth: authRouter,
    info: infoRouter,
    account: accountRouter,
    organization: organizationRouter,
});

export type AppRouter = typeof appRouter;
