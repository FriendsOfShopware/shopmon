import Layout from '@/views/auth/Layout.vue';
import Login from '@/views/auth/Login.vue';
import Register from '@/views/auth/Register.vue';
import AccountConfirm from '@/views/auth/AccountConfirm.vue';
import ForgotPassword from '@/views/auth/ForgotPassword.vue';
import ForgotPasswordConfirm from '@/views/auth/ForgotPasswordConfirm.vue';

import Settings from '@/views/account/Settings.vue';
import ListShops from '@/views/account/ListShops.vue';
import AddShop from '@/views/account/AddShop.vue';

export default {
    path: '/account',
    component: Layout,
    children: [
        // without login
        { name: 'account.login', path: 'login', component: Login },
        { name: 'account.register', path: 'register', component: Register },
        { name: 'account.confirm', path: 'confirm/:token', component: AccountConfirm },
        { name: 'account.forgot.password', path: 'forgot-password', component: ForgotPassword },
        { name: 'account.forgot.password.confirm', path: 'forgot-password/:token', component: ForgotPasswordConfirm },

        // with login
        { name: 'account.settings', path: 'settings', component: Settings },
        { name: 'account.shops.list', path: 'shops', component: ListShops},
        { name: 'account.shops.new', path: 'shops/new', component: AddShop},
    ]
};
