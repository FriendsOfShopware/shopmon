import Toucan from "toucan-js";
import { Context } from "toucan-js/dist/types";

export function createSentry(ctx: ExecutionContext|Context, env: Env, req: Request|null = null) {
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
