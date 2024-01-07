import { createTRPCProxyClient, httpBatchLink } from '@trpc/client'
import type { AppRouter } from '../../../api/src/trpc/router'

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
