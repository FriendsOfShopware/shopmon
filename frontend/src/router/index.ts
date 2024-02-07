import { createRouter, createWebHistory } from 'vue-router';

import { useAuthStore } from '@/stores/auth.store';
import Home from '@/views/Home.vue';
import accountRoutes from './account.routes';
import { useAlertStore } from '@/stores/alert.store';
import { nextTick } from 'vue';

export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    linkActiveClass: 'active',
    routes: [
        { path: '/', component: Home },
        { ...accountRoutes },
        // catch all redirect to home page
        { path: '/:pathMatch(.*)*', redirect: '/' },
    ],
});

router.beforeEach(async (to) => {
    // clear alert on route change
    const alertStore = useAlertStore();
    alertStore.clear();

    // redirect to login page if not logged in and trying to access a restricted page
    const publicPages = [
        'account.login',
        'account.register',
        'account.confirm',
        'account.forgot.password',
        'account.forgot.password.confirm',
    ];
    const authRequired = !publicPages.includes(to.name as string);
    const authStore = useAuthStore();

    if(import.meta.env.VITE_DISABLE_REGISTRATION && to.name as string === 'account.register') {
        return '/';
    }

    if (authRequired && !authStore.user) {
        authStore.returnUrl = to.fullPath;
        return '/account/login';
    } else if (authStore.user && publicPages.includes(to.name as string)) {
        // redirect to home page if logged in and trying to access a public page
        return '/';
    }
});

const DEFAULT_TITLE = 'Shopware Monitoring';
router.afterEach(async (to) => {
    await nextTick();

    const title = to.meta.title;
    if(typeof title === 'string') {
        document.title = title;
        return;
    }

    document.title = DEFAULT_TITLE;
});
