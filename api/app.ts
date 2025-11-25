import './src/sentry.ts';
import { captureException } from '@sentry/node';
import { auth } from './src/auth.ts';
import { createContext } from './src/trpc/context.ts';
import { appRouter } from './src/trpc/router.ts';
import { fetchRequestHandler } from '@trpc/server/adapters/fetch';

export default {
    async fetch(request: Request): Promise<Response> {
        const pathName = new URL(request.url).pathname;
        if (pathName.startsWith('/auth')) {
            return auth.handler(request);
        }

        if (pathName.startsWith('/trpc')) {
            return fetchRequestHandler({
                req: request,
                router: appRouter,
                endpoint: '/trpc',
                createContext: createContext(),
                onError: (err) => {
                    if (err.error.code === 'INTERNAL_SERVER_ERROR') {
                        console.error(
                            `[tRPC] Error on path: ${err.path}`,
                            err.error,
                        );
                        captureException(err.error, {
                            user: {
                                id: err.ctx?.user?.id,
                            },
                            extra: {
                                type: err.type,
                                path: err.path,
                            },
                        });
                    }
                },
            });
        }

        return new Response(null, {status: 404})
    },
};
