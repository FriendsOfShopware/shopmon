/* eslint-disable max-len */
import Layout from '@/views/auth/Layout.vue';

export default {
    path: '/account',
    component: Layout,
    children: [
        // without login
        { name: 'account.login', path: 'login', component: () => import('@/views/auth/Login.vue') },
        { name: 'account.register', path: 'register', component: () => import('@/views/auth/Register.vue') },
        { name: 'account.confirm', path: 'confirm/:token', component: () => import('@/views/auth/AccountConfirm.vue') },
        { name: 'account.forgot.password', path: 'forgot-password', component: () => import('@/views/auth/ForgotPassword.vue') },
        { name: 'account.forgot.password.confirm', path: 'forgot-password/:token', component: () => import('@/views/auth/ForgotPasswordConfirm.vue') },

        // with login
        { name: 'account.settings', path: 'settings', component: () => import('@/views/account/Settings.vue') },
        { name: 'account.shops.list', path: 'shops', component: () => import('@/views/account/ListShops.vue') },
        { name: 'account.shops.new', path: 'shops/new', component: () => import('@/views/account/AddShop.vue') },
        { name: 'account.shops.edit', path: 'organizations/edit/:organizationId(\\d+)/:shopId(\\d+)', component: () => import('@/views/account/EditShop.vue') },
        { name: 'account.shops.detail', path: 'organizations/:organizationId(\\d+)/:shopId(\\d+)', component: () => import('@/views/account/DetailShop.vue') },
        { name: 'account.organizations.list', path: 'organizations', component: () => import('@/views/account/ListOrganizations.vue') },
        { name: 'account.organizations.new', path: 'organizations/new', component: () => import('@/views/account/AddOrganization.vue') },
        { name: 'account.organizations.detail', path: 'organizations/:organizationId(\\d+)', component: () => import('@/views/account/DetailOrganization.vue') },
        { name: 'account.organizations.edit', path: 'organizations/edit/:organizationId(\\d+)', component: () => import('@/views/account/EditOrganization.vue') },
        { name: 'account.extension.list', path: 'extensions', component: () => import('@/views/account/ListExtensions.vue') },
    ],
};
