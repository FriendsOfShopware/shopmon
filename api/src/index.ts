import router from './router'
import { createSentry } from './sentry';

export default {
    fetch(request: Request, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env, request);

        return router
            .handle(request, env, ctx, sentry)
            .catch((err) => {
                sentry.captureException(err);
                return new Response(
                    JSON.stringify({ message: err.message }),
                    { 
                        status: 500,
                        headers: {
                            'Content-Type': 'application/json'
                        }
                    }
                );
            })
    }
}

export { ShopScrape } from './object/ShopScrape';
export { PagespeedScrape } from './object/PagespeedScrape';
export { UserSocket } from './object/UserSocket';