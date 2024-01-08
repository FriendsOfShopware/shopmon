import { Toucan } from 'toucan-js';
import { Context } from 'toucan-js/dist/types';

interface SentryOptions {
    dsn: string;
    release: string;
    context: ExecutionContext | Context;
    allowedHeaders?: string[];
    allowedSearchParams?: RegExp;
    request?: Request;
}

export function createSentry(
    ctx: ExecutionContext | Context,
    env: Env,
    req: Request | null = null,
) {
    const options: SentryOptions = {
        dsn: env.SENTRY_DSN,
        release: SENTRY_RELEASE,
        context: ctx,
        allowedHeaders: ['user-agent'],
        allowedSearchParams: /(.*)/,
    };

    if (req) {
        options.request = req;
    }

    return new Toucan(options);
}
