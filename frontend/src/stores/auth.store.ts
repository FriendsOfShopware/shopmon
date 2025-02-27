import { defineStore } from 'pinia';
import { ref, computed } from 'vue';
import { router } from '@/router';
import { useNotificationStore } from './notification.store';
import { trpcClient, RouterOutput } from '@/helpers/trpc';
import { client } from '@passwordless-id/webauthn';

export const useAuthStore = defineStore('auth', () => {
    const user = ref<RouterOutput['account']['currentUser'] | null>(JSON.parse(localStorage.getItem('user') as string));
    const returnUrl = ref<string | null>(null);
    const access_token = ref<string | null>(localStorage.getItem('access_token'));
    const passkeys = ref<RouterOutput['auth']['passkey']['listDevices'] | null>(null);

    const isAuthenticated = computed(() => {
        return user.value !== null && access_token.value !== null;
    });

    async function login(email: string, password: string) {
        const token = await trpcClient.auth.loginWithPassword.mutate({
            email,
            password,
        });

        localStorage.setItem('access_token', token);
        access_token.value = token;

        const userData = await trpcClient.account.currentUser.query();
        user.value = userData;
        localStorage.setItem('user', JSON.stringify(userData));

        const notificationStore = useNotificationStore();
        notificationStore.connect(token);
        await notificationStore.loadNotifications();
    }

    async function loginWithPasskey() {
        const challenge = await trpcClient.auth.passkey.challenge.mutate();

        const authentication = await client.authenticate({
            challenge,
            userVerification: 'required',
            timeout: 60000,
        });

        const token = await trpcClient.auth.passkey.authenticateDevice.mutate(authentication);

        localStorage.setItem('access_token', token);
        access_token.value = token;

        const userData = await trpcClient.account.currentUser.query();
        user.value = userData;
        localStorage.setItem('user', JSON.stringify(userData));

        const notificationStore = useNotificationStore();
        notificationStore.connect(token);
        await notificationStore.loadNotifications();
    }

    async function logout() {
        user.value = null;
        access_token.value = null;
        localStorage.removeItem('user');
        localStorage.removeItem('access_token');
        router.push('/account/login');
    }

    async function refreshUser() {
        try {
            const userData = await trpcClient.account.currentUser.query();
            user.value = userData;
            localStorage.setItem('user', JSON.stringify(userData));
        } catch (error) {
            logout();
        }
    }

    async function loadPasskeys() {
        passkeys.value = await trpcClient.auth.passkey.listDevices.query();
    }

    async function registerPasskey(name: string) {
        const challenge = await trpcClient.auth.passkey.registrationChallenge.mutate();

        const registration = await client.register(name, challenge, {
            authenticatorType: 'auto',
            userVerification: 'required',
            timeout: 60000,
            debug: false,
        });

        await trpcClient.auth.passkey.registerDevice.mutate(registration);
        await loadPasskeys();
    }

    async function deletePasskey(credentialId: string) {
        await trpcClient.auth.passkey.deleteDevice.mutate({ credentialId });
        await loadPasskeys();
    }

    return {
        user,
        returnUrl,
        access_token,
        passkeys,
        isAuthenticated,
        login,
        loginWithPasskey,
        logout,
        refreshUser,
        loadPasskeys,
        registerPasskey,
        deletePasskey
    };
});
