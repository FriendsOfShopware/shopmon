import { router } from '.';
import { accountRouter } from './account';
import { authRouter } from './auth';
import { infoRouter } from './info';

export const appRouter = router({
    auth: authRouter,
    info: infoRouter,
    account: accountRouter,
})

export type AppRouter = typeof appRouter
