import {
    type RouteLocationNormalized,
    createRouter,
    createWebHistory,
} from 'vue-router';

import { useReturnUrl } from '@/composables/useReturnUrl';
import { authClient } from '@/helpers/auth-client';
import AuthenticatedLayout from '@/layouts/AuthenticatedLayout.vue';
import UnauthenticatedLayout from '@/layouts/UnauthenticatedLayout.vue';
import adminRoutes from '@/router/admin';
import Home from '@/views/Home.vue';
import { nextTick } from 'vue';

const session = authClient.useSession();

export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    linkActiveClass: 'active',
    routes: [
        {
            path: '/',
            redirect: { name: 'home' },
        },
        {
            path: '/account',
            component: UnauthenticatedLayout,
            children: [
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
                {
                    name: 'privacy.unauthenticated',
                    path: 'privacy',
                    component: () => import('@/views/Privacy.vue'),
                },
                {
                    name: 'imprint.unauthenticated',
                    path: 'imprint',
                    component: () => import('@/views/Imprint.vue'),
                },
            ],
        },
        {
            path: '/app',
            component: AuthenticatedLayout,
            children: [
                { path: '/', name: 'home', component: Home },
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
                    path: 'organizations/:slug/shops/:shopId(\\d+)/edit',
                    component: () => import('@/views/shop/EditShop.vue'),
                },
                {
                    name: 'account.shops.detail',
                    path: 'organizations/:slug/shops/:shopId(\\d+)',
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
                    path: 'organizations/:slug',
                    component: () =>
                        import('@/views/organization/DetailOrganization.vue'),
                },
                {
                    name: 'account.organizations.edit',
                    path: 'organizations/edit/:slug',
                    component: () =>
                        import('@/views/organization/EditOrganization.vue'),
                },
                {
                    name: 'account.organizations.sso',
                    path: 'organizations/:slug/sso',
                    component: () =>
                        import('@/views/organization/ManageSSO.vue'),
                },
                {
                    name: 'account.extension.list',
                    path: 'extensions',
                    component: () =>
                        import('@/views/account/ListExtensions.vue'),
                },
                {
                    name: 'account.organization.accept',
                    path: 'organizations/accept/:token',
                    component: () =>
                        import(
                            '@/views/organization/AcceptRejectInvitation.vue'
                        ),
                    props: {
                        action: 'accept',
                    },
                },
                {
                    name: 'account.organization.reject',
                    path: 'organizations/reject/:token',
                    component: () =>
                        import(
                            '@/views/organization/AcceptRejectInvitation.vue'
                        ),
                    props: {
                        action: 'reject',
                    },
                },
                {
                    name: 'privacy.authenticated',
                    path: 'privacy',
                    component: () => import('@/views/Privacy.vue'),
                },
                {
                    name: 'imprint.authenticated',
                    path: 'imprint',
                    component: () => import('@/views/Imprint.vue'),
                },
            ],
        },
        ...adminRoutes,
        // catch all redirect to home page
        {
            path: '/:pathMatch(.*)*',
            name: 'not-found',
            redirect: { name: 'home' },
        },
    ],
});

router.beforeEach(async (to: RouteLocationNormalized) => {
    if (session.value.isPending) {
        await new Promise((resolve) => {
            const a = setInterval(() => {
                if (!session.value.isPending) {
                    clearInterval(a);
                    resolve(true);
                }
            }, 50);
        });
    }

    // redirect to login page if not logged in and trying to access a restricted page
    const publicPages = [
        'account.login',
        'account.register',
        'account.confirm',
        'account.forgot.password',
        'account.forgot.password.confirm',
        'privacy.unauthenticated',
        'imprint.unauthenticated',
    ];
    const authRequired = !publicPages.includes(to.name as string);
    const { setReturnUrl } = useReturnUrl();

    if (authRequired && !session.value.data) {
        setReturnUrl(to.fullPath);
        return { name: 'account.login' };
    }
    if (session.value.data && publicPages.includes(to.name as string)) {
        // redirect to home page if logged in and trying to access a public page
        return { name: 'home' };
    }

    // Check admin routes
    if (to.path.startsWith('/admin') && session.value.data) {
        const userRole = session.value.data.user.role;
        if (userRole !== 'admin') {
            return { name: 'home' };
        }
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
