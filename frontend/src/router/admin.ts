import AdminLayout from '@/layouts/AdminLayout.vue';
import type { RouteRecordRaw } from 'vue-router';

const adminRoutes: RouteRecordRaw[] = [
    {
        path: '/admin',
        component: AdminLayout,
        meta: { requiresAuth: true },
        children: [
            {
                path: '',
                redirect: '/admin/dashboard',
            },
            {
                path: 'dashboard',
                name: 'admin.dashboard',
                component: () => import('@/views/admin/Dashboard.vue'),
            },
            {
                path: 'users',
                name: 'admin.users',
                component: () => import('@/views/admin/ListUsers.vue'),
            },
            {
                path: 'organizations',
                name: 'admin.organizations',
                component: () => import('@/views/admin/ListOrganizations.vue'),
            },
            {
                path: 'shops',
                name: 'admin.shops',
                component: () => import('@/views/admin/ListShops.vue'),
            },
        ],
    },
];

export default adminRoutes;
