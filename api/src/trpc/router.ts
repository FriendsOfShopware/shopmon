import { router } from '.';
import { authRouter } from './auth';
import { infoRouter } from './info';

export const appRouter = router({
    auth: authRouter,
    info: infoRouter,
})

export type AppRouter = typeof appRouter
