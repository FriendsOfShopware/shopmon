import './sentry.ts';
import { serve } from '@hono/node-server';
import { Hono } from 'hono';
import { auth } from './auth.ts';
import { trpcServer } from './middleware/trpc.ts';
import { appRouter } from './trpc/router.ts';
import './cron/index.ts';

const app = new Hono();

// Better Auth routes
app.on(['POST', 'GET'], '/auth/*', (c) => {
    return auth.handler(c.req.raw);
});

// tRPC routes
app.use('/trpc/*', trpcServer({ router: appRouter }));

// Health check endpoint
app.get('/health', (c) => {
    return c.json({ status: 'ok' });
});

serve({
    fetch: app.fetch,
    port: Number.parseInt(process.env.PORT || '3000', 10),
});

export default app;
