import { router, publicProcedure } from '..';
import { server } from '@passwordless-id/webauthn';
import { loggedInUserMiddleware } from '../middleware';
import { z } from 'zod';
import { schema } from '../../db';
import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { randomString } from '../../util';
import { ACCESS_TOKEN_TTL, Token } from '.';

export const passkeyRouter = router({
    challenge: publicProcedure.mutation(async ({ ctx }) => {
        const challange = btoa(
            crypto.getRandomValues(new Uint32Array(4)).toString(),
        ).replaceAll('=', '');

        await ctx.env.kvStorage.put(
            `webauthn_challange_${challange}`,
            challange,
            {
                expirationTtl: 120,
            },
        );

        return challange;
    }),
    registerDevice: publicProcedure
        .use(loggedInUserMiddleware)
        .input(
            z.object({
                username: z.string().email(),
                credential: z.object({
                    id: z.string(),
                    publicKey: z.string(),
                    algorithm: z.union([
                        z.literal('ES256'),
                        z.literal('RS256'),
                    ]),
                }),
                authenticatorData: z.string(),
                clientData: z.string(),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const expected = {
                challenge: async (challange: string) => {
                    const ret = await ctx.env.kvStorage.get(
                        `webauthn_challange_${challange}`,
                    );

                    if (ret === null) {
                        return false;
                    }

                    await ctx.env.kvStorage.delete(
                        `webauthn_challange_${challange}`,
                    );

                    return true;
                },
                origin: ctx.env.FRONTEND_URL,
            };

            const registrationParsed = await server.verifyRegistration(
                input,
                expected,
            );

            await ctx.drizzle
                .insert(schema.userPasskeys)
                .values({
                    id: input.credential.id,
                    name: registrationParsed.authenticator.name,
                    userId: ctx.user,
                    key: registrationParsed,
                    createdAt: new Date(),
                })
                .execute();
        }),

    listDevices: publicProcedure
        .use(loggedInUserMiddleware)
        .query(async ({ ctx }) => {
            const result = await ctx.drizzle.query.userPasskeys.findMany({
                columns: {
                    id: true,
                    name: true,
                    createdAt: true,
                },
                where: eq(schema.userPasskeys.userId, ctx.user),
            });

            return result;
        }),
    removeDevice: publicProcedure
        .use(loggedInUserMiddleware)
        .input(z.string())
        .mutation(async ({ input, ctx }) => {
            await ctx.drizzle
                .delete(schema.userPasskeys)
                .where(
                    and(
                        eq(schema.userPasskeys.id, input),
                        eq(schema.userPasskeys.userId, ctx.user),
                    ),
                )
                .execute();

            return true;
        }),
    authenticateDevice: publicProcedure
        .input(
            z.object({
                authenticatorData: z.string(),
                clientData: z.string(),
                signature: z.string(),
                credentialId: z.string(),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const credential = await ctx.drizzle.query.userPasskeys.findFirst({
                columns: {
                    key: true,
                    userId: true,
                },
                where: eq(schema.userPasskeys.id, input.credentialId),
            });

            console.log(credential?.key.credential.id);

            if (credential === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid credential',
                });
            }

            const expected = {
                challenge: async (challange: string) => {
                    console.log(`Looking for: ${challange}`);
                    const ret = await ctx.env.kvStorage.get(
                        `webauthn_challange_${challange}`,
                    );

                    if (ret === null) {
                        return false;
                    }

                    console.log(ret, challange);

                    await ctx.env.kvStorage.delete(
                        `webauthn_challange_${challange}`,
                    );

                    return true;
                },
                origin: ctx.env.FRONTEND_URL,
                userVerified: true,
            };

            await server.verifyAuthentication(
                input,
                credential.key.credential,
                expected,
            );

            const token = `u-${credential.userId}-${randomString(32)}`;
            await ctx.env.kvStorage.put(
                token,
                JSON.stringify({
                    id: credential.userId,
                } as Token),
                {
                    expirationTtl: ACCESS_TOKEN_TTL,
                },
            );

            return token;
        }),
});
