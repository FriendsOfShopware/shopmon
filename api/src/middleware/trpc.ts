import { captureException } from '@sentry/node';
import type { AnyRouter } from '@trpc/server';
import type { FetchHandlerRequestOptions } from '@trpc/server/adapters/fetch';
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';
import type { MiddlewareHandler } from 'hono';
import { createContext } from '../trpc/context.ts';

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
                    console.error(
                        `[tRPC] Error on path: ${err.path}`,
                        err.error,
                    );
                    captureException(err.error, {
                        user: {
                            id: err.ctx.user || null,
                        },
                        extra: {
                            type: err.type,
                            path: err.path,
                        },
                    });
                }
            },
        });
        return res;
    };
};
