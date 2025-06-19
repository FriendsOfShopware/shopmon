import { Hono } from 'hono';
import { trpcServer } from './middleware/trpc.ts';
import { appRouter } from './trpc/router.ts';

const app = new Hono();

app.use('/trpc/*', trpcServer({ router: appRouter }));

export default app;
