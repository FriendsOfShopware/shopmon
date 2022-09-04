import Toucan from "toucan-js";

export function createSentry(ctx: ExecutionContext, env: Env) {
    return new Toucan({
        dsn: env.SENTRY_DSN,
        context: ctx,
        allowedHeaders: ['user-agent'],
        allowedSearchParams: /(.*)/,
    });
}
