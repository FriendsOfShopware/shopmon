import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers';
import { router } from '@/router';
import { useAlertStore } from '@/stores';
import type { User } from '@apiTypes/user';

export const useAuthStore = defineStore('auth', {
  state: (): {isAuthenticated: boolean, user: User|null, returnUrl: string|null} => ({
    isAuthenticated: sessionStorage.getItem('user') !== null,
    user: JSON.parse(sessionStorage.getItem('user') as string),
    returnUrl: null,
  }),
  actions: {
    async refreshUser() {
      const user = await fetchWrapper.get(`/account/me`, '', {
        token: this.user!!.token,
      });

      user.token = this.user!!.token;

      this.user = user;

      sessionStorage.setItem('user', JSON.stringify(user));
    },

    async login(email: string, password: string) {
      try {
        const login = await fetchWrapper.post(`/auth/login`, {
          email,
          password,
        });

        const user = await fetchWrapper.get(`/account/me`, '', {
          token: login.token,
        });

        user.token = login.token

        // update pinia state
        this.user = user;

        // store user details and jwt in local storage to keep user logged in between page refreshes
        sessionStorage.setItem('user', JSON.stringify(user));

        // redirect to previous url or default to home page
        router.push(this.returnUrl || '/');
      } catch (error: any) {
        const alertStore = useAlertStore();
        alertStore.error(error);
      }
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

    async updateProfile(info: any) {
      await fetchWrapper.patch(`/account/me`, info);

      await this.refreshUser();
    },
    logout() {
      this.user = null;
      localStorage.removeItem('user');
      router.push('/account/login');
    },
    async delete() {
      await fetchWrapper.delete(`/account/me`);
      this.logout();
    }
  },
});
