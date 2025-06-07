import type { FetchCreateContextFnOptions } from '@trpc/server/adapters/fetch';
import type { Session, User } from 'better-auth/types';
import { auth } from '../auth.ts';
import { type Drizzle, getConnection } from '../db.ts';

export type context = {
    user: (User & { notifications: string[] }) | null;
    session: (Session & { activeOrganizationId?: string }) | null;
    drizzle: Drizzle;
    headers: Headers;
};

export function createContext() {
    return async ({ req }: FetchCreateContextFnOptions) => {
        let user = null;
        let session = null;

        const sessionResponse = await auth.api.getSession({
            headers: req.headers,
        });

        if (sessionResponse?.session && sessionResponse?.user) {
            session = sessionResponse.session;
            user = sessionResponse.user;
        }

        return {
            user,
            session,
            headers: req.headers,
            drizzle: getConnection(),
        };
    };
}
