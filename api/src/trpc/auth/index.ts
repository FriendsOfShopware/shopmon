import { router, publicProcedure } from '../index.ts';
import { getLastInsertId, schema, user } from '../../db.ts';
import { eq, and, gt } from 'drizzle-orm';
import { z } from 'zod';
import { randomString } from '../../util.ts';
import Users from '../../repository/users.ts';
import { TRPCError } from '@trpc/server';
import { loggedInUserMiddleware } from '../middleware.ts';
import Organizations from '../../repository/organization.ts';
import { sendMailConfirmToUser, sendMailResetPassword } from '../../mail/mail.ts';
import { passkeyRouter } from './passkey.ts';

export const ACCESS_TOKEN_TTL = 60 * 60 * 6; // 6 hours

export interface Token {
    id: number;
}

export const authRouter = router({
    passkey: passkeyRouter,
    loginWithPassword: publicProcedure
        .input(
            z.object({
                email: z.string().email(),
                password: z.string(),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const loginError = new TRPCError({
                code: 'BAD_REQUEST',
                message: 'Invalid email or password',
            });

            const result = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                    password: true,
                    verifyCode: true,
                },
                where: and(eq(schema.user.email, input.email.toLowerCase())),
            });

            if (result === undefined) {
                throw loginError;
            }

            if (result.verifyCode !== null) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Please confirm your email address first',
                });
            }

            const passwordIsValid = await Bun.password.verify(
                input.password,
                result.password,
            );

            if (!passwordIsValid) {
                throw loginError;
            }

            const token = `u-${result.id}-${randomString(32)}`;
            await ctx.drizzle.insert(schema.sessions).values({
                id: token,
                userId: result.id,
                expires: new Date(Date.now() + ACCESS_TOKEN_TTL * 1000),
            }).execute();

            return token;
        }),
    register: publicProcedure
        .input(
            z.object({
                email: z.string().email(),
                password: z.string().min(5),
                displayName: z.string(),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            input.email = input.email.toLowerCase();

            const result = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                },
                where: eq(schema.user.email, input.email),
            });

            if (result !== undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Given email address is already registered',
                });
            }

            const hashedPassword = await Bun.password.hash(input.password, {
                algorithm: 'bcrypt',
            });

            const token = randomString(32);

            const userInsertResult = await ctx.drizzle
                .insert(schema.user)
                .values({
                    email: input.email,
                    displayName: input.displayName,
                    password: hashedPassword,
                    verifyCode: token,
                    createdAt: new Date(),
                });

            // @ts-expect-error drizzle-lib-error
            const lastId = getLastInsertId(userInsertResult);

            await Organizations.create(
                ctx.drizzle,
                `${input.displayName}'s Organization`,
                lastId,
            );

            await sendMailConfirmToUser(input.email, token);

            return true;
        }),
    confirmVerification: publicProcedure
        .input(z.string())
        .mutation(async ({ input, ctx }) => {
            const result = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                },
                where: eq(schema.user.verifyCode, input),
            });

            if (result === undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid confirm token',
                });
            }

            await ctx.drizzle
                .update(schema.user)
                .set({ verifyCode: null })
                .where(eq(schema.user.id, result.id))
                .execute();

            return true;
        }),
    passwordResetRequest: publicProcedure
        .input(z.string().email())
        .mutation(async ({ input, ctx }) => {
            const token = randomString(32);

            input = input.toLowerCase();

            const result = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                },
                where: eq(schema.user.email, input),
            });

            if (result === undefined) {
                return true;
            }

            const expires = new Date(Date.now() + 60 * 60 * 1000); // 1 hour from now
            
            await ctx.drizzle.insert(schema.passwordResetTokens).values({
                id: randomString(32),
                userId: result.id,
                token: token,
                expires: expires,
                createdAt: new Date(),
            });

            await sendMailResetPassword(input, token);

            return true;
        }),
    passwordResetAvailable: publicProcedure
        .input(z.string())
        .mutation(async ({ input, ctx }) => {
            const result = await ctx.drizzle.query.passwordResetTokens.findFirst({
                where: and(
                    eq(schema.passwordResetTokens.token, input),
                    gt(schema.passwordResetTokens.expires, new Date())
                ),
            });
            return result !== undefined;
        }),
    passwordResetConfirm: publicProcedure
        .input(
            z.object({
                token: z.string(),
                password: z.string().min(8),
            }),
        )
        .mutation(async ({ input, ctx }) => {
            const { token, password } = input;

            const tokenData = await ctx.drizzle.query.passwordResetTokens.findFirst({
                where: and(
                    eq(schema.passwordResetTokens.token, token),
                    gt(schema.passwordResetTokens.expires, new Date())
                ),
            });

            if (!tokenData) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid token',
                });
            }

            // Delete the token
            await ctx.drizzle
                .delete(schema.passwordResetTokens)
                .where(eq(schema.passwordResetTokens.id, tokenData.id))
                .execute();

            const newPassword = await Bun.password.hash(password, {
                algorithm: 'bcrypt',
            });

            await ctx.drizzle
                .update(schema.user)
                .set({ password: newPassword, verifyCode: null })
                .where(eq(schema.user.id, tokenData.userId))
                .execute();

            return true;
        }),
    logout: publicProcedure
        .use(loggedInUserMiddleware)
        .mutation(async ({ ctx }) => {
            if (ctx.user) {
                await Users.revokeUserSessions(ctx.user);
            }
        }),
});
