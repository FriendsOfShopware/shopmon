import { router } from '.';
import { accountRouter } from './account';
import { authRouter } from './auth';
import { infoRouter } from './info';
import { organizationRouter } from './organization';

export const appRouter = router({
    auth: authRouter,
    info: infoRouter,
    account: accountRouter,
    organization: organizationRouter
})

export type AppRouter = typeof appRouter
