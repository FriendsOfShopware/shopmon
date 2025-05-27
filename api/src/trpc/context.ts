import { TRPCError } from '@trpc/server';
import type { FetchCreateContextFnOptions } from '@trpc/server/adapters/fetch';
import { eq } from 'drizzle-orm';
import { type Drizzle, getConnection, schema } from '../db.ts';

export type context = {
    user: number | null;
    drizzle: Drizzle;
};

export function createContext() {
    return async ({ req }: FetchCreateContextFnOptions) => {
        const auth = req.headers.get('authorization');

        let user = null;

        if (auth) {
            const token = await getConnection()
                .select({
                    id: schema.sessions.userId,
                    expires: schema.sessions.expires,
                })
                .from(schema.sessions)
                .where(eq(schema.sessions.id, auth));

            if (token.length === 0) {
                throw new TRPCError({
                    code: 'UNAUTHORIZED',
                    message: 'Token expired',
                });
            }

            if (token[0].expires < new Date()) {
                await getConnection()
                    .delete(schema.sessions)
                    .where(eq(schema.sessions.id, auth));

                throw new TRPCError({
                    code: 'UNAUTHORIZED',
                    message: 'Token expired',
                });
            }

            user = token[0].id as number;
        }

        return {
            user,
            drizzle: getConnection(),
        };
    };
}
