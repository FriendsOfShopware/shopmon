import { Hono } from "hono";
import { sentry } from './middleware/sentry';
import { trpcServer } from "./middleware/trpc";
import { appRouter } from './trpc/router'

export type Bindings = {
    MAIL_ACTIVE: 'true' | 'false';
    MAIL_FROM: string;
    MAIL_FROM_NAME: string;

    MAIL_DKIM_PRIVATE_KEY: string | undefined;
    MAIL_DKIM_DOMAIN: string;
    MAIL_DKIM_SELECTOR: string;

    FRONTEND_URL: string;
    SENTRY_DSN: string;
    PAGESPEED_API_KEY: string;
    APP_SECRET: string;
    DISABLE_REGISTRATION: boolean;

    // cloudflare bindings
    kvStorage: KVNamespace;
    SHOPS_SCRAPE: DurableObjectNamespace;
    PAGESPEED_SCRAPE: DurableObjectNamespace;
    USER_SOCKET: DurableObjectNamespace;
    FILES: R2Bucket;
    shopmonDB: D1Database;
    sendMail: SendEmail;
}

const app = new Hono<{ Bindings: Bindings }>()

app.use('*', sentry());
app.use(
    '/trpc/*',
    trpcServer({ router: appRouter }),
)

// serve screenshots
app.get('/pagespeed/:uuid/screenshot.jpg', async (c) => {
    const uuid = c.req.param('uuid');

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404
        });
    }

    const file = await c.env.FILES.get(`pagespeed/${uuid}/screenshot.jpg`)

    if (uuid.indexOf('/') !== -1) {
        return new Response('', {
            status: 404
        });
    }

    return new Response(file?.body, {
        status: 200,
        headers: {
            "content-type": "image/jpeg"
        }
    });
});

app.get('/api/ws', async (c) => {
    const authToken = new URL(c.req.url).searchParams.get('auth_token');

    if (!authToken) {
        return new Response('Invalid token', { status: 400 });
    }

    const token = await c.env.kvStorage.get(authToken);

    if (token === null) {
        return new Response('Invalid token', { status: 400 });
    }

    const data = JSON.parse(token) as { id: number };

    return c.env.USER_SOCKET.get(c.env.USER_SOCKET.idFromName(data.id.toString())).fetch(c.req.raw)
});

export default app;

