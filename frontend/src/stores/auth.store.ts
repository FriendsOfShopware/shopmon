import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers';
import { router } from '@/router';
import { useAlertStore } from '@/stores';

export const useAuthStore = defineStore({
  id: 'auth',
  state: () => ({
    user: JSON.parse(localStorage.getItem('user')),
    returnUrl: null,
  }),
  actions: {
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
        localStorage.setItem('user', JSON.stringify(user));

        // redirect to previous url or default to home page
        router.push(this.returnUrl || '/');
      } catch (error) {
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
    logout() {
      this.user = null;
      localStorage.removeItem('user');
      router.push('/account/login');
    },
  },
});
