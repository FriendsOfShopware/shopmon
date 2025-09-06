import {
    type RouteLocationNormalized,
    createRouter,
    createWebHistory,
} from 'vue-router';

import { useReturnUrl } from '@/composables/useReturnUrl';
import { authClient } from '@/helpers/auth-client';
import AppLayout from '@/layouts/AppLayout.vue';
import LoginLayout from '@/layouts/LoginLayout.vue';
import DefaultLayout from '@/layouts/DefaultLayout.vue';
import ShopDetailLayout from '@/layouts/ShopDetailLayout.vue';
import adminRoutes from '@/router/admin';
import { nextTick } from 'vue';

import FaShop from '~icons/fa6-solid/shop';
import FaCircleCheck from '~icons/fa6-solid/circle-check';
import FaFileWaveform from '~icons/fa6-solid/file-waveform';
import FaListCheck from '~icons/fa6-solid/list-check';
import FaPlug from '~icons/fa6-solid/plug';
import FaRocket from '~icons/fa6-solid/rocket';
import Dashboard from '~icons/ri/dashboard-fill';
import FaFolder from '~icons/fa6-solid/folder';
import FaBuilding from '~icons/fa6-solid/building';

const session = authClient.useSession();

export const router = createRouter({
    history: createWebHistory(import.meta.env.BASE_URL),
    linkActiveClass: 'active',
    routes: [
        {
            path: '/',
            component: DefaultLayout,
            children: [
                {
                    path: '',
                    name: 'home',
                    component: () => import('@/views/Home.vue'),
                },
                {
                    path: 'privacy',
                    name: 'privacy',
                    component: () => import('@/views/Privacy.vue'),
                },
                {
                    path: 'imprint',
                    name: 'imprint',
                    component: () => import('@/views/Imprint.vue'),
                },
            ],
        },
        {
            path: '/account',
            component: LoginLayout,
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
            ],
        },
        {
            path: '/app',
            component: AppLayout,
            children: [
                {
                    path: 'dashboard',
                    name: 'account.dashboard',
                    component: () => import('@/views/Dashboard.vue'),
                    meta: {
                        title: 'Dashboard',
                        icon: Dashboard
                    },
                },
                {
                    name: 'account.settings',
                    path: 'settings',
                    component: () => import('@/views/account/Settings.vue'),
                    meta: {
                        title: 'Settings',
                        icon: FaShop
                    },
                },
                {
                    name: 'account.project.list',
                    path: 'projects',
                    component: () => import('@/views/shop/ListProjects.vue'),
                    meta: {
                        title: 'My Projects',
                        icon: FaFolder
                    },
                },
                {
                    name: 'account.shops.new',
                    path: 'shops/new',
                    component: () => import('@/views/shop/AddShop.vue'),
                },
                {
                    name: 'account.projects.new',
                    path: 'projects/new',
                    component: () => import('@/views/shop/AddProject.vue'),
                },
                {
                    name: 'account.shops.edit',
                    path: 'organizations/:slug/shops/:shopId(\\d+)/edit',
                    component: () => import('@/views/shop/EditShop.vue'),
                },
                {
                    path: 'organizations/:slug/shops/:shopId(\\d+)',
                    component: ShopDetailLayout,
                    children: [
                        {
                            name: 'account.shops.detail',
                            path: '',
                            component: () =>
                                import('@/views/shop/detail/DetailShop.vue'),
                            meta: {
                                title: 'Shop informations',
                                icon: FaShop
                            },
                        },
                        {
                            name: 'account.shops.detail.checks',
                            path: 'checks',
                            component: () =>
                                import('@/views/shop/detail/DetailChecks.vue'),
                            meta: {
                                title: 'Checks',
                                icon: FaCircleCheck
                            },
                        },
                        {
                            name: 'account.shops.detail.extensions',
                            path: 'extensions',
                            component: () =>
                                import('@/views/shop/detail/DetailExtensions.vue'),
                            meta: {
                                title: 'Extensions',
                                icon: FaPlug
                            },
                        },
                        {
                            name: 'account.shops.detail.tasks',
                            path: 'tasks',
                            component: () =>
                                import('@/views/shop/detail/DetailScheduledTasks.vue'),
                            meta: {
                                title: 'Scheduled tasks',
                                icon: FaListCheck
                            },
                        },
                        {
                            name: 'account.shops.detail.queue',
                            path: 'queue',
                            component: () =>
                                import('@/views/shop/detail/DetailQueue.vue'),
                            meta: {
                                title: 'Queue',
                                icon: FaCircleCheck
                            },
                        },
                        {
                            name: 'account.shops.detail.sitespeed',
                            path: 'sitespeed',
                            component: () =>
                                import('@/views/shop/detail/DetailSitespeed.vue'),
                            meta: {
                                title: 'Sitespeed',
                                icon: FaRocket
                            },
                        },
                        {
                            name: 'account.shops.detail.changelog',
                            path: 'changelog',
                            component: () =>
                                import('@/views/shop/detail/DetailChangelog.vue'),
                            meta: {
                                title: 'Changelog',
                                icon: FaFileWaveform
                            },
                        },
                    ],
                },
                {
                    name: 'account.organizations.list',
                    path: 'organizations',
                    component: () =>
                        import('@/views/organization/ListOrganizations.vue'),
                    meta: {
                        title: 'My Organisations',
                        icon: FaBuilding
                    },
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
                    meta: {
                        title: 'My Extensions',
                        icon: FaPlug
                    },
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

    // redirect to the login page if not logged in and trying to access a restricted page
    const publicPages = [
        'home',
        'privacy',
        'imprint',
        'account.login',
        'account.register',
        'account.confirm',
        'account.forgot.password',
        'account.forgot.password.confirm',
    ];
    const authRequired = !publicPages.includes(to.name as string);
    const { setReturnUrl } = useReturnUrl();

    if (authRequired && !session.value.data) {
        setReturnUrl(to.fullPath);
        return { name: 'account.login' };
    } else if (to.name === 'account.login') {
        setReturnUrl('/app/dashboard');
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
