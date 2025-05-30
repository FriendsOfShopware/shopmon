import { betterAuth } from 'better-auth';
import { drizzleAdapter } from 'better-auth/adapters/drizzle';
import { passkey } from 'better-auth/plugins/passkey';
import { getConnection } from './db.js';

export const auth = betterAuth({
    baseURL: process.env.FRONTEND_URL || 'http://localhost:3000',
    basePath: '/auth',
    secret: process.env.APP_SECRET,
    database: drizzleAdapter(getConnection(), {
        provider: 'sqlite',
    }),
    emailAndPassword: {
        enabled: true,
        requireEmailVerification: true,
        password: {
            hash: async (password) => {
                return Bun.password.hash(password, {
                    algorithm: 'bcrypt',
                });
            },
            verify: async (data) => {
                return Bun.password.verify(data.password, data.hash);
            },
        },
    },
    emailVerification: {
        sendVerificationEmail: async ({ user, token }) => {
            // Import dynamically to avoid circular dependency
            const { sendMailConfirmToUser } = await import('./mail/mail.js');
            await sendMailConfirmToUser(user.email, token);
        },
    },
    account: {
        accountLinking: {
            enabled: true,
        },
    },
    forgetPassword: {
        enabled: true,
        sendResetPasswordEmail: async ({ user, token }) => {
            // Import dynamically to avoid circular dependency
            const { sendMailResetPassword } = await import('./mail/mail.js');
            await sendMailResetPassword(user.email, token);
        },
    },
    rateLimit: {
        enabled: true,
        window: 60, // 1 minute
        max: 10, // 10 requests per minute
    },
    plugins: [
        passkey({
            rpID: process.env.FRONTEND_URL
                ? new URL(process.env.FRONTEND_URL).hostname
                : 'localhost',
            rpName: 'Shopmon',
        }),
    ],
});

// Export auth client type for frontend
export type AuthClient = typeof auth;
