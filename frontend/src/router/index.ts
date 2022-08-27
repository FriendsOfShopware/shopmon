import { createRouter, createWebHistory, RouteLocationNormalized } from 'vue-router';

import { useAuthStore, useAlertStore } from '@/stores';
import Home from '@/views/Home.vue';
import accountRoutes from './account.routes';


export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    linkActiveClass: 'active',
    routes: [
        { path: '/', component: Home },
        { ...accountRoutes },
        // catch all redirect to home page
        { path: '/:pathMatch(.*)*', redirect: '/' }
    ]
});

router.beforeEach(async (to: RouteLocationNormalized) => {
    // clear alert on route change
    const alertStore = useAlertStore();
    alertStore.clear();

    // redirect to login page if not logged in and trying to access a restricted page
    const publicPages = ['account.login', 'account.register', 'account.confirm', 'account.forgot.password', 'account.forgot.password.confirm'];
    const authRequired = !publicPages.includes(to.name as string);
    const authStore = useAuthStore();

    if (authRequired && !authStore.user) {
        authStore.returnUrl = to.fullPath;
        return '/account/login';
    } else if (authStore.user && publicPages.includes(to.name as string)) {
        // redirect to home page if logged in and trying to access a public page
        return '/';
    }
});
