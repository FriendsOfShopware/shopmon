import router from '../src/router'
import { createSentry } from '../src/sentry';

export default {
    fetch(request: Request, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env, request);

        return router
            .handle(request, env, ctx, sentry)
            .catch((err) => {
                sentry.captureException(err);
                return new Response(err.message, {status: 500});
            })
    }
}