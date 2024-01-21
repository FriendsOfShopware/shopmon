import { defineStore } from 'pinia';
import { router } from '@/router';
import { useNotificationStore } from './notification.store';
import { trpcClient, RouterOutput } from '@/helpers/trpc';
import { client } from '@passwordless-id/webauthn';

export const useAuthStore = defineStore('auth', {
    state: (): {
        user: RouterOutput['account']['currentUser'] | null,
        returnUrl: string | null,
        access_token: string | null,
        refresh_token: string | null,
        passkeys: RouterOutput['auth']['passkey']['listDevices'] | null,
    } => ({
        user: JSON.parse(localStorage.getItem('user') as string),
        returnUrl: null,
        access_token: localStorage.getItem('access_token'),
        refresh_token: localStorage.getItem('refresh_token'),
        passkeys: null,
    }),
    getters: {
        isAuthenticated(): boolean {
            return this.user !== null && this.access_token !== null && this.refresh_token !== null;
        },
    },
    actions: {
        async refreshUser() {
            if (!this.user) {
                return;
            }

            this.user = await trpcClient.account.currentUser.query();

            localStorage.setItem('user', JSON.stringify(this.user));
        },

        async loadPasskeys() {
            this.passkeys = await trpcClient.auth.passkey.listDevices.query();

            return this.passkeys;
        },

        async login(email: string, password: string) {
            const token = await trpcClient.auth.loginWithPassword.mutate({
                email,
                password,
            });

            localStorage.setItem('access_token', token);

            const user = await trpcClient.account.currentUser.query();

            // update pinia state
            this.user = user;

            // store user details and jwt in local storage to keep user logged in between page refreshes
            localStorage.setItem('user', JSON.stringify(user));

            useNotificationStore().connect(token);
            await useNotificationStore().loadNotifications();
        },

        async loginWithPasskey() {
            const challenge = await trpcClient.auth.passkey.challenge.mutate();
            const authentication = await client.authenticate([], challenge, {
                authenticatorType: 'auto',
                userVerification: 'required',
                timeout: 60000,
            });

            const token = await trpcClient.auth.passkey.authenticateDevice.mutate(authentication);

            localStorage.setItem('access_token', token);
            const user = await trpcClient.account.currentUser.query();

            // update pinia state
            this.user = user;

            // store user details and jwt in local storage to keep user logged in between page refreshes
            localStorage.setItem('user', JSON.stringify(user));

            useNotificationStore().connect(token);
            await useNotificationStore().loadNotifications();
        },

        async register(user: { email: string, password: string, displayName: string }) {
            trpcClient.auth.register.mutate(user);
        },

        async confirmMail(token: string) {
            await trpcClient.auth.confirmVerification.mutate(token);
        },

        async resetPassword(email: string) {
            await trpcClient.auth.passwordResetRequest.mutate(email);
        },

        async resetAvailable(token: string) {
            return await trpcClient.auth.passwordResetAvailable.mutate(token);
        },

        async confirmResetPassword(token: string, password: string) {
            return await trpcClient.auth.passwordResetConfirm.mutate({ token, password });
        },

        async updateProfile(info: { displayName: string, email: string, currentPassword: string, newPassword: string }) {
            await trpcClient.account.updateCurrentUser.mutate(info);

            if (info.newPassword && info.email) {
                await this.login(info.email, info.newPassword);
            }

            await this.refreshUser();
        },

        async logout() {
            // If session expired, it's fine that trpc fails
            try {
                await trpcClient.auth.logout.mutate();
            } catch (e) { }

            this.user = null;
            this.refresh_token = null;
            this.access_token = null;
            localStorage.removeItem('user');
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            router.push('/account/login');
        },

        async delete() {
            useNotificationStore().disconnect();
            await trpcClient.account.deleteCurrentUser.mutate();
            this.logout();
        },
    },
});
