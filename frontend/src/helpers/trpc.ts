import { createTRPCProxyClient, httpBatchLink } from '@trpc/client'
import type { AppRouter } from '../../../api/src/trpc/router'
import type { inferRouterInputs, inferRouterOutputs } from '@trpc/server';

export type RouterOutput = inferRouterOutputs<AppRouter>;
export type RouterInput = inferRouterInputs<AppRouter>;

export const trpcClient = createTRPCProxyClient<AppRouter>({
    links: [
        httpBatchLink({
            url: '/trpc',
            async headers() {
                const headers: { Authorization?: string } = {};

                if (localStorage.getItem('access_token')) {
                    headers['Authorization'] = `${localStorage.getItem('access_token')}`;
                }

                return headers;
            }
        }),

    ],
})
