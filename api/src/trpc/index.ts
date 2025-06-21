import { context as otelContext, trace } from '@opentelemetry/api';
import { initTRPC } from '@trpc/server';
import { flatten } from 'flat';
import type { context } from './context.ts';

export const t = initTRPC.context<context>().create();
const tracer = trace.getTracer('trpc');

const tracing = t.middleware((opts) => {
    const span = tracer.startSpan(
        `trpc.${opts.path}`,
        {
            attributes: {
                'trpc.method': opts.path,
            },
        },
        otelContext.active(),
    );

    span.setAttributes(flatten({ input: opts.input }));

    return otelContext.with(trace.setSpan(otelContext.active(), span), () => {
        return opts
            .next({
                ctx: {
                    ...opts.ctx,
                    span,
                },
            })
            .finally(() => {
                span.end();
            });
    });
});

export const publicProcedure = t.procedure.use(tracing);
export const router = t.router;
