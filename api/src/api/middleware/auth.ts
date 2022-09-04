import Toucan from "toucan-js";

export async function validateToken(req: Request, env: Env, ctx: ExecutionContext, sentry: Toucan): Promise<Response|void> {
    if (req.headers.get('token') === null) {
        return new Response('Invalid token', {
            status: 401
        });
    }

    const token = req.headers.get('token');

    const result = await env.kvStorage.get(token as string);

    if (result === null) {
        return new Response('Invalid token', {
            status: 401
        });
    }

    const data = JSON.parse(result);

    req.userId = data.id;

    sentry.setUser({id: req.userId })
}