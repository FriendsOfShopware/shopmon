import * as Sentry from '@sentry/node';
import { initTRPC } from '@trpc/server';
import type { context } from './context.ts';

export const t = initTRPC.context<context>().create();

const sentryMiddleware = t.middleware(
    Sentry.trpcMiddleware({
        attachRpcInput: true,
    }),
);

export const publicProcedure = t.procedure.use(sentryMiddleware);
export const router = t.router;
