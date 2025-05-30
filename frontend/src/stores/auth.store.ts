import { authClient } from '@/helpers/auth-client';
import { type RouterOutput, trpcClient } from '@/helpers/trpc';
import { router } from '@/router';
import { defineStore } from 'pinia';
import { computed, ref } from 'vue';
import { useNotificationStore } from './notification.store';

export const useAuthStore = defineStore('auth', () => {
    // const {
    //     session,
    //     isPending, //loading state
    //     error, //error object
    //     refetch //refetch the session
    // } = authClient.useSession();

    const user = ref<RouterOutput['account']['currentUser'] | null>(
        JSON.parse(localStorage.getItem('user') as string),
    );

    const userAvatar = ref(
        'https://api.dicebear.com/7.x/personas/svg/?seed=default?d=identicon',
    );

    const organizations = ref<
        RouterOutput['account']['listOrganizations'] | null
    >(null);

    const returnUrl = ref<string | null>(null);

    authClient.getSession().then(async (session) => {
        if (session.data?.user) {
            trpcClient.account.listOrganizations.query().then((orgs) => {
                organizations.value = orgs;
            });

            const userEmailSha256 = Array.from(
                new Uint8Array(
                    await crypto.subtle.digest(
                        'SHA-256',
                        new TextEncoder().encode(session.data.user.email),
                    ),
                ),
            )
                .map((b) => b.toString(16).padStart(2, '0'))
                .join('');

            userAvatar.value = `https://api.dicebear.com/7.x/personas/svg/?seed=${userEmailSha256}?d=identicon`;

            trpcClient.account.currentUser.query().then((currentUser) => {
                user.value = currentUser;
                localStorage.setItem('user', JSON.stringify(currentUser));
            });
        } else {
            user.value = null;
            localStorage.removeItem('user');
        }
    });

    const isAuthenticated = computed(() => {
        return user.value !== null;
    });

    async function login(email: string, password: string) {
        await authClient.signIn.email({
            email,
            password,
        });

        user.value = await trpcClient.account.currentUser.query();

        const notificationStore = useNotificationStore();
        await notificationStore.loadNotifications();
    }

    async function loginWithPasskey() {
        await authClient.signIn.passkey();
        user.value = await trpcClient.account.currentUser.query();

        const notificationStore = useNotificationStore();
        await notificationStore.loadNotifications();
    }

    async function logout() {
        user.value = null;
        localStorage.removeItem('user');
        authClient.signOut();
        router.push('/account/login');
    }

    async function refreshUser() {
        try {
            const userData = await trpcClient.account.currentUser.query();
            user.value = userData;
        } catch (error) {
            logout();
        }
    }

    async function register(user: {
        email: string;
        password: string;
        displayName: string;
    }) {
        await authClient.signUp.email({
            email: user.email,
            password: user.password,
            name: user.displayName,
        });
    }

    return {
        user,
        userAvatar,
        organizations,
        returnUrl,
        isAuthenticated,
        login,
        loginWithPasskey,
        logout,
        refreshUser,
        register,
    };
});
