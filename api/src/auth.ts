import { compare, hash } from 'bcrypt';
import { betterAuth } from 'better-auth';
import { drizzleAdapter } from 'better-auth/adapters/drizzle';
import { admin, organization } from 'better-auth/plugins';
import { passkey } from 'better-auth/plugins/passkey';
import { sso } from 'better-auth/plugins/sso';
import { getConnection } from './db.ts';
import shops from './repository/shops.ts';

export const auth = betterAuth({
    baseURL: process.env.FRONTEND_URL || 'http://localhost:3000',
    basePath: '/auth',
    secret: process.env.APP_SECRET,
    database: drizzleAdapter(getConnection(), {
        provider: 'sqlite',
    }),
    user: {
        additionalFields: {
            notifications: {
                type: 'string[]',
                defaultValue: [],
                returned: true,
                required: false,
            },
        },
    },
    emailAndPassword: {
        enabled: true,
        requireEmailVerification: true,
        async sendResetPassword(data, _request) {
            const { sendMailResetPassword } = await import('./mail/mail.ts');
            await sendMailResetPassword(data.user.email, data.token);
        },
        password: {
            hash: async (password) => {
                return hash(password, 10);
            },
            verify: async (data) => {
                return compare(data.password, data.hash);
            },
        },
    },
    socialProviders: {
        github: {
            clientId: process.env.APP_OAUTH_GITHUB_CLIENT_ID,
            clientSecret: process.env.APP_OAUTH_GITHUB_CLIENT_SECRET,
        },
    },
    emailVerification: {
        sendVerificationEmail: async ({ user, token }) => {
            const { sendMailConfirmToUser } = await import('./mail/mail.ts');
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
            const { sendMailResetPassword } = await import('./mail/mail.ts');
            await sendMailResetPassword(user.email, token);
        },
    },
    rateLimit: {
        enabled: true,
        window: 60, // 1 minute
        max: 10, // 10 requests per minute
    },
    plugins: [
        admin(),
        passkey({
            rpID: process.env.FRONTEND_URL
                ? new URL(process.env.FRONTEND_URL).hostname
                : 'localhost',
            rpName: 'Shopmon',
        }),
        organization({
            sendInvitationEmail: async (data) => {
                const { sendMailInviteToOrganization } = await import(
                    './mail/mail.js'
                );
                await sendMailInviteToOrganization(
                    data.email,
                    data.organization.name,
                    data.inviter.user.name,
                    data.invitation.id,
                );
            },
            cancelPendingInvitationsOnReInvite: true,
            organizationDeletion: {
                async beforeDelete(data, _request) {
                    return shops.deleteShopsByOrganization(
                        data.organization.id,
                    );
                },
            },
        }),
        sso({
            organizationProvisioning: {
                disabled: false,
                defaultRole: 'member',
            },
        }),
    ],
});

// Export auth client type for frontend
export type AuthClient = typeof auth;
