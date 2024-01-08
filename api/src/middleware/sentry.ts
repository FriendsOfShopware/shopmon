import type { MiddlewareHandler } from 'hono';
import { Toucan } from 'toucan-js';
import { createSentry } from '../toucan';

declare module 'hono' {
    interface ContextVariableMap {
        sentry: Toucan;
    }
}

export const sentry = (): MiddlewareHandler => {
    return async (c, next) => {
        const sentry = createSentry(c.executionCtx, c.env, c.req.raw);

        c.set('sentry', sentry);

        await next();
        if (c.error) {
            sentry.captureException(c.error);
        }
    };
};
