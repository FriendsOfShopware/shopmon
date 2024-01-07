import { Router } from "itty-router";
import teamRouter from './api/team';
import accountRouter from "./api/account";
import { JsonResponse } from "./api/common/response";
import infoRouter from "./api/info";
import { Hono } from "hono";
import { sentry } from './middleware/sentry';
import { trpcServer } from "./middleware/trpc";
import { appRouter } from './trpc/router'

const router = Router();

export type Bindings = {
    MAIL_ACTIVE: 'true' | 'false';
    MAIL_FROM: string;
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

router.all("/api/account/*", accountRouter.handle);
router.all('/api/team/*', teamRouter.handle);
router.all('/api/info/*', infoRouter.handle);

router.get('/api/ws', async (req: Request, env: Env) => {
    const authToken = new URL(req.url).searchParams.get('auth_token');

    if (!authToken) {
        return new Response('Invalid token', { status: 400 });
    }

    const token = await env.kvStorage.get(authToken);

    if (token === null) {
        return new Response('Invalid token', { status: 400 });
    }

    const data = JSON.parse(token) as { id: number };

    return env.USER_SOCKET.get(env.USER_SOCKET.idFromName(data.id.toString())).fetch(req)
});

router.all('*', () => new JsonResponse({ message: 'Not found' }, 404));

app.all('*', async (c) => {
    return await router.handle(c.req.raw, c.env, c.executionCtx, c.get('sentry'))
})

export default app;

