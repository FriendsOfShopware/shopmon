import { router, publicProcedure } from '../index.ts';
import { getLastInsertId, schema } from '../../db.ts';
import { eq, and } from 'drizzle-orm';
import { z } from 'zod';
import bcryptjs from 'bcryptjs';
import { randomString } from '../../util.ts';
import Users from '../../repository/users.ts';
import { TRPCError } from '@trpc/server';
import { loggedInUserMiddleware } from '../middleware.ts';
import Organizations from '../../repository/organization.ts';
import { sendMailConfirmToUser, sendMailResetPassword } from '../../mail/mail.ts';
import { passkeyRouter } from './passkey.ts';
const passwordReset = new Map<string, string>();

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

            const passwordIsValid = await bcryptjs.compare(
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

            console.log('Registering user with email:', input.email);
            console.log('User already exists:', result);

            if (result !== undefined) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Given email address is already registered',
                });
            }

            const salt = bcryptjs.genSaltSync(10);
            const hashedPassword = await bcryptjs.hash(input.password, salt);

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

            passwordReset.set(token, result.id.toString());

            setTimeout(() => {
                passwordReset.delete(token);
            }, 60 * 60 * 1000); // 1 hour

            await sendMailResetPassword(input, token);

            return true;
        }),
    passwordResetAvailable: publicProcedure
        .input(z.string())
        .mutation(async ({ input, ctx }) => {
            return passwordReset.has(input);
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

            const id = passwordReset.get(token);
            passwordReset.delete(token);

            if (!id) {
                throw new TRPCError({
                    code: 'BAD_REQUEST',
                    message: 'Invalid token',
                });
            }

            const salt = await bcryptjs.genSalt(10);
            const newPassword = await bcryptjs.hash(password, salt);

            await ctx.drizzle
                .update(schema.user)
                .set({ password: newPassword })
                .where(eq(schema.user.id, parseInt(id)))
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
