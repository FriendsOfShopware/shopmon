import {
    type RouteLocationNormalized,
    createRouter,
    createWebHistory,
} from 'vue-router';

import Layout from '@/layouts/Layout.vue';
import { useAlertStore } from '@/stores/alert.store';
import { useAuthStore } from '@/stores/auth.store';
import Home from '@/views/Home.vue';
import { nextTick } from 'vue';

export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    linkActiveClass: 'active',
    routes: [
        { path: '/', name: 'home', component: Home },
        {
            path: '/account',
            component: Layout,
            children: [
                // without login
                {
                    name: 'account.login',
                    path: 'login',
                    component: () => import('@/views/auth/Login.vue'),
                },
                {
                    name: 'account.register',
                    path: 'register',
                    component: () => import('@/views/auth/Register.vue'),
                },
                {
                    name: 'account.confirm',
                    path: 'confirm/:token',
                    component: () => import('@/views/auth/AccountConfirm.vue'),
                },
                {
                    name: 'account.forgot.password',
                    path: 'forgot-password',
                    component: () => import('@/views/auth/ForgotPassword.vue'),
                },
                {
                    name: 'account.forgot.password.confirm',
                    path: 'forgot-password/:token',
                    component: () =>
                        import('@/views/auth/ForgotPasswordConfirm.vue'),
                },

                // with login
                {
                    name: 'account.settings',
                    path: 'settings',
                    component: () => import('@/views/account/Settings.vue'),
                },
                {
                    name: 'account.shops.list',
                    path: 'shops',
                    component: () => import('@/views/shop/ListShops.vue'),
                },
                {
                    name: 'account.shops.new',
                    path: 'shops/new',
                    component: () => import('@/views/shop/AddShop.vue'),
                },
                {
                    name: 'account.shops.edit',
                    path: 'organizations/edit/:organizationId(\\d+)/:shopId(\\d+)',
                    component: () => import('@/views/shop/EditShop.vue'),
                },
                {
                    name: 'account.shops.detail',
                    path: 'organizations/:organizationId(\\d+)/:shopId(\\d+)',
                    component: () =>
                        import('@/views/shop/detail/DetailShop.vue'),
                },
                {
                    name: 'account.organizations.list',
                    path: 'organizations',
                    component: () =>
                        import('@/views/organization/ListOrganizations.vue'),
                },
                {
                    name: 'account.organizations.new',
                    path: 'organizations/new',
                    component: () =>
                        import('@/views/organization/AddOrganization.vue'),
                },
                {
                    name: 'account.organizations.detail',
                    path: 'organizations/:organizationId(\\d+)',
                    component: () =>
                        import('@/views/organization/DetailOrganization.vue'),
                },
                {
                    name: 'account.organizations.edit',
                    path: 'organizations/edit/:organizationId(\\d+)',
                    component: () =>
                        import('@/views/organization/EditOrganization.vue'),
                },
                {
                    name: 'account.extension.list',
                    path: 'extensions',
                    component: () =>
                        import('@/views/account/ListExtensions.vue'),
                },
            ],
        },
        // catch all redirect to home page
        { path: '/:pathMatch(.*)*', redirect: { name: 'home' } },
    ],
});

router.beforeEach(async (to: RouteLocationNormalized) => {
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

    if (
        import.meta.env.VITE_DISABLE_REGISTRATION &&
        (to.name as string) === 'account.register'
    ) {
        return { name: 'home' };
    }

    if (authRequired && !authStore.user) {
        authStore.returnUrl = to.fullPath;
        return { name: 'account.login' };
    }
    if (authStore.user && publicPages.includes(to.name as string)) {
        // redirect to home page if logged in and trying to access a public page
        return { name: 'home' };
    }
});

const DEFAULT_TITLE = 'Shopware Monitoring';
router.afterEach(async (to) => {
    await nextTick();

    const title = to.meta.title;
    if (typeof title === 'string') {
        document.title = title;
        return;
    }

    document.title = DEFAULT_TITLE;
});
