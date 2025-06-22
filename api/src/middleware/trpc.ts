import { SpanStatusCode, trace } from '@opentelemetry/api';
import type { AnyRouter } from '@trpc/server';
import type { FetchHandlerRequestOptions } from '@trpc/server/adapters/fetch';
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
import type { MiddlewareHandler } from 'hono';
import { createContext } from '../trpc/context.ts';
import { logger } from '../util.ts';

type tRPCOptions = Omit<
    FetchHandlerRequestOptions<AnyRouter>,
    'req' | 'endpoint'
> &
    Partial<Pick<FetchHandlerRequestOptions<AnyRouter>, 'endpoint'>>;

export const trpcServer = ({
    endpoint = '/trpc',
    ...rest
}: tRPCOptions): MiddlewareHandler => {
    return async (c) => {
        const res = fetchRequestHandler({
            ...rest,
            endpoint,
            req: c.req.raw,
            createContext: createContext(),
            onError: (err) => {
                if (err.error.code === 'INTERNAL_SERVER_ERROR') {
                    logger.error(
                        `[tRPC] Error on path: ${err.path}, message: ${err.error.message}`,
                    );

                    const activeSpan = trace.getActiveSpan();

                    if (activeSpan) {
                        // Set error status on the span
                        activeSpan.setStatus({
                            code: SpanStatusCode.ERROR,
                            message: err.error.message,
                        });

                        // Record the error as an exception
                        activeSpan.recordException(err.error);

                        // Add error attributes
                        activeSpan.setAttributes({
                            'error.type': err.error.code,
                            'error.message': err.error.message,
                        });
                    }
                }
            },
        });
        return res;
    };
};
