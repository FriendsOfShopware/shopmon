import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers/fetch-wrapper';
import { router } from '@/router';
import type { User } from '@apiTypes/user'; 

export const useAuthStore = defineStore('auth', {
    state: (): {isAuthenticated: boolean, user: User|null, returnUrl: string|null, access_token: string|null, refresh_token: string|null} => ({
        isAuthenticated: localStorage.getItem('user') !== null,
        user: JSON.parse(localStorage.getItem('user') as string),
        returnUrl: null,
        access_token: localStorage.getItem('access_token'),
        refresh_token: localStorage.getItem('refresh_token'),
    }),
    actions: {
        async refreshUser() {
            if(!this.user) {
                return;
            }

            this.user = await fetchWrapper.get(`/account/me`);

            localStorage.setItem('user', JSON.stringify(this.user));
        },

        async login(email: string, password: string) {
            const login = await fetchWrapper.post(`/auth/token`, {
                client_id: 'shopmon',
                grant_type: 'password',
                username: email,
                password: password,
            });

            this.setAccessToken(login.access_token, login.refresh_token);

            const user = await fetchWrapper.get(`/account/me`);

            // update pinia state
            this.user = user;

            // store user details and jwt in local storage to keep user logged in between page refreshes
            localStorage.setItem('user', JSON.stringify(user));
        },
        async register(user: object) {
            await fetchWrapper.post(`/auth/register`, user);
        },
        async confirmMail(token: string) {
            await fetchWrapper.post(`/auth/confirm/${token}`);
        },
        async resetPassword(email: string) {
            await fetchWrapper.post(`/auth/reset`, { email });
        },
        async resetAvailable(token: string) {
            await fetchWrapper.get(`/auth/reset/${token}`);
        },
        async confirmResetPassword(token: string, password: string) {
            await fetchWrapper.post(`/auth/reset/${token}`, { password });
        },

        async updateProfile(info: {username: string, email: string, currentPassword: string, newPassword: string}) {
            await fetchWrapper.patch(`/account/me`, info);

            if (info.newPassword && info.email) {
                await this.login(info.email, info.newPassword);
            }

            await this.refreshUser();
        },

        setAccessToken(access_token: string, refresh_token: string|null = null) {
            this.access_token = access_token;
            localStorage.setItem('access_token', access_token);
            
            if (refresh_token) {
                this.refresh_token = refresh_token;
                localStorage.setItem('refresh_token', refresh_token);
            }
        },

        logout() {
            this.user = null;
            localStorage.removeItem('user');
            localStorage.removeItem('access_token');
            localStorage.removeItem('refresh_token');
            router.push('/account/login');
        },
        async delete() {
            await fetchWrapper.delete(`/account/me`);
            this.logout();
        }
    },
});
