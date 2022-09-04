import Toucan from "toucan-js";

export function createSentry(ctx: ExecutionContext, env: Env, req: Request|null = null) {
    const options: any = {
        dsn: env.SENTRY_DSN,
        context: ctx,
        allowedHeaders: ['user-agent'],
        allowedSearchParams: /(.*)/,
    };

    if (req) {
        options.request = req;
    }

    return new Toucan(options);
}
