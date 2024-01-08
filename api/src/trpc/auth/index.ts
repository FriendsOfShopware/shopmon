import { router, publicProcedure } from '..';
import { schema } from '../../db';
import { eq, and, isNull } from 'drizzle-orm';
import { z } from 'zod';
import bcryptjs from "bcryptjs";
import { randomString } from '../../util';
import Users from '../../repository/users';
import { context } from '../context';
import { TRPCError } from '@trpc/server';
import { loggedInUserMiddleware } from '../middleware';
import Teams from '../../repository/teams';
import { sendMailConfirmToUser, sendMailResetPassword } from '../../mail/mail';

const REFRESH_TOKEN_TTL = 60 * 60 * 6; // 6 hours
const ACCESS_TOKEN_TTL = 60 * 30; // 30 minutes

export interface Token {
    id: number;
}

export const authRouter = router({
    loginWithPassword: publicProcedure
        .input(z.object({
            email: z.string().email(),
            password: z.string(),
        }))
        .mutation(async ({ input, ctx }) => {
            const loginError = new TRPCError({
                code: 'UNAUTHORIZED',
                message: 'Invalid email or password',
            })

            const result = await ctx.drizzle.query.user.findFirst({
                columns: {
                    id: true,
                    password: true
                },
                where: and(eq(schema.user.email, input.email.toLowerCase()), isNull(schema.user.verify_code))
            })

            if (result === undefined) {
                throw loginError;
            }

            const passwordIsValid = await bcryptjs.compare(input.password, result.password);

            if (!passwordIsValid) {
                throw loginError;
            }

            const refreshToken = `r-${result.id}-${randomString(32)}`;
            await ctx.env.kvStorage.put(
                refreshToken,
                JSON.stringify({
                    id: result.id
                }),
                {
                    expirationTtl: REFRESH_TOKEN_TTL,
                }
            );

            const accessToken = await getAuthentifikationToken(ctx, refreshToken);

            return {
                accessToken,
                refreshToken
            };
        }),
    refreshToken: publicProcedure.input(z.string()).mutation(async ({ input, ctx }) => {
        const accessToken = await getAuthentifikationToken(ctx, input);

        return {
            accessToken
        };
    }),
    register: publicProcedure.input(
        z.object({
            email: z.string().email(),
            password: z.string().min(5),
            username: z.string().min(8),
        })
    ).mutation(async ({ input, ctx }) => {
        input.email = input.email.toLowerCase();

        const result = await ctx.drizzle.query.user.findFirst({
            columns: {
                id: true
            },
            where: eq(schema.user.email, input.email)
        })

        if (result !== undefined) {
            throw new TRPCError({
                code: 'BAD_REQUEST',
                message: 'Given email address is already registered',
            });
        }

        const salt = bcryptjs.genSaltSync(10)
        const hashedPassword = await bcryptjs.hash(input.password, salt)

        const token = randomString(32)

        const userInsertResult = await ctx.drizzle.insert(schema.user)
            .values({
                created_at: (new Date()).toISOString(),
                email: input.email,
                username: input.username,
                password: hashedPassword,
                verify_code: token,
            })

        if (!userInsertResult.success) {
            throw new TRPCError({
                code: 'INTERNAL_SERVER_ERROR',
                message: 'Failed to create user',
            });
        }

        await Teams.create(ctx.drizzle, `${input.email}'s Team`, userInsertResult.meta.last_row_id);

        await sendMailConfirmToUser(ctx.env, input.email, token);

        return true;
    }),
    confirmVerification: publicProcedure.input(z.string()).mutation(async ({ input, ctx }) => {
        const result = await ctx.drizzle.query.user.findFirst({
            columns: {
                id: true
            },
            where: eq(schema.user.verify_code, input)
        })

        if (result === undefined) {
            throw new TRPCError({
                code: 'BAD_REQUEST',
                message: 'Invalid confirm token',
            });
        }

        await ctx.drizzle
            .update(schema.user)
            .set({ verify_code: null })
            .where(eq(schema.user.id, result.id))
            .execute();

        return true;
    }),
    passwordResetRequest: publicProcedure.input(z.string().email()).mutation(async ({ input, ctx }) => {
        const token = randomString(32);

        input = input.toLowerCase();

        const result = await ctx.drizzle.query.user.findFirst({
            columns: {
                id: true
            },
            where: eq(schema.user.email, input)
        })

        if (result === undefined) {
            return true;
        }

        await ctx.env.kvStorage.put(`reset_${token}`, result.id.toString(), {
            expirationTtl: 60 * 60, // 1 hour
        });

        await sendMailResetPassword(ctx.env, input, token);

        return true;
    }),
    passwordResetAvailable: publicProcedure.input(z.string()).mutation(async ({ input, ctx }) => {
        const result = await ctx.env.kvStorage.get(`reset_${input}`);

        return result !== null;
    }),
    passwordResetConfirm: publicProcedure.input(z.object({
        token: z.string(),
        password: z.string().min(8),
    })).mutation(async ({ input, ctx }) => {
        const { token, password } = input;

        const id = await ctx.env.kvStorage.get(`reset_${token}`);

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
    logout: publicProcedure.use(loggedInUserMiddleware).mutation(async ({ ctx }) => {
        if (ctx.user) {
            await Users.revokeUserSessions(ctx.env.kvStorage, ctx.user);
        }
    }),
});


async function getAuthentifikationToken(ctx: context, refreshToken: string) {
    const token = await ctx.env.kvStorage.get(refreshToken);

    if (token === null) {
        throw new Error('Invalid refresh token');
    }

    const data = JSON.parse(token) as Token;

    const existence = await Users.existsById(ctx.drizzle, data.id);

    if (existence === false) {
        throw new Error('Invalid refresh token');
    }

    const accessToken = `u-${data.id}-${randomString(32)}`;
    await ctx.env.kvStorage.put(
        accessToken,
        JSON.stringify({
            id: data.id
        } as Token),
        {
            expirationTtl: ACCESS_TOKEN_TTL,
        }
    );

    return accessToken;
}
