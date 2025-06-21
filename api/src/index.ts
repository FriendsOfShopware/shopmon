import './sentry.ts';
import { existsSync, mkdirSync, readFileSync } from 'node:fs';
import { serve } from '@hono/node-server';
import { serveStatic } from '@hono/node-server/serve-static';
import { Hono } from 'hono';
import { auth } from './auth.ts';
import { trpcServer } from './middleware/trpc.ts';
import { appRouter } from './trpc/router.ts';
import './cron/index.ts';

const filesDir = process.env.APP_FILES_DIR || './files';

if (!existsSync(filesDir)) {
    mkdirSync(filesDir, { recursive: true });
}

const app = new Hono();

// Better Auth routes
app.on(['POST', 'GET'], '/auth/*', (c) => {
    return auth.handler(c.req.raw);
});

// tRPC routes
app.use('/trpc/*', trpcServer({ router: appRouter }));

app.get(
    '/sitespeed',
    serveStatic({
        root: process.env.APP_SITESPEED_DATA_FOLDER || './sitespeed-results',
    }),
);

// Health check endpoint
app.get('/health', (c) => {
    return c.json({ status: 'ok' });
});

if (existsSync('./dist')) {
    app.use('*', serveStatic({ root: './dist' }));
    const indexHtml = readFileSync('./dist/index.html', 'utf-8');
    app.notFound((c) => {
        return c.html(indexHtml, 200, {
            'Content-Type': 'text/html',
        });
    });
}

serve({
    fetch: app.fetch,
    port: Number.parseInt(process.env.PORT || '3000', 10),
});
export default app;
