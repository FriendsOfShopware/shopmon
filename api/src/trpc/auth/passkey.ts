import { server } from '@passwordless-id/webauthn';
import type {
    AuthenticationResponseJSON,
    RegistrationJSON,
} from '@passwordless-id/webauthn/dist/esm/types';
import { TRPCError } from '@trpc/server';
import { and, eq } from 'drizzle-orm';
import { z } from 'zod';
import { schema } from '../../db.ts';
import { randomString } from '../../util.ts';
import { publicProcedure, router } from '../index.ts';
import { loggedInUserMiddleware } from '../middleware.ts';
import { ACCESS_TOKEN_TTL } from './index.ts';
const webauthnChallanges = new Map<string, string>();

export const passkeyRouter = router({
    challenge: publicProcedure.mutation(async ({ ctx }) => {
        const challange = btoa(
            crypto.getRandomValues(new Uint32Array(4)).toString(),
        ).replaceAll('=', '');

        webauthnChallanges.set(challange, challange);
        setTimeout(() => {
            webauthnChallanges.delete(challange);
        }, 120 * 1000);

        return challange;
    }),
    registerDevice: publicProcedure
        .use(loggedInUserMiddleware)
        .input(
            z.object({
                // Credential ID fields
                id: z.string(), // Base64URLString
                rawId: z.string(), // Base64URLString

                // Response data
                response: z.object({
                    attestationObject: z.string(), // Base64URLString
                    authenticatorData: z.string(), // Base64URLString
                    clientDataJSON: z.string(), // Base64URLString
                    transports: z.array(
                        z.enum([
                            'ble',
                            'hybrid',
                            'internal',
                            'nfc',
                            'smart-card',
                            'usb',
                        ]),
                    ),
                    publicKey: z.string(), // Base64URLString
                    publicKeyAlgorithm: z.number(), // COSEAlgorithmIdentifier
                }),

                // Optional authenticator attachment
                authenticatorAttachment: z
                    .enum(['cross-platform', 'platform'])
                    .optional(),

                // Client extension results
                clientExtensionResults: z.record(z.unknown()),

                // Credential type
                type: z.literal('public-key'),

                // User information
                user: z.object({
                    id: z.string().optional(),
                    name: z.string(),
                    displayName: z.string().optional(),
                }),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const expected = {
                challenge: async (challange: string) => {
                    if (!webauthnChallanges.has(challange)) {
                        return false;
                    }

                    webauthnChallanges.delete(challange);

                    return true;
                },
                origin: process.env.FRONTEND_URL,
            };

            const registrationParsed = await server.verifyRegistration(
                input as RegistrationJSON,
                expected,
            );

            await ctx.drizzle
                .insert(schema.userPasskeys)
                // @ts-expect-error
                .values({
                    id: input.id,
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
                id: z.string(),
                rawId: z.string(),
                type: z.string(),
                response: z.object({
                    clientDataJSON: z.string(),
                    authenticatorData: z.string(),
                    signature: z.string(),
                    userHandle: z.string().optional(),
                }),
                authenticatorAttachment: z.string().optional(),
                clientExtensionResults: z.record(z.any()),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const credential = await ctx.drizzle.query.userPasskeys.findFirst({
                columns: {
                    key: true,
                    userId: true,
                },
                where: eq(schema.userPasskeys.id, input.id),
            });

            if (credential === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid credential',
                });
            }

            const expected = {
                challenge: async (challange: string) => {
                    if (!webauthnChallanges.has(challange)) {
                        return false;
                    }
                    webauthnChallanges.delete(challange);

                    return true;
                },
                origin: process.env.FRONTEND_URL,
                userVerified: true,
            };

            await server.verifyAuthentication(
                input as AuthenticationResponseJSON,
                // @ts-expect-error
                credential.key.credential,
                expected,
            );

            const token = `u-${credential.userId}-${randomString(32)}`;

            await ctx.drizzle.insert(schema.sessions).values({
                id: token,
                userId: credential.userId,
                expires: new Date(Date.now() + ACCESS_TOKEN_TTL * 1000),
            });

            return token;
        }),
});
