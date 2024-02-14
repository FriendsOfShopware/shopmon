import { Toucan } from 'toucan-js';
import { Context } from 'toucan-js/dist/types';
import type { Bindings } from './router';

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
    env: Bindings,
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
