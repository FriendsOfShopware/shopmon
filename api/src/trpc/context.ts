import { context as otelContext, trace } from '@opentelemetry/api';
import type { FetchCreateContextFnOptions } from '@trpc/server/adapters/fetch';
import type { Session, User } from 'better-auth/types';
import { auth } from '../auth.ts';
import { type Drizzle, getConnection } from '../db.ts';

const tracer = trace.getTracer('trpc');

export type context = {
    user: (User & { notifications: string[] }) | null;
    session: (Session & { activeOrganizationId?: string }) | null;
    drizzle: Drizzle;
    headers: Headers;
};

export function createContext() {
    return async ({ req }: FetchCreateContextFnOptions) => {
        const span = tracer.startSpan(
            `trpc.createContext`,
            {},
            otelContext.active(),
        );

        return otelContext.with(
            trace.setSpan(otelContext.active(), span),
            async () => {
                let user = null;
                let session = null;

                const sessionResponse = await auth.api.getSession({
                    headers: req.headers,
                });

                if (sessionResponse?.session && sessionResponse?.user) {
                    session = sessionResponse.session;
                    user = sessionResponse.user;

                    span.setAttribute('user.id', user.id);
                    span.setAttribute('session.id', session.id);
                }

                span.end();

                return {
                    user,
                    session,
                    headers: req.headers,
                    drizzle: getConnection(),
                };
            },
        );
    };
}
