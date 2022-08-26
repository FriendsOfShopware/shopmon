import { Layout, Login, Register, AccountConfirm, ForgotPassword } from '@/views/account';

export default {
    path: '/account',
    component: Layout,
    children: [
        { name: 'account.login', path: 'login', component: Login },
        { name: 'account.register', path: 'register', component: Register },
        { name: 'account.confirm', path: 'confirm/:token', component: AccountConfirm },
        { path: 'forgot-password', component: ForgotPassword },
    ]
};