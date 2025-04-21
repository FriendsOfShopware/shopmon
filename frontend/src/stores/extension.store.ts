import { defineStore } from 'pinia';
import { ref } from 'vue';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

export const useExtensionStore = defineStore('extension', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const extensions = ref<RouterOutput['account']['currentUserApps']>([]);

    async function loadExtensions() {
        isLoading.value = true;
        extensions.value = await trpcClient.account.currentUserApps.query();
        isLoading.value = false;
    }

    return {
        isLoading,
        isRefreshing,
        extensions,
        loadExtensions
    };
});
