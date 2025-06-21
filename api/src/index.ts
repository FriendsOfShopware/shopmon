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

// Serve sitespeed results
app.get('/sitespeed/result/:shopId/*', async (c) => {
    const shopId = c.req.param('shopId');
    const filePath = c.req.param('*') || '';

    // Validate shop ID
    if (!/^\d+$/.test(shopId)) {
        return c.json({ error: 'Invalid shop ID' }, 400);
    }

    // Prevent directory traversal
    if (filePath.includes('..') || filePath.includes('\\')) {
        return c.json({ error: 'Invalid file path' }, 400);
    }

    const sitespeedDataFolder =
        process.env.APP_SITESPEED_DATA_FOLDER || './sitespeed-results';
    const fullPath = path.join(sitespeedDataFolder, shopId, filePath);

    try {
        const stat = await fs.stat(fullPath);

        if (stat.isDirectory()) {
            // List directory contents
            const files = await fs.readdir(fullPath);
            return c.json({
                type: 'directory',
                path: `/${filePath}`,
                files: files.map((file) => ({ name: file })),
            });
        }
        // Serve file
        const file = await fs.readFile(fullPath);
        const ext = path.extname(fullPath).toLowerCase();

        let contentType = 'application/octet-stream';
        if (ext === '.json') contentType = 'application/json';
        else if (ext === '.html') contentType = 'text/html';
        else if (ext === '.css') contentType = 'text/css';
        else if (ext === '.js') contentType = 'application/javascript';
        else if (ext === '.png') contentType = 'image/png';
        else if (ext === '.jpg' || ext === '.jpeg') contentType = 'image/jpeg';
        else if (ext === '.gif') contentType = 'image/gif';
        else if (ext === '.svg') contentType = 'image/svg+xml';
        else if (ext === '.mp4') contentType = 'video/mp4';
        else if (ext === '.webm') contentType = 'video/webm';

        return new Response(file, {
            status: 200,
            headers: {
                'content-type': contentType,
                'cache-control': 'public, max-age=3600', // Cache for 1 hour
            },
        });
    } catch (_e) {
        return c.json({ error: 'File not found' }, 404);
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
