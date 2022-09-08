import router from '../router'
import { createSentry } from '../sentry';

export default {
    fetch(request: Request, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env, request);

        return router
            .handle(request, env, ctx, sentry)
            .catch((err) => {
                sentry.captureException(err);
                return new Response(err.message, { status: 500 });
            })
    }
}

export { ShopScrape } from '../object/ShopScrape';
export { PagespeedScrape } from '../object/PagespeedScrape';
export { UserSocket } from '../object/UserSocket';