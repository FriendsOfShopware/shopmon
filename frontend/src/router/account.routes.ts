import Layout from '@/views/account/Layout.vue';
import Profile from '@/views/account/Profile.vue';
import Login from '@/views/account/Login.vue';
import Register from '@/views/account/Register.vue';
import AccountConfirm from '@/views/account/AccountConfirm.vue';
import ForgotPassword from '@/views/account/ForgotPassword.vue';
import ForgotPasswordConfirm from '@/views/account/ForgotPasswordConfirm.vue';

export default {
    path: '/account',
    component: Layout,
    children: [
        { name: 'account.settings', path: 'settings', component: Profile },
        { name: 'account.login', path: 'login', component: Login },
        { name: 'account.register', path: 'register', component: Register },
        { name: 'account.confirm', path: 'confirm/:token', component: AccountConfirm },
        { name: 'account.forgot.password', path: 'forgot-password', component: ForgotPassword },
        { name: 'account.forgot.password.confirm', path: 'forgot-password/:token', component: ForgotPasswordConfirm },
    ]
};
