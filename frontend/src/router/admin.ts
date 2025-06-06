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
                redirect: '/admin/users',
            },
            {
                path: 'users',
                name: 'admin.users',
                component: () => import('@/views/admin/ListUsers.vue'),
            },
        ],
    },
];

export default adminRoutes;
