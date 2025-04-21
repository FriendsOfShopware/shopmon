import { defineStore } from 'pinia';
import { ref } from 'vue';
import { trpcClient, RouterOutput } from '@/helpers/trpc';

export const useDashboardStore = defineStore('dashboard', () => {
    const isLoading = ref(false);
    const isRefreshing = ref(false);
    const changelogs = ref<RouterOutput['account']['currentUserChangelogs']>([]);

    async function loadChangelogs() {
        isLoading.value = true;
        changelogs.value = await trpcClient.account.currentUserChangelogs.query();
        isLoading.value = false;
    }

    return {
        isLoading,
        isRefreshing,
        changelogs,
        loadChangelogs
    };
});
