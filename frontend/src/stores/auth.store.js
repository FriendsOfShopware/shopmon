import { defineStore } from 'pinia';

import { fetchWrapper } from '@/helpers';
import { router } from '@/router';
import { useAlertStore } from '@/stores';

const baseUrl = `/api`;

export const useAuthStore = defineStore({
  id: 'auth',
  state: () => ({
    // initialize state from local storage to enable user to stay logged in
    user: JSON.parse(localStorage.getItem('user')),
    returnUrl: null,
  }),
  actions: {
    async login(email, password) {
      try {
        const login = await fetchWrapper.post(`${baseUrl}/auth/login`, {
          email,
          password,
        });

        const user = await fetchWrapper.get(`${baseUrl}/account/me`, '', {
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
    async confirmMail(token) {
      await fetchWrapper.post(`${baseUrl}/auth/confirm/${token}`);
    },
    logout() {
      this.user = null;
      localStorage.removeItem('user');
      router.push('/account/login');
    },
  },
});