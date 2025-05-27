import { Hono } from 'hono';
import { trpcServer } from './middleware/trpc';
import { appRouter } from './trpc/router';

const app = new Hono();

app.use('/trpc/*', trpcServer({ router: appRouter }));

export default app;
