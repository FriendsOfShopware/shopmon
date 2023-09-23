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
        { name: 'account.shops.edit', path: 'teams/edit/:teamId(\\d+)/:shopId(\\d+)', component: () => import('@/views/account/EditShop.vue') },
        { name: 'account.shops.detail', path: 'teams/:teamId(\\d+)/:shopId(\\d+)', component: () => import('@/views/account/DetailShop.vue') },
        { name: 'account.teams.list', path: 'teams', component: () => import('@/views/account/ListTeams.vue') },
        { name: 'account.teams.new', path: 'teams/new', component: () => import('@/views/account/AddTeam.vue') },
        { name: 'account.teams.detail', path: 'teams/:teamId(\\d+)', component: () => import('@/views/account/DetailTeam.vue') },
        { name: 'account.teams.edit', path: 'teams/edit/:teamId(\\d+)', component: () => import('@/views/account/EditTeam.vue') },
        { name: 'account.extension.list', path: 'extensions', component: () => import('@/views/account/ListExtensions.vue') },
    ]
};
