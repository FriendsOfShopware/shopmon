import { FetchCreateContextFnOptions } from '@trpc/server/adapters/fetch';
import { Bindings } from '../router';
import { getConnection, Drizzle } from '../db';
import { TRPCError } from '@trpc/server';

export type context = {
    user: number | null;
    env: Bindings;
    drizzle: Drizzle;
    executionCtx: ExecutionContext;
};

export function createContext(
    bindings: Bindings,
    executionCtx: ExecutionContext,
) {
    return async ({ req }: FetchCreateContextFnOptions) => {
        const auth = req.headers.get('authorization');

        let user = null;

        if (auth) {
            const token = await bindings.kvStorage.get(auth);

            if (token === null) {
                throw new TRPCError({
                    code: 'UNAUTHORIZED',
                    message: 'Token expired',
                });
            }

            const data = JSON.parse(token) as { id: number };
            user = data.id;
        }

        return {
            user,
            env: bindings,
            drizzle: getConnection(bindings),
            executionCtx,
        };
    };
}
