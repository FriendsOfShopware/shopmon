import { createTRPCClient, httpBatchLink } from '@trpc/client';
import type { inferRouterInputs, inferRouterOutputs } from '@trpc/server';
import type { AppRouter } from '../../../api/src/trpc/router';

export type RouterOutput = inferRouterOutputs<AppRouter>;
export type RouterInput = inferRouterInputs<AppRouter>;

export const trpcClient = createTRPCClient<AppRouter>({
    links: [
        httpBatchLink({
            url: '/trpc',
            fetch: async (
                requestInfo: RequestInfo | URL | string,
                init?: RequestInit,
            ) => {
                const resp = await fetch(requestInfo, init);

                if (resp.status === 401) {
                    const clonedResp = resp.clone();
                    let json = (await clonedResp.json()) as {
                        error: { message: string };
                    }[];

                    // check is json is a array
                    if (!Array.isArray(json)) {
                        json = [json];
                    }

                    for (const error of json) {
                        if (error.error.message === 'Token expired') {
                            localStorage.removeItem('user');
                            window.location.reload();
                        }
                    }
                }

                return resp;
            },
        }),
    ],
});
