import { Router } from "itty-router";
import teamRouter from './api/team';
import authRouter from "./api/auth";
import accountRouter from "./api/account";
import { JsonResponse } from "./api/common/response";
import infoRouter from "./api/info";
import { Token } from "./api/auth/oauth";
import { Hono } from "hono";
import { sentry } from '@hono/sentry';

const router = Router();

const app = new Hono();

app.use('*', sentry({
    release: SENTRY_RELEASE,
}));

router.all("/api/account/*", accountRouter.handle);
router.all('/api/team/*', teamRouter.handle);
router.all('/api/auth/*', authRouter.handle);
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

    const data = JSON.parse(token) as Token;

    return env.USER_SOCKET.get(env.USER_SOCKET.idFromName(data.id.toString())).fetch(req)
});

router.all('*', () => new JsonResponse({ message: 'Not found' }, 404));


app.mount('*', router.handle);

export default app;

