import {onSchedule} from './cron/schedule';
import router from './router'
import Toucan from "toucan-js";

export default {
    fetch(request: Request, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env);

        return router
            .handle(request, env, ctx, sentry)
            .catch((err) => {
                sentry.captureException(err);
                return new Response(err.message, {status: 500});
            })
    },

    scheduled(event: ScheduledEvent, env: Env, ctx: ExecutionContext) {
        const sentry = createSentry(ctx, env);

        ctx.waitUntil(
            onSchedule(env)
            .catch(err => sentry.captureException(err))
        )
    }
}

function createSentry(ctx: ExecutionContext, env: Env) {
    return new Toucan({
        dsn: env.SENTRY_DSN,
        context: ctx,
        allowedHeaders: ['user-agent'],
        allowedSearchParams: /(.*)/,
    });
}
