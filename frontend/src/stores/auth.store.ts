import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers/fetch-wrapper';
import { router } from '@/router';
import type { User } from '@apiTypes/user';
import { useNotificationStore } from './notification.store';
import { trpcClient } from '@/helpers/trpc';

export const useAuthStore = defineStore('auth', {
    state: (): { user: User | null, returnUrl: string | null, access_token: string | null, refresh_token: string | null } => ({
        user: JSON.parse(localStorage.getItem('user') as string),
        returnUrl: null,
        access_token: localStorage.getItem('access_token'),
        refresh_token: localStorage.getItem('refresh_token'),
    }),
    getters: {
        isAuthenticated(): boolean {
            return this.user !== null && this.access_token !== null && this.refresh_token !== null;
        }
    },
    actions: {
        async refreshUser() {
            if (!this.user) {
                return;
            }

            this.user = await fetchWrapper.get(`/account/me`);

            localStorage.setItem('user', JSON.stringify(this.user));
        },

        async login(email: string, password: string) {
            const result = await trpcClient.auth.loginWithPassword.mutate({
                email,
                password
            })

            this.setAccessToken(result.accessToken, result.refreshToken);

            const user = await fetchWrapper.get(`/account/me`);

            // update pinia state
            this.user = user;

            // store user details and jwt in local storage to keep user logged in between page refreshes
            localStorage.setItem('user', JSON.stringify(user));

            useNotificationStore().connect(result.accessToken);
            await useNotificationStore().loadNotifications();
        },

        async register(user: { email: string, password: string, username: string }) {
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

        async updateProfile(info: { username: string, email: string, currentPassword: string, newPassword: string }) {
            await fetchWrapper.patch(`/account/me`, info);

            if (info.newPassword && info.email) {
                await this.login(info.email, info.newPassword);
            }

            await this.refreshUser();
        },

        setAccessToken(access_token: string, refresh_token: string | null = null) {
            this.access_token = access_token;
            localStorage.setItem('access_token', access_token);

            if (refresh_token) {
                this.refresh_token = refresh_token;
                localStorage.setItem('refresh_token', refresh_token);
            }
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
            await fetchWrapper.delete(`/account/me`);
            this.logout();
        }
    },
});
