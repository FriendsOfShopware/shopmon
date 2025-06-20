import './sentry.ts';
import { existsSync, promises as fs, mkdirSync, readFileSync } from 'node:fs';
import path from 'node:path';
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

// Serve screenshots from local files
app.get('/pagespeed/:uuid/screenshot.jpg', async (c) => {
    const uuid = c.req.param('uuid');

    if (uuid.includes('/') || uuid.includes('..')) {
        return c.json({ error: 'Invalid UUID' }, 404);
    }

    const filePath = path.join(filesDir, 'pagespeed', uuid, 'screenshot.jpg');

    try {
        const file = await fs.readFile(filePath);
        return new Response(file, {
            status: 200,
            headers: {
                'content-type': 'image/jpeg',
                'cache-control': 'public, max-age=86400', // Cache for 1 day
            },
        });
    } catch (_e) {
        return c.json({ error: 'Screenshot not found' }, 404);
    }
});

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
