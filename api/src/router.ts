import { Hono } from 'hono';
import { sentry } from './middleware/sentry';
import { trpcServer } from './middleware/trpc';
import { appRouter } from './trpc/router';

export type Bindings = {
    MAIL_ACTIVE: 'true' | 'false';
    MAIL_FROM: string;

    RESEND_API_KEY: string;

    FRONTEND_URL: string;
    SENTRY_DSN: string;
    PAGESPEED_API_KEY: string;
    APP_SECRET: string;
    DISABLE_REGISTRATION: boolean;

    LIBSQL_URL?: string;
    LIBSQL_AUTH_TOKEN?: string;

    // cloudflare bindings
    kvStorage: KVNamespace;
    SHOPS_SCRAPE: DurableObjectNamespace;
    PAGESPEED_SCRAPE: DurableObjectNamespace;
    FILES: R2Bucket;
    shopmonDB: D1Database;
    sendMail: SendEmail;
};

const app = new Hono<{ Bindings: Bindings }>();

app.use('*', sentry());
app.use('/trpc/*', trpcServer({ router: appRouter }));

// serve screenshots
app.get('/pagespeed/:uuid/screenshot.jpg', async (c) => {
    const uuid = c.req.param('uuid');

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404,
        });
    }

    const file = await c.env.FILES.get(`pagespeed/${uuid}/screenshot.jpg`);

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404,
        });
    }

    return new Response(file?.body, {
        status: 200,
        headers: {
            'content-type': 'image/jpeg',
        },
    });
});

export default app;
