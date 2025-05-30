import { authClient } from '@/helpers/auth-client';
import { defineStore } from 'pinia';
import { ref, watch } from 'vue';
const session = authClient.useSession();

export const useAuthStore = defineStore('auth', () => {
    const returnUrl = ref<string | null>(null);

    return {
        returnUrl,
    };
});
