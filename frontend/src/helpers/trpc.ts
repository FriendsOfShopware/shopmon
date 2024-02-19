import { createTRPCProxyClient, httpBatchLink, getFetch } from '@trpc/client'
import type { FetchEsque } from '@trpc/client/src/internals/types'
import type { AppRouter } from '../../../api/src/trpc/router'
import type { inferRouterInputs, inferRouterOutputs } from '@trpc/server';

export type RouterOutput = inferRouterOutputs<AppRouter>;
export type RouterInput = inferRouterInputs<AppRouter>;

export const trpcClient = createTRPCProxyClient<AppRouter>({
    links: [
        httpBatchLink({
            url: '/trpc',
            fetch: async (requestInfo: RequestInfo | URL | string, init?: RequestInit) => {
                const resp = await fetch(requestInfo, init);

                if (resp.status === 401 && localStorage.getItem('access_token')) {
                    const clonedResp = resp.clone()
                    let json = await clonedResp.json() as { error: { message: string }}[];

                    // check is json is a array
                    if (!Array.isArray(json)) {
                      json = [json];
                    }

                    for (const error of json) {
                        if (error.error.message === 'Token expired') {
                            localStorage.removeItem('access_token');
                            localStorage.removeItem('user');
                            window.location.reload();
                        }
                    }
                }

                return resp;
            },
            async headers() {
                const headers: { Authorization?: string } = {};

                if (localStorage.getItem('access_token')) {
                    headers.Authorization = `${localStorage.getItem('access_token')}`;
                }

                return headers;
            }
        }),
    ],

})
